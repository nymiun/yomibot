package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) pause(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Load(info.GuildID)
	if !exist {
		return nil
	}

	gs := gsi.(*guildState)

	if gs.paused {
		return nil
	}

	gs.pauser <- struct{}{}

	return nil
}

func (a *agata) resume(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Load(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)

	if !gs.paused {
		return nil
	}

	gs.paused = false
	gs.resumer <- struct{}{}

	return nil
}
