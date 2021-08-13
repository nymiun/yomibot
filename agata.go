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
	paused     bool
	looping    bool
	stopper    chan struct{}
	pauser     chan struct{}
	resumer    chan struct{}
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
		"stop",
		"pause",
		"resume",
		// "skip",
		// "next",
		// "seek",
		// "queue",
		// "q",
		"leave",
		// "history",
		// "nowplaying",
		"loop",
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
	case "stop":
		return a.stop(bot, info)
	case "pause":
		return a.pause(bot, info)
	case "resume":
		return a.resume(bot, info)
	case "loop":
		return a.loop(bot, info)
	default:
		return
	}
}
