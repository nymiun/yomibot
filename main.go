package main

import (
	"os"

	"github.com/nemphi/sento"
)

func main() {
	bot, err := sento.New(
		sento.UseConfig(&sento.Config{
			Token: os.Getenv("DISCORD_TOKEN"),
			// Perms: "2184211776",
		}),
		sento.UseHandlers(&agata{youtubeKey: os.Getenv("YOUTUBE_KEY")}),
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
