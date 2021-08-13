package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) loop(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Load(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)

	if gs.fetcherCmd == nil {
		return nil
	}
	gs.looping = !gs.looping

	return nil
}
