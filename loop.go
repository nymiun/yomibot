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

	gs.looping = !gs.looping

	return nil
}
