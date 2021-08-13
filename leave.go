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

	gs.Lock()
	if gs.fetcherOut == nil {
		gs.Unlock()
		return nil
	}
	gs.looping = false
	gs.stopping = true
	gs.queue.Clear()
	if gs.paused {
		gs.resumer <- struct{}{}
	}
	gs.stopper <- struct{}{}
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
	// err = gs.fetcherCmd.Wait()
	// if err != nil {
	// 	gs.Unlock()
	// 	bot.LogError(err.Error())
	// 	return err
	// }
	// err = gs.ffmpegCmd.Wait()
	// if err != nil {
	// 	gs.Unlock()
	// 	bot.LogError(err.Error())
	// 	return err
	// }
	gs.leaving = true
	close(gs.stopper)
	close(gs.pauser)
	close(gs.resumer)
	a.guildMap.Delete(info.GuildID)
	gs.Unlock()

	return nil
}
