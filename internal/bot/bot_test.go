package bot

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/moritys/world_of_gophers/internal/models"
)

type fakeStore struct {
	player  models.Player
	getErr  error
	created bool
}

type fakeBot struct {
	sent []string
}

func (f *fakeBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if msg, ok := c.(tgbotapi.MessageConfig); ok {
		f.sent = append(f.sent, msg.Text)
	}
	return tgbotapi.Message{}, nil
}

func TestNewPlayer(t *testing.T) {

}

func TestExistPlayer(t *testing.T) {

}

func TestErrorDB(t *testing.T) {

}
