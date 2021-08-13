package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) leave(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Load(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)

	if gs.fetcherCmd == nil {
		return nil
	}
	gs.stopper <- struct{}{}
	gs.fetcherCmd.Close()
	err := gs.voice.Disconnect()
	if err != nil {
		return err
	}
	close(gs.stopper)
	a.guildMap.Delete(info.GuildID)

	return nil
}
