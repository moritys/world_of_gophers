package bot

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/moritys/world_of_gophers/internal/models"
)

// ═══════════════════════════════════════════════════════════
// ОБВЯЗКА: фейки и помощник. Используются всеми тестами.
// ═══════════════════════════════════════════════════════════

// fakeStore подменяет настоящее хранилище.
//
// Поля двух сортов:
//
//	player, getErr  — ЧТО ОТДАВАТЬ. Задаём до вызова, управляем поведением.
//	created, ...    — ЧТО ЗАПОМНИЛИ. Заполняется во время вызова, проверяем после.
type fakeStore struct {
	// что отдавать
	player models.Player
	getErr error

	// что запомнили
	created     bool
	createdID   int64
	createdName string
}

// Методы с УКАЗАТЕЛЬНЫМ получателем — иначе запись в поля не будет видна снаружи.
// Значит интерфейсу удовлетворяет *fakeStore, и передавать надо &fakeStore{}.
func (f *fakeStore) GetPlayer(_ context.Context, _ int64) (models.Player, error) {
	return f.player, f.getErr
}

func (f *fakeStore) CreatePlayer(_ context.Context, id int64, name string) error {
	f.created = true
	f.createdID = id
	f.createdName = name
	return nil
}

// fakeBot подменяет Телеграм: вместо отправки складывает тексты в слайс.
type fakeBot struct {
	sent []string
}

func (f *fakeBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	// Chattable — интерфейс с неэкспортируемыми методами, подделать его нельзя.
	// Зато можно достать из него конкретный тип и прочитать текст.
	if msg, ok := c.(tgbotapi.MessageConfig); ok {
		f.sent = append(f.sent, msg.Text)
	}
	return tgbotapi.Message{}, nil
}

// Проверки на этапе компиляции: фейки действительно подходят под интерфейсы.
// Если сломаешь сигнатуру — сборка упадёт сразу, а не в середине теста.
var (
	_ PlayerStore   = (*fakeStore)(nil)
	_ MessageSender = (*fakeBot)(nil)
)

// makeUpdate собирает Update руками.
//
// Возвращает ЗНАЧЕНИЕ, а не указатель — HandleMessage принимает tgbotapi.Update.
// Заполняем только три поля, которые функция реально читает; остальные
// пусть остаются нулевыми, тесту до них дела нет.
func makeUpdate(id int64, name, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{ // Message — указатель
			From: &tgbotapi.User{ // From тоже указатель
				ID:       id,
				UserName: name,
			},
			Text: text,
		},
	}
}

func assertContainsPlayer(t *testing.T, text string, p models.Player) {
	t.Helper()

	checks := []struct {
		what string
		want string
	}{
		{"имя", p.Name},
		{"уровень", strconv.Itoa(p.Level)},
		{"XP", strconv.Itoa(p.XP)},
		{"золото", strconv.Itoa(p.Gold)},
	}

	for _, c := range checks {
		if !strings.Contains(text, c.want) {
			t.Errorf("в сообщении нет %s (%q):\n%s", c.what, c.want, text)
		}
	}
}

// ═══════════════════════════════════════════════════════════
// ТЕСТ 1 — образец. Разбери построчно.
// ═══════════════════════════════════════════════════════════

// Сценарий: игрока в базе нет, значит его надо создать и поприветствовать.
func TestHandleMessage_NewPlayer(t *testing.T) {
	// ── 1. ПОДГОТОВИТЬ ──────────────────────────────────────
	// Настраиваем фейк так, чтобы он изобразил «игрок не найден».
	// Именно на эту ошибку смотрит HandleMessage через errors.Is.
	store := &fakeStore{getErr: pgx.ErrNoRows}
	sender := &fakeBot{}
	update := makeUpdate(42, "masha", "/start")

	// ── 2. ВЫПОЛНИТЬ ────────────────────────────────────────
	// Ровно один вызов тестируемой функции.
	HandleMessage(sender, update, store)

	// ── 3. ПРОВЕРИТЬ ────────────────────────────────────────
	// Проверяем ОБЕ стороны поведения: что сделали с хранилищем
	// и что отправили пользователю.

	// сторона хранилища
	if !store.created {
		// Fatalf, а не Errorf: если игрока не создали, проверять
		// остальное бессмысленно — сценарий уже провален.
		t.Fatal("CreatePlayer не вызвана, а игрока в базе не было")
	}
	if store.createdID != 42 {
		t.Errorf("создали игрока с id=%d, хотим 42", store.createdID)
	}
	if store.createdName != "masha" {
		t.Errorf("создали игрока с именем %q, хотим \"masha\"", store.createdName)
	}

	// сторона пользователя
	if len(sender.sent) != 1 {
		t.Fatalf("отправлено %d сообщений, хотим ровно 1: %v", len(sender.sent), sender.sent)
	}
	if sender.sent[0] != WelcomeText {
		t.Errorf("отправили:\n%q\nхотим WelcomeText:\n%q", sender.sent[0], WelcomeText)
	}
}

// ═══════════════════════════════════════════════════════════
// ТЕСТ 2 — твой
// ═══════════════════════════════════════════════════════════

// Сценарий: игрок уже есть в базе.
//
// Подготовить: store с заполненным player (имя, уровень, XP) и БЕЗ getErr.
// Проверить:
//   - store.created остался false — существующего игрока создавать нельзя
//   - отправлено ровно одно сообщение
//   - в тексте есть имя и уровень игрока
//
// Подсказка: сравнивать текст целиком неудобно — он собран через Sprintf
// из многострочной константы. Возьми strings.Contains и проверь вхождение
// имени и уровня по отдельности. Уровень придётся превратить в строку:
// strconv.Itoa или fmt.Sprintf("%d", ...).
func TestHandleMessage_ExistingPlayer(t *testing.T) {
	store := &fakeStore{}
	store.player = models.Player{
		ID:    67,
		Name:  "sasha",
		Level: 99,
		XP:    3480,
		Gold:  44,
	}
	sender := &fakeBot{}
	update := makeUpdate(67, "sasha", "/start")

	HandleMessage(sender, update, store)

	if store.created {
		t.Fatal("игрок был создан, хотя уже сущестует")
	}
	if len(sender.sent) != 1 {
		t.Fatalf("отправлено %d сообщений, хотим ровно 1: %v", len(sender.sent), sender.sent)
	}
	assertContainsPlayer(t, sender.sent[0], store.player)
}

// ═══════════════════════════════════════════════════════════
// ТЕСТ 3 — твой
// ═══════════════════════════════════════════════════════════

// Сценарий: база вернула произвольную ошибку (не ErrNoRows).
//
// Подготовить: store с getErr = errors.New("что угодно").
// Проверить:
//   - store.created остался false
//   - отправлен ErrorText
func TestHandleMessage_StorageError(t *testing.T) {
	store := &fakeStore{}
	store.getErr = errors.New("какая то ошибка случилась")
	sender := &fakeBot{}
	update := makeUpdate(67, "sasha", "/start")

	HandleMessage(sender, update, store)

	if store.created {
		t.Errorf("пользователь %d был создан, но не должен был", store.player.ID)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("отправлено %d сообщений, хотим ровно 1: %v", len(sender.sent), sender.sent)
	}
	if sender.sent[0] != ErrorText {
		t.Errorf("отправили сообщение: %s, хотели: %s", sender.sent[0], ErrorText)
	}
}
