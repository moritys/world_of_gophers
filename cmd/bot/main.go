package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.ParseConfig()

	// db pool
	pool, err := database.GetDBPool(ctx, cfg.DBURL)
	if err != nil {
		return fmt.Errorf("подключение к БД:%w", err)
	}
	defer func() {
		log.Println("закрываю пул БД...")
		pool.Close()
		log.Println("пул закрыт")
	}()
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
	defer bot.StopReceivingUpdates()

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				log.Println("канал обновлений закрыт, завершаюсь")
				return nil
			}

			if update.Message == nil {
				continue
			}

			func() {
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				handleBot.HandleMessage(ctx, bot, update, storage)
			}()

		case <-ctx.Done():
			log.Println("получен сигнал, завершаюсь:", context.Cause(ctx))
			return nil
		}
	}
}
