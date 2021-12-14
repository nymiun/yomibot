package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/nemphi/lavago"
	"github.com/patrickmn/go-cache"
)

type guildState struct {
	looping       bool
	textChannelID string
	sync.RWMutex
}

type agata struct {
	spotifyClientID     string
	spotifyClientSecret string
	spotifyAccessToken  string
	dbDsn               string

	lavaHost      string
	lavaPort      string
	lavaPassword  string
	lavaResumeKey string
	lavaSSL       bool

	client *disgord.Client
	// bot      *sento.Bot
	guildMap *cache.Cache
	db       *pgdb
	lavaNode *lavago.Node
}

func (a *agata) Start(client *disgord.Client) (err error) {
	a.client = client
	a.guildMap = cache.New(time.Minute*10, time.Minute*11)

	lavaCfg := lavago.NewConfig()
	lavaCfg.Hostname = a.lavaHost
	lavaPort, err := strconv.Atoi(a.lavaPort)
	if err != nil {
		return err
	}
	lavaCfg.Port = lavaPort
	lavaCfg.Authorization = a.lavaPassword
	lavaCfg.ResumeKey = a.lavaResumeKey
	lavaCfg.SSL = a.lavaSSL
	lavaCfg.BufferSize = 1024
	lavaNode, err := lavago.NewNode(lavaCfg)
	if err != nil {
		log.Error("Creating node: " + err.Error())
		return err
	}

	user, err := client.CurrentUser().Get()
	if err != nil {
		return
	}

	lavaNode.TrackEnded = a.trackEnded
	lavaNode.TrackStarted = a.trackStarted
	lavaNode.ConnectVoice = func(guildID, channelID string, deaf bool) error {
		_, _, err := client.Guild(disgord.ParseSnowflakeString(guildID)).VoiceChannel(disgord.ParseSnowflakeString(channelID)).JoinManual(false, deaf)
		// lavaNode.OnVoiceStateUpdate(user.ID.String(), stateUpdate.UserID.String(), guildID, stateUpdate.SessionID)
		// lavaNode.OnVoiceServerUpdate(guildID, serverUpdate.Endpoint, serverUpdate.Token)
		return err
	}

	client.Gateway().VoiceStateUpdate(func(sess disgord.Session, evt *disgord.VoiceStateUpdate) {
		lavaNode.OnVoiceStateUpdate(user.ID.String(), evt.UserID.String(), evt.GuildID.String(), evt.SessionID)
	})

	client.Gateway().VoiceServerUpdate(func(sess disgord.Session, evt *disgord.VoiceServerUpdate) {
		lavaNode.OnVoiceServerUpdate(evt.GuildID.String(), evt.Endpoint, evt.Token)
	})

	b, err := client.Gateway().GetBot()
	if err != nil {
		return err
	}
	err = lavaNode.Connect(user.ID.String(), fmt.Sprint(b.Shards))
	if err != nil {
		log.Error("Connecting TO node: " + err.Error())
		return err
	}

	a.lavaNode = lavaNode

	a.db, err = newDBConnection(a.dbDsn)
	if err != nil {
		log.Error("Connecting DB: " + err.Error())
		return err
	}

	err = a.getSpotifyToken(false)

	return
}
func (a *agata) Stop() (err error) {
	return a.lavaNode.Close()
}

func (a *agata) Name() string {
	return "Agata"
}

func (a *agata) Triggers() map[string]func(*disgord.Message) error {
	return map[string]func(*disgord.Message) error{
		"p":      a.play,
		"play":   a.play,
		"s":      a.skip,
		"skip":   a.skip,
		"stop":   a.stop,
		"ss":     a.stop,
		"pause":  a.pause,
		"resume": a.resume,

		"seek": a.seek,

		"queue": a.queue,
		"q":     a.queue,
		"move":  a.move,
		"next":  a.move,
		"swap":  a.swap,
		"clear": a.clear,
		"leave": a.leave,
		"quit":  a.leave,
		// "history",
		// "nowplaying",
		"loop": a.loop,
		// "speed",
		"volume": a.volume,
	}
}

func (a *agata) Handle(msgChan chan *disgord.MessageCreate) {
	for msg := range msgChan {
		msgSplitContent := strings.Split(msg.Message.Content, " ")
		msgContent := strings.Join(msgSplitContent[1:], " ")
		trigger, ok := a.Triggers()[strings.ToLower(msgSplitContent[0])]
		if !ok {
			continue
		}
		msg.Message.Content = msgContent

		go func() {
			err := trigger(msg.Message)
			if err != nil {
				log.Error(err)
			}
		}()
	}
}
