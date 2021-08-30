package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) loop(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)
	gs.Lock()
	gs.looping = !gs.looping
	gs.Unlock()
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
