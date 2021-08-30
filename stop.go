package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) stop(bot *sento.Bot, info sento.HandleInfo) error {
	if !a.lavaNode.HasPlayer(info.GuildID) {
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(info.GuildID)
	p.Stop()
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
