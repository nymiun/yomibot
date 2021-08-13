package main

import (
	"github.com/nemphi/sento"
)

type agata struct {
	youtubeKey string
}

func (a *agata) Start(_ *sento.Bot) (err error) { return }
func (a *agata) Stop(_ *sento.Bot) (err error)  { return }

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
		// "leave",
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
	default:
		return
	}
}
