package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	WelcomeText = `
	Добро пожаловать в World of Gophers!

	Ты начинаешь путь ученика Go.
	Текущий уровень: 1
	XP: 0
	`
)

func HandleMessage(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID
		text := update.Message.Text

		if text == "/start" {
			bot.Send(tgbotapi.NewMessage(chatID, WelcomeText))
		}
	}
}