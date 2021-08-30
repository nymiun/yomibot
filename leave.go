package main

import (
	"github.com/nemphi/lavago"
	"github.com/nemphi/sento"
)

func (a *agata) leave(bot *sento.Bot, info sento.HandleInfo) error {
	if !a.lavaNode.HasPlayer(info.GuildID) {
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(info.GuildID)
	if p.State == lavago.PlayerStatePlaying {
		p.Stop()
	}
	a.lavaNode.Leave(info.GuildID)
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
