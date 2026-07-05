## Герой

```go
type User struct {
	ID        int64
	Name      string
	Level     int
	XP        int
	Gold      int
	Strength  int
	Knowledge int
	Focus     int
}
```

# Профиль игрока

Что-то вроде:

```
Маша
Backend Novice

Уровень: 8
XP: 720/1000

🔥 Серия: 6 дней

Навыки:
Go ............ 12
Backend ....... 9
Базы данных ... 4
Алгоритмы ..... 7

Побеждено боссов: 8/24

Золото: 320
```

> Заметка: в текущем коде (`internal/models`) структура называется `Player` и пока
> содержит только `ID`, `Name`, `Level` — без `XP`, `Gold` и характеристик.
> Расширение этой структуры до версии выше — это [этап 01 в roadmap](../../roadmap/01-hero-foundation.md).
