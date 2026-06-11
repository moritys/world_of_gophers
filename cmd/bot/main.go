package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	handleBot "github.com/moritys/world_of_gophers/internal/bot"
	"github.com/moritys/world_of_gophers/internal/config"
	"github.com/moritys/world_of_gophers/internal/database"
)

func main() {
	ctx := context.Background()
	cfg := config.ParseConfig()

	// db pool
	db, err := database.GetDBPool(cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connected!")

	if err := database.CreateTables(ctx, db); err != nil {
		log.Fatal(err)
	}

	// bot start
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Авторизован как %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleBot.HandleMessage(bot, update, db)
		}
	}
}
