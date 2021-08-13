package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) leave(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)

	if gs.fetcherOut == nil {
		return nil
	}
	gs.stopper <- struct{}{}
	err := gs.fetcherCmd.Process.Kill()
	if err != nil {
		return err
	}
	err = gs.voice.Disconnect()
	if err != nil {
		return err
	}
	close(gs.stopper)
	close(gs.pauser)
	close(gs.resumer)
	a.guildMap.Delete(info.GuildID)

	return nil
}
