package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) clear(bot *sento.Bot, info sento.HandleInfo) error {
	if !a.lavaNode.HasPlayer(info.GuildID) {
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(info.GuildID)
	p.Lock()
	p.Queue.Clear()
	p.Unlock()
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)
	gs.Lock()
	gs.queue.Clear()
	gs.Unlock()
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
