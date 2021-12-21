package main

import (
	"context"
	"fmt"

	"github.com/andersfylling/disgord"
	"github.com/nemphi/lavago"
)

func (a *yomi) queue(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	if !a.lavaNode.HasPlayer(channel.GuildID.String()) {
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(channel.GuildID.String())
	p.Lock()
	description := ""
	for i := 0; i < p.Queue.Size(); i++ {
		tr, _ := p.Queue.Get(i)
		description += fmt.Sprintf("%v %s\n", i+1, tr.(*lavago.Track).Info.Title)
	}
	_, err = a.client.SendMsg(channel.ID, &disgord.Embed{
		Title:       "Queue",
		Description: description,
	})
	p.Unlock()
	if err != nil {
		msg.React(context.Background(), a.client, "🛑")
		return err
	}
	msg.React(context.Background(), a.client, "✅")
	return nil
}
