package main

import (
	"context"
	"strconv"

	"github.com/andersfylling/disgord"
)

func (a *agata) volume(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	if !a.lavaNode.HasPlayer(channel.GuildID.String()) {
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	vol, err := strconv.Atoi(msg.Content)
	if err != nil {
		// TODO: make prettier
		a.client.SendMsg(channel.ID, "index is not number")
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(channel.GuildID.String())
	p.UpdateVolume(vol)
	msg.React(context.Background(), a.client, "✅")
	return nil
}
