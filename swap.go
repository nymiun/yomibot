package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/andersfylling/disgord"
)

func (a *agata) swap(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	args := strings.Split(msg.Content, " ")
	if len(args) < 2 {
		// TODO: send err msg
		a.client.SendMsg(channel.ID, "Not enough args")
		return nil
	}
	a1, err := strconv.Atoi(args[0])
	if err != nil {
		// TODO: make prettier
		a.client.SendMsg(channel.ID, "a1 is not number")
		return nil
	}
	a2, err := strconv.Atoi(args[1])
	if err != nil {
		// TODO: make prettier
		a.client.SendMsg(channel.ID, "a2 is not number")
		return nil
	}
	p := a.lavaNode.GetPlayer(channel.GuildID.String())
	p.Lock()
	p.Queue.Swap(a1-1, a2-1)
	p.Unlock()
	msg.React(context.Background(), a.client, "✅")
	return nil
}
