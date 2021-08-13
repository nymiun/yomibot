package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) clear(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)
	gs.Lock()
	gs.queue.Clear()
	gs.Unlock()

	return nil
}
