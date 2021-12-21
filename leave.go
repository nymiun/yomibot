package main

import (
	"context"

	"github.com/andersfylling/disgord"
	"github.com/nemphi/lavago"
)

func (a *yomi) leave(msg *disgord.Message) error {
	channel, err := a.client.Channel(msg.ChannelID).Get()
	if err != nil {
		return err
	}
	if !a.lavaNode.HasPlayer(channel.GuildID.String()) {
		a.client.SendMsg(channel.ID)
		msg.React(context.Background(), a.client, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(channel.GuildID.String())
	if p.State == lavago.PlayerStatePlaying {
		p.Stop()
	}
	err = a.lavaNode.Leave(channel.GuildID.String())
	if err != nil {
		msg.React(context.Background(), a.client, "🛑")
		return err
	}
	_, _, err = a.client.Guild(channel.GuildID).VoiceChannel(disgord.ParseSnowflakeString("")).JoinManual(false, false)
	if err != nil {
		msg.React(context.Background(), a.client, "🛑")
		return err
	}
	msg.React(context.Background(), a.client, "✅")
	return nil
}
