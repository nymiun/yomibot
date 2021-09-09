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
	err := a.lavaNode.Leave(info.GuildID)
	if err != nil {
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return err
	}
	err = a.bot.Sess().ChannelVoiceJoinManual(info.GuildID, "", false, false)
	if err != nil {
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return err
	}
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
