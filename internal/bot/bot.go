package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moritys/world_of_gophers/internal/database"
)

const (
	WelcomeText = `
	Добро пожаловать в World of Gophers!

	Ты начинаешь путь ученика Go.
	Текущий уровень: 1
	XP: 0
	`
	ReturnText = `
	С возвращением в World of Gophers, %s!

	Текущий уровень: %d
	XP: 0
	Gold: 0
	`
	ErrorText = `
	Произошла ошибка, попробуйте позже 🩹
	`
)

func reply(bot *tgbotapi.BotAPI, chatID int64, text string) {
	if _, err := bot.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
		log.Printf("отправка сообщения пользователю %d: %v", chatID, err)
	}
}

func HandleMessage(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
	pool *pgxpool.Pool,
) {
	ctx := context.Background()
	userID := update.Message.From.ID
	name := update.Message.From.UserName
	text := update.Message.Text

	if text == "/start" {
		existPlayer, err := database.GetUserByID(ctx, pool, userID)

		if errors.Is(err, pgx.ErrNoRows) {
			err = database.CreatePlayer(ctx, pool, userID, name)
			if err != nil {
				log.Printf("создание игрока %d: %v", userID, err)
				reply(bot, userID, ErrorText)
				return
			}
			reply(bot, userID, WelcomeText)
			return
		}

		if err != nil {
			log.Printf("получение игрока %d: %v", userID, err)
			reply(bot, userID, ErrorText)
			return
		}

		reply(bot, userID, fmt.Sprintf(ReturnText, existPlayer.Name, existPlayer.Level))
		return
	}
}
