package main

import (
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/nemphi/lavago"
	"github.com/nemphi/sento"
	"github.com/patrickmn/go-cache"
)

type rawTrack struct {
	songID string
	title  string
	artist string
	url    string
}

func (a *agata) play(bot *sento.Bot, info sento.HandleInfo) error {

	vs, err := bot.Sess().State.VoiceState(info.GuildID, info.AuthorID)
	if err != nil {
		return err
	}

	p, err := a.lavaNode.Join(info.GuildID, vs.ChannelID)
	if err != nil {
		return err
	}

	var track *lavago.Track

	var sr *lavago.SearchResult

	var gs *guildState

	gsi, exists := a.guildMap.Get(info.GuildID)
	if !exists {
		gs = &guildState{
			queue: arraylist.New(),
		}
	} else {
		gs = gsi.(*guildState)
	}

	gs.textChannelID = info.ChannelID

	a.guildMap.Set(info.GuildID, gs, cache.DefaultExpiration)

	url, err := url.Parse(info.MessageContent)
	if err == nil &&
		(strings.Contains(url.Host, "youtube.com") ||
			strings.Contains(url.Host, "youtu.be") ||
			strings.Contains(url.Host, "twitch.tv")) {
		sr, err = a.lavaNode.Search(lavago.Direct, info.MessageContent)
	} else if err == nil && strings.Contains(url.Host, "spotify.com") {
		track, err = a.spotify(bot, info, gs, url, a.lavaNode, p)
		if err != nil {
			return err
		}
		goto playTrack
	} else if info.MessageContent == "file" {
		var msg *discordgo.Message
		msg, err = info.Message(bot)
		if err != nil {
			return err
		}
		if len(msg.Attachments) < 1 {
			bot.Send(info, "Missing attachment")
			return nil
		}
		sr, err = a.lavaNode.Search(lavago.Direct, msg.Attachments[0].URL)
	} else {
		sr, err = a.lavaNode.Search(lavago.YouTube, info.MessageContent)
	}
	if err != nil {
		return err
	}

	switch sr.Status {
	case lavago.NoMatchesSearchStatus:
		bot.Send(info, "No Matches")
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
		bot.Send(info, "Quitting default")
		return nil
	}

playTrack:
	if p.State == lavago.PlayerStatePlaying || p.State == lavago.PlayerStatePaused {
		p.Lock()
		p.Queue.Add(track)
		p.Unlock()
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
		return nil
	}
	if track == nil {
		return nil
	}
	err = p.PlayTrack(track)
	if err != nil {
		bot.LogError("ERR Playing Track: " + err.Error())
		return err
	}
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}

func (a *agata) trackStarted(evt lavago.TrackStartedEvent) {
	gsi, exists := a.guildMap.Get(evt.Player.GuildID)
	if exists {
		gs := gsi.(*guildState)
		gs.Lock()
		if !gs.queue.Empty() {
			rtI, ok := gs.queue.Get(0)
			if ok {
				rt := rtI.(rawTrack)
				track, _ := a.nodeSearchTrack(rt.songID, rt.title, rt.artist, rt.url)
				if track != nil {
					evt.Player.Lock()
					evt.Player.Queue.Add(track)
					evt.Player.Unlock()
				}
				gs.queue.Remove(0)
			}
		}
		a.bot.Sess().ChannelMessageSend(gs.textChannelID, "Playing "+evt.Player.Track.Info.Title)
		gs.Unlock()
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
			if exists {
				gs := gsi.(*guildState)
				gs.Lock()
				gs.queue.Clear()
				gs.Unlock()
			}
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
