package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moritys/world_of_gophers/internal/database"
)

var (
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
)

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
		existPlayer, err := database.GetUserByID(ctx, pool, int(userID))

		if err != nil {
			bot.Send(tgbotapi.NewMessage(userID, WelcomeText))
			err := database.CreatePlayer(ctx, pool, int(userID), name)
			if err != nil {
				fmt.Println("Ошибка создания игрока:", err)
			}
			return
		}

		bot.Send(tgbotapi.NewMessage(userID, fmt.Sprintf(ReturnText, existPlayer.Name, existPlayer.Level)))
		return
	}
}
