package main

import (
	"flag"
	"os"

	"github.com/nemphi/sento"
)

var (
	discordBotToken     = flag.String("t", "", "Discord Bot Token")
	discordBotPrefix    = flag.String("p", "", "Discord Bot Default Prefix")
	spotifyClientID     = flag.String("sci", "", "Spotify Client ID")
	spotifyClientSecret = flag.String("scs", "", "Spotify Client Secret")
	dbDsn               = flag.String("dsn", "", "DB Postgrest DSN")
)

func main() {
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
	bot, err := sento.New(
		sento.UseConfig(&sento.Config{
			Token:  token,
			Prefix: prefix,
		}),
		sento.UseHandlers(&agata{
			spotifyClientID:     sClientID,
			spotifyClientSecret: sClientSecret,
			dbDsn:               dsn,
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
