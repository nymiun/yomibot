package main

import (
	"strconv"

	"github.com/nemphi/sento"
)

func (a *agata) move(bot *sento.Bot, info sento.HandleInfo) error {
	if !a.lavaNode.HasPlayer(info.GuildID) {
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return nil
	}
	index, err := strconv.Atoi(info.MessageContent)
	if err != nil {
		// TODO: make prettier
		bot.Send(info, "index is not number")
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return nil
	}
	index--
	p := a.lavaNode.GetPlayer(info.GuildID)
	p.Lock()
	if index-1 <= p.Queue.Size() {
		song, _ := p.Queue.Get(index - 1)
		p.Queue.Insert(0, song)
		p.Queue.Remove(index)
	} else {
		index -= p.Queue.Size()
		gsi, exist := a.guildMap.Get(info.GuildID)
		if !exist {
			return nil
		}
		gs := gsi.(*guildState)
		gs.Lock()
		song, ok := gs.queue.Get(index)
		if !ok {
			// TODO: make prettier
			bot.Send(info, "index is not valid index")
			bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
			gs.Unlock()
			p.Unlock()
			return nil
		}
		rt := song.(rawTrack)
		track, err := a.nodeSearchTrack(rt.songID, rt.title, rt.artist, rt.url)
		if err != nil {
			bot.Send(info, "Could not move track")
			bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
			gs.Unlock()
			p.Unlock()
			return nil
		}
		p.Queue.Insert(0, track)
		gs.queue.Remove(index)
		gs.Unlock()
	}
	p.Unlock()
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
