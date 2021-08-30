package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nemphi/lavago"
	"github.com/nemphi/sento"
)

func (a *agata) queue(bot *sento.Bot, info sento.HandleInfo) error {
	if !a.lavaNode.HasPlayer(info.GuildID) {
		bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "🛑")
		return nil
	}
	p := a.lavaNode.GetPlayer(info.GuildID)
	gsi, exist := a.guildMap.Get(info.GuildID)
	if !exist {
		return nil
	}
	gs := gsi.(*guildState)
	p.Lock()
	description := ""
	for i := 0; i < p.Queue.Size(); i++ {
		tr, _ := p.Queue.Get(i)
		description += fmt.Sprintf("%v %s\n", i+1, tr.(*lavago.Track).Info.Title)
	}
	gs.Lock()
	for i := 0; i < gs.queue.Size(); i++ {
		rt, _ := gs.queue.Get(i)
		description += fmt.Sprintf("%v %s\n", i+p.Queue.Size()+1, rt.(rawTrack).title)
	}
	gs.Unlock()
	bot.Sess().ChannelMessageSendEmbed(info.ChannelID, &discordgo.MessageEmbed{
		Title:       "Queue",
		Description: description,
	})
	p.Unlock()
	bot.Sess().MessageReactionAdd(info.ChannelID, info.MessageID, "✅")
	return nil
}
