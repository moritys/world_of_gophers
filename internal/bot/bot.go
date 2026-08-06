package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/moritys/world_of_gophers/internal/database"
	"github.com/moritys/world_of_gophers/internal/models"
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
	XP: %d
	Gold: %d
	`
	ErrorText = `
	Произошла ошибка, попробуйте позже 🩹
	`
)

type PlayerStore interface {
	CreatePlayer(ctx context.Context, id int64, name string) error
	GetPlayer(ctx context.Context, id int64) (models.Player, error)
}

type MessageSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

func reply(sender MessageSender, id int64, text string) {
	if _, err := sender.Send(tgbotapi.NewMessage(id, text)); err != nil {
		log.Printf("отправка сообщения пользователю %d: %v", id, err)
	}
}

func HandleMessage(
	sender MessageSender,
	update tgbotapi.Update,
	store PlayerStore,
) {
	ctx := context.Background()
	userID := update.Message.From.ID
	name := update.Message.From.UserName
	text := update.Message.Text

	if text == "/start" {
		existPlayer, err := store.GetPlayer(ctx, userID)

		if errors.Is(err, database.ErrPlayerNotFound) {
			err = store.CreatePlayer(ctx, userID, name)
			if err != nil {
				log.Printf("создание игрока %d: %v", userID, err)
				reply(sender, userID, ErrorText)
				return
			}
			reply(sender, userID, WelcomeText)
			return
		}

		if err != nil {
			log.Printf("получение игрока %d: %v", userID, err)
			reply(sender, userID, ErrorText)
			return
		}

		reply(sender, userID, fmt.Sprintf(ReturnText, existPlayer.Name, existPlayer.Level, existPlayer.XP, existPlayer.Gold))
		return
	}
}
