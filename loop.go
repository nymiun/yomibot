package main

import (
	"context"

	"github.com/andersfylling/disgord"
)

func (a *agata) loop(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	gsi, exist := a.guildMap.Get(channel.GuildID.String())
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)
	gs.Lock()
	gs.looping = !gs.looping
	gs.Unlock()
	msg.React(context.Background(), a.client, "✅")
	return nil
}
