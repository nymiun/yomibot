package main

import (
	"os"

	"github.com/nemphi/sento"
)

func main() {
	bot, err := sento.New(
		sento.UseConfig(&sento.Config{
			Token:  os.Getenv("DISCORD_TOKEN"),
			Prefix: os.Getenv("BOT_PREFIX"),
		}),
		sento.UseHandlers(&agata{
			youtubeKey:          os.Getenv("YOUTUBE_KEY"),
			spotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
			spotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
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
