package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/nemphi/lavago"
	"github.com/nemphi/sento"
	"github.com/patrickmn/go-cache"
)

type guildState struct {
	looping       bool
	queue         *arraylist.List
	textChannelID string
	sync.RWMutex
}

type agata struct {
	spotifyClientID     string
	spotifyClientSecret string
	spotifyAccessToken  string
	dbDsn               string
	bot                 *sento.Bot
	guildMap            *cache.Cache
	db                  *DB
	lavaNode            *lavago.Node
}

func (a *agata) Start(bot *sento.Bot) (err error) {
	a.bot = bot
	a.guildMap = cache.New(time.Minute*10, time.Minute*11)

	lavaCfg := lavago.NewConfig()
	lavaCfg.Hostname = "34.228.116.74"
	lavaCfg.Authorization = "yes"
	lavaCfg.BufferSize = 1024
	lavaNode, err := lavago.NewNode(lavaCfg)
	if err != nil {
		bot.LogError("Creating node: " + err.Error())
		return err
	}

	lavaNode.TrackEnded = a.trackEnded
	lavaNode.TrackStarted = a.trackStarted
	lavaNode.ConnectVoice = func(guildID, channelID string, deaf bool) error {
		return bot.Sess().ChannelVoiceJoinManual(guildID, channelID, false, deaf)
	}

	bot.Sess().AddHandler(func(sess *discordgo.Session, evt *discordgo.VoiceStateUpdate) {
		lavaNode.OnVoiceStateUpdate(bot.Sess().State.User.ID, sess.State.User.ID, evt.GuildID, evt.SessionID)
	})

	bot.Sess().AddHandler(func(sess *discordgo.Session, evt *discordgo.VoiceServerUpdate) {
		lavaNode.OnVoiceServerUpdate(evt.GuildID, evt.Endpoint, evt.Token)
	})

	err = lavaNode.Connect(bot.Sess().State.User.ID, fmt.Sprint(bot.Sess().ShardCount))
	if err != nil {
		bot.LogError("Connecting TO node: " + err.Error())
		return err
	}

	a.lavaNode = lavaNode

	a.db, err = NewDBConnection(a.dbDsn)
	if err != nil {
		bot.LogError("Connecting DB: " + err.Error())
		return err
	}

	err = a.getSpotifyToken(bot, false)

	return
}
func (a *agata) Stop(_ *sento.Bot) (err error) {
	return a.lavaNode.Close()
}

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
		"queue",
		// "q",
		"move",
		"swap",
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
	case "move":
		return a.move(bot, info)
	case "swap":
		return a.swap(bot, info)
	case "clear":
		return a.clear(bot, info)
	case "queue":
		return a.queue(bot, info)
	default:
		return
	}
}
