package main

import (
	"strconv"
	"strings"

	"github.com/nemphi/sento"
)

func (a *agata) swap(bot *sento.Bot, info sento.HandleInfo) error {
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	args := strings.Split(info.MessageContent, " ")
	if len(args) < 2 {
		// TODO: send err msg
		bot.Send(info, "Not enough args")
		return nil
	}
	a1, err := strconv.Atoi(args[0])
	if err != nil {
		// TODO: make prettier
		bot.Send(info, "a1 is not number")
		return nil
	}
	a2, err := strconv.Atoi(args[1])
	if err != nil {
		// TODO: make prettier
		bot.Send(info, "a2 is not number")
		return nil
	}
	gs := gsi.(*guildState)
	gs.Lock()
	gs.queue.Swap(a1, a2)
	gs.Unlock()

	return nil
}
