package main

import (
	"context"
	"strconv"

	"github.com/andersfylling/disgord"
)

func (a *yomi) move(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	if !a.lavaNode.HasPlayer(channel.GuildID.String()) {
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	index, err := strconv.Atoi(msg.Content)
	if err != nil {
		// TODO: make prettier
		a.client.SendMsg(channel.ID, "index is not number")
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(channel.GuildID.String())
	p.Lock()
	if index-1 <= p.Queue.Size() {
		song, _ := p.Queue.Get(index - 1)
		p.Queue.Insert(0, song)
		p.Queue.Remove(index)
	} else {
		a.client.SendMsg(channel.ID, "Could not move track, index out of range")
		msg.React(context.Background(), a.client, "🛑")
	}
	p.Unlock()
	msg.React(context.Background(), a.client, "✅")
	return nil
}
