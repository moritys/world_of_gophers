package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	handleBot "github.com/moritys/world_of_gophers/internal/bot"
	"github.com/moritys/world_of_gophers/internal/config"
)

func main() {
	cfg := config.ParseConfig()

	// bot start
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Авторизован как %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleBot.HandleMessage(bot, update)
		}
	}
}
