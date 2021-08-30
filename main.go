package main

import (
	"flag"
	"os"

	"github.com/nemphi/sento"
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
	lavalinkResumeKey   = flag.String("lrk", "Agata", "Lavalink ResumeKey")
	lavalinkSSL         = flag.Bool("lssl", false, "Lavalink Enable SSL")
)

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
	bot, err := sento.New(
		sento.UseConfig(&sento.Config{
			Token:  token,
			Prefix: prefix,
		}),
		sento.UseHandlers(&agata{
			spotifyClientID:     sClientID,
			spotifyClientSecret: sClientSecret,
			dbDsn:               dsn,
			lavaHost:            lh,
			lavaPort:            lp,
			lavaPassword:        lpw,
			lavaResumeKey:       lrk,
			lavaSSL:             *lavalinkSSL,
		}),
	)

	if err != nil {
		panic(err)
	}

	err = bot.Start()
	if err != nil {
		panic(err)
	}
	bot.Stop()
}
