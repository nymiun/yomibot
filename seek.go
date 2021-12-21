package main

import (
	"context"
	"time"

	"github.com/andersfylling/disgord"
)

func (a *yomi) seek(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	if !a.lavaNode.HasPlayer(channel.GuildID.String()) {
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	dur, err := time.ParseDuration(msg.Content)
	if err != nil {
		// TODO: make prettier
		a.client.SendMsg(channel.ID, "invalid input")
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(channel.GuildID.String())
	err = p.Seek(int(dur.Milliseconds()))
	p.Lock()
	if err != nil {
		msg.React(context.Background(), a.client, "🛑")
		return err
	}
	p.Unlock()
	msg.React(context.Background(), a.client, "✅")
	return nil
}
