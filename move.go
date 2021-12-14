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
	p := a.lavaNode.GetPlayer(info.GuildID)
	p.Lock()
	if index-1 <= p.Queue.Size() {
		song, _ := p.Queue.Get(index - 1)
		p.Queue.Insert(0, song)
		p.Queue.Remove(index)
	} else {
		bot.Send(info, "Could not move track, index out of range")
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
	}
	p.Unlock()
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
