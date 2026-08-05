package main

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	handleBot "github.com/moritys/world_of_gophers/internal/bot"
	"github.com/moritys/world_of_gophers/internal/config"
	"github.com/moritys/world_of_gophers/internal/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var _ handleBot.PlayerStore = (*database.Storage)(nil)
	var _ handleBot.MessageSender = (*tgbotapi.BotAPI)(nil)
	ctx := context.Background()
	cfg := config.ParseConfig()

	// db pool
	pool, err := database.GetDBPool(cfg.DBURL)
	if err != nil {
		return fmt.Errorf("подключение к БД:%w", err)
	}
	defer pool.Close()
	log.Println("Database connected!")

	if err := database.CreateTables(ctx, pool); err != nil {
		return fmt.Errorf("создание таблицы БД:%w", err)
	}

	storage := database.NewStorage(pool)

	// bot start
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return fmt.Errorf("создание клиента бота:%w", err)
	}

	bot.Debug = true

	log.Printf("Авторизован как %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleBot.HandleMessage(bot, update, storage)
		}
	}

	return nil
}
