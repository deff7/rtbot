package main

import (
	"log"
	"os"

	"github.com/deff7/rutracker/internal/bot"
	"github.com/deff7/rutracker/internal/rutracker"
)

func main() {
	var (
		login    = os.Getenv("RT_LOGIN")
		password = os.Getenv("RT_PASSWORD")
	)
	rtclient := rutracker.NewClient()
	err := rtclient.Login(login, password)
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("TG_BOT_TOKEN")
	bot, err := bot.NewBot(token, rtclient)
	if err != nil {
		log.Fatal(err)
	}
	bot.Run()
}
