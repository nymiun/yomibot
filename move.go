package main

import (
	"strconv"

	"github.com/nemphi/sento"
)

func (a *agata) move(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	a1, err := strconv.Atoi(info.MessageContent)
	if err != nil {
		// TODO: make prettier
		bot.Send(info, "a1 is not number")
		return nil
	}
	gs := gsi.(*guildState)
	gs.Lock()
	song, ok := gs.queue.Get(a1)
	if !ok {
		// TODO: make prettier
		bot.Send(info, "a1 is not valid index")
		return nil
	}
	gs.queue.Insert(0, song)
	gs.queue.Remove(a1 + 1)
	gs.Unlock()

	return nil
}
