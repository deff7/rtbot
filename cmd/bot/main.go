package main

import (
	"log"
	"os"

	"github.com/deff7/rutracker/internal/bot"
)

func main() {
	token := os.Getenv("TG_BOT_TOKEN")
	bot, err := bot.NewBot(token)
	if err != nil {
		log.Fatal(bot)
	}
	bot.Run()
}
