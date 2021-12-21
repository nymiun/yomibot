package main

import (
	"context"

	"github.com/andersfylling/disgord"
)

func (a *yomi) stop(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	if !a.lavaNode.HasPlayer(channel.GuildID.String()) {
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(channel.GuildID.String())
	p.Stop()
	msg.React(context.Background(), a.client, "✅")
	return nil
}
