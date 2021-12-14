package main

import (
	"context"
	"net/url"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/nemphi/lavago"
	"github.com/patrickmn/go-cache"
)

func (a *agata) play(msg *disgord.Message) error {

	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		log.Println(err)
		return err
	}

	guild, err := a.client.Guild(channel.GuildID).Get()
	if err != nil {
		log.Println(err)
		return err
	}
	vs := &disgord.VoiceState{}
	for _, v := range guild.VoiceStates {
		if v.UserID == msg.Author.ID {
			vs = v
			break
		}
	}

	p, err := a.lavaNode.Join(guild.ID.String(), vs.ChannelID.String())
	if err != nil {
		return err
	}

	var track *lavago.Track

	var sr *lavago.SearchResult

	var gs *guildState

	gsi, exists := a.guildMap.Get(guild.String())
	if !exists {
		gs = &guildState{}
	} else {
		gs = gsi.(*guildState)
	}

	gs.textChannelID = channel.ID.String()

	a.guildMap.Set(guild.ID.String(), gs, cache.DefaultExpiration)

	url, err := url.Parse(msg.Content)
	if err == nil &&
		(strings.Contains(url.Host, "youtube.com") ||
			strings.Contains(url.Host, "youtu.be") ||
			strings.Contains(url.Host, "twitch.tv")) {
		sr, err = a.lavaNode.Search(lavago.Direct, msg.Content)
	} else if err == nil && strings.Contains(url.Host, "spotify.com") {
		track, err = a.spotify(a.client, gs, url, a.lavaNode, p)
		if err != nil {
			return err
		}
		goto playTrack
	} else if msg.Content == "file" {
		if len(msg.Attachments) < 1 {
			a.client.SendMsg(channel.ID, "No file attached")
			return nil
		}
		sr, err = a.lavaNode.Search(lavago.Direct, msg.Attachments[0].URL)
	} else {
		sr, err = a.lavaNode.Search(lavago.YouTube, msg.Content)
	}
	if err != nil {
		return err
	}

	switch sr.Status {
	case lavago.NoMatchesSearchStatus:
		a.client.SendMsg(channel.ID, "No Matches")
		return nil
	case lavago.SearchResultSearchStatus:
		track = sr.Tracks[0]
	case lavago.TrackLoadedSearchStatus:
		track = sr.Tracks[0]
	case lavago.PlaylistLoadedSearchStatus:
		if sr.Playlist.SelectedTrack != -1 {
			track = sr.Tracks[sr.Playlist.SelectedTrack]
		} else {
			track = sr.Tracks[0]
			p.Lock()
			for i := 0; i < len(sr.Tracks); i++ {
				if i != 0 {
					p.Queue.Add(sr.Tracks[i])
				}
			}
			p.Unlock()
		}
	default:
		a.client.SendMsg(channel.ID, "Quitting default")
		return nil
	}

playTrack:
	if p.State == lavago.PlayerStatePlaying || p.State == lavago.PlayerStatePaused {
		p.Lock()
		p.Queue.Add(track)
		p.Unlock()
		msg.React(context.Background(), a.client, "✅")
		return nil
	}
	if track == nil {
		return nil
	}
	err = p.PlayTrack(track)
	if err != nil {
		log.Println("ERR Playing Track: " + err.Error())
		return err
	}
	msg.React(context.Background(), a.client, "✅")
	return nil
}

func (a *agata) trackStarted(evt lavago.TrackStartedEvent) {
	gsi, exists := a.guildMap.Get(evt.Player.GuildID)
	if exists {
		gs := gsi.(*guildState)
		a.client.SendMsg(disgord.ParseSnowflakeString(gs.textChannelID), "Playing "+evt.Player.Track.Info.Title)
	}
}

func (a *agata) trackEnded(evt lavago.TrackEndedEvent) {
	go a.db.Select("GuildID", "Track").Create(&historyItem{GuildID: evt.Player.GuildID, Track: evt.Track.Track})
	if evt.Reason != lavago.FinishedReason &&
		// evt.Reason != lavago.ReplacedReason &&
		evt.Reason != lavago.StoppedReason {
		return
	}
	gsi, exists := a.guildMap.Get(evt.Player.GuildID)
	if exists {
		gs := gsi.(*guildState)
		gs.Lock()
		if gs.looping {
			if evt.Reason == lavago.FinishedReason {
				evt.Player.PlayTrack(evt.Player.Track)
				gs.Unlock()
				return
			} else {
				gs.looping = false
			}
		}
		gs.Unlock()
	}

	evt.Player.Lock()
	if !evt.Player.Queue.Empty() {
		if evt.Reason == lavago.StoppedReason {
			evt.Player.Queue.Clear()
		} else {
			trI, ok := evt.Player.Queue.Get(0)
			if ok {
				track := trI.(*lavago.Track)
				evt.Player.Queue.Remove(0)
				evt.Player.Unlock()
				evt.Player.PlayTrack(track)
				evt.Player.Lock()
			}
		}
	}
	evt.Player.Unlock()
}
