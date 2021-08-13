package main

import (
	"io"
	"os/exec"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/nemphi/sento"
	"layeh.com/gopus"
)

type guildState struct {
	encoder    *gopus.Encoder
	fetcherCmd io.ReadCloser
	ffmpegCmd  *exec.Cmd
	voice      *discordgo.VoiceConnection
}

type agata struct {
	youtubeKey string
	guildMap   *sync.Map
}

func (a *agata) Start(_ *sento.Bot) (err error) {
	a.guildMap = &sync.Map{}
	return
}
func (a *agata) Stop(_ *sento.Bot) (err error) { return }

func (a *agata) Name() string {
	return "Agata"
}

func (a *agata) Triggers() []string {
	return []string{
		"p",
		"play",
		// "s",
		// "stop",
		// "pause",
		// "resume",
		// "skip",
		// "next",
		// "seek",
		// "queue",
		// "q",
		"leave",
		// "history",
		// "nowplaying",
		// "loop",
		// "speed",
		// "volume",
	}
}

func (a *agata) Handle(bot *sento.Bot, info sento.HandleInfo) (err error) {
	switch info.Trigger {
	case "p", "play":
		return a.play(bot, info)
	case "leave":
		return a.leave(bot, info)
	default:
		return
	}
}
