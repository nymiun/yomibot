package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) stop(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Load(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)

	if gs.fetcherCmd == nil {
		return nil
	}
	gs.stopper <- struct{}{}
	if gs.paused {
		gs.resumer <- struct{}{}
	}
	gs.looping = false
	gs.fetcherCmd.Close()

	return nil
}
