package main

import (
	"github.com/nemphi/sento"
)

func (a *agata) stop(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)
	gs.Lock()

	if gs.fetcherOut == nil {
		return nil
	}
	gs.stopper <- struct{}{}
	if gs.paused {
		gs.resumer <- struct{}{}
	}
	gs.looping = false
	gs.stopping = true
	gs.queue.Clear()
	err := gs.fetcherCmd.Process.Kill()
	if err != nil {
		gs.Unlock()
		bot.LogError(err.Error())
		return err
	}
	err = gs.ffmpegCmd.Process.Kill()
	if err != nil {
		gs.Unlock()
		bot.LogError(err.Error())
		return err
	}

	gs.Unlock()

	return err
}