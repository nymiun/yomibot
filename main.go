package main

import (
	"context"
	"flag"
	"os"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
	"github.com/sirupsen/logrus"
)

var (
	discordBotToken     = flag.String("t", "", "Discord Bot Token")
	discordBotPrefix    = flag.String("p", "$", "Discord Bot Default Prefix")
	spotifyClientID     = flag.String("sci", "", "Spotify Client ID")
	spotifyClientSecret = flag.String("scs", "", "Spotify Client Secret")
	dbDsn               = flag.String("dsn", "", "DB Postgrest DSN")
	lavalinkHostname    = flag.String("lh", "127.0.0.1", "Lavalink Hostname")
	lavalinkPort        = flag.String("lp", "2333", "Lavalink Port")
	lavalinkPassword    = flag.String("lpw", "youshallnotpass", "Lavalink Password")
	lavalinkResumeKey   = flag.String("lrk", "Yomi", "Lavalink ResumeKey")
	lavalinkSSL         = flag.Bool("lssl", false, "Lavalink Enable SSL")
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.InfoLevel,
}

func main() {
	flag.Parse()
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		token = *discordBotToken
	}
	prefix := os.Getenv("DISCORD_BOT_PREFIX")
	if prefix == "" {
		prefix = *discordBotPrefix
	}
	sClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if sClientID == "" {
		sClientID = *spotifyClientID
	}
	sClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if sClientSecret == "" {
		sClientSecret = *spotifyClientSecret
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = *dbDsn
	}
	lh := os.Getenv("LAVA_HOSTNAME")
	if lh == "" {
		lh = *lavalinkHostname
	}
	lp := os.Getenv("LAVA_PORT")
	if lp == "" {
		lp = *lavalinkPort
	}
	lpw := os.Getenv("LAVA_PASSWORD")
	if lpw == "" {
		lpw = *lavalinkPassword
	}
	lrk := os.Getenv("LAVA_RESUME_KEY")
	if lrk == "" {
		lrk = *lavalinkResumeKey
	}
	client := disgord.New(disgord.Config{
		ProjectName: "Yomi",
		BotToken:    token,
		Logger:      log,
		Intents:     disgord.IntentGuilds | disgord.IntentGuildMessages | disgord.IntentGuildVoiceStates,
	})
	defer client.Gateway().StayConnectedUntilInterrupted()

	msgChan := make(chan *disgord.MessageCreate)

	filter, err := std.NewMsgFilter(context.Background(), client)
	if err != nil {
		panic(err)
	}
	filter.SetPrefix(prefix)
	client.Gateway().WithMiddleware(
		filter.NotByBot,
		filter.HasPrefix,
		filter.StripPrefix,
	).MessageCreateChan(msgChan)

	a := &yomi{
		spotifyClientID:     sClientID,
		spotifyClientSecret: sClientSecret,
		dbDsn:               dsn,
		lavaHost:            lh,
		lavaPort:            lp,
		lavaPassword:        lpw,
		lavaResumeKey:       lrk,
		lavaSSL:             *lavalinkSSL,
	}
	err = a.Start(client)
	if err != nil {
		panic(err)
	}

	go a.Handle(msgChan)

	client.Gateway().BotReady(func() {
		log.Println("Bot Ready")
	})

}
