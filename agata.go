package main

import (
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/nemphi/sento"
	"github.com/patrickmn/go-cache"
	"layeh.com/gopus"
)

type guildState struct {
	encoder    *gopus.Encoder
	fetcherOut io.ReadCloser
	fetcherCmd *exec.Cmd
	ffmpegCmd  *exec.Cmd
	leaving    bool
	stopping   bool
	playing    bool
	paused     bool
	looping    bool
	stopper    chan struct{}
	pauser     chan struct{}
	resumer    chan struct{}
	voice      *discordgo.VoiceConnection
	queue      *arraylist.List
	sync.RWMutex
}

type agata struct {
	youtubeKey string
	guildMap   *cache.Cache
}

func (a *agata) Start(_ *sento.Bot) (err error) {
	a.guildMap = cache.New(time.Minute*10, time.Minute*11)
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
		"s",
		"skip",
		"stop",
		"pause",
		"resume",
		// "next",
		// "seek",
		// "queue",
		// "q",
		"clear",
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
	case "s", "skip":
		return a.skip(bot, info)
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
	case "clear":
		return a.clear(bot, info)
	default:
		return
	}
}
