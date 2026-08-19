package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

	// DB
	pool, err := setupDB(ctx, &cfg)
	if err != nil {
		return err
	}
	defer closeDB(pool)

	if err := database.CreateTables(ctx, pool); err != nil {
		return fmt.Errorf("создание таблицы БД:%w", err)
	}

	// bot
	bot, err := setupBot(&cfg)
	if err != nil {
		return err
	}

	return serve(ctx, bot, pool)
}

func waitTimeout(wg *sync.WaitGroup, d time.Duration) bool {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(d):
		return false
	}
}

func setupDB(ctx context.Context, cfg *config.Config) (pool *pgxpool.Pool, err error) {
	pool, err = database.GetDBPool(ctx, cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("подключение к БД:%w", err)
	}

	log.Println("Database connected!")
	return pool, nil
}

func closeDB(pool *pgxpool.Pool) {
	log.Println("закрываю пул БД...")
	pool.Close()
	log.Println("пул закрыт")
}

func setupBot(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("создание клиента бота:%w", err)
	}

	bot.Debug = true

	log.Printf("Авторизован как %s\n", bot.Self.UserName)
	return bot, nil
}

func serve(ctx context.Context, bot *tgbotapi.BotAPI, pool *pgxpool.Pool) error {
	sem := make(chan struct{}, 4) // максимум горутин одновременно
	var (
		inFlight atomic.Int64 // сколько прямо сейчас в работе
		finished atomic.Int64 // сколько успели закончить
	)
	storage := database.NewStorage(pool)

	var wg sync.WaitGroup
	defer func() {
		log.Printf("ожидаю завершения %d обработчиков...", inFlight.Load())
		if waitTimeout(&wg, 5*time.Second) {
			log.Printf("все завершились, обработано %d", finished.Load())
		} else {
			log.Printf("warn: таймаут, брошено %d", inFlight.Load())
		}
	}()

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

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				log.Println("получен сигнал во время ожидания места")
				return nil
			}

			inFlight.Add(1)
			wg.Go(func() {
				defer func() { <-sem }()
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				defer inFlight.Add(-1)
				defer finished.Add(1)
				handleBot.HandleMessage(ctx, bot, update, storage)
			})

		case <-ctx.Done():
			log.Println("получен сигнал, завершаюсь:", context.Cause(ctx))
			return nil
		}
	}
}
