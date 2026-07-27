# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Telegram bot game called "World of Gophers" — a Go-learning RPG where Telegram users register as players and progress through levels. Written in Go, backed by PostgreSQL.

## Running locally

Start the database and bot together:
```
docker compose up --build
```

To run only the database (for local bot development):
```
docker compose up db
```

Then run the bot directly:
```
go run ./cmd/bot
```

Connect to the local DB (exposed on port 5332):
```
psql -h localhost -p 5332 -U MASHA -d wog
```

## Commands

```
go build -o bot ./cmd/bot   # build
go vet ./...                # lint
go test ./...               # run tests (none exist yet — added in roadmap stages 07–09)
```

The Dockerfile builds with `CGO_ENABLED=0` for a static binary on `debian:bookworm-slim`.

## Environment variables

Required in `.env`:
- `BOT_TOKEN` — Telegram bot token
- `DATABASE_URL` — full Postgres connection string (e.g. `postgres://MASHA:1234@db:5432/wog?sslmode=disable`; use `localhost:5332` when connecting outside Docker)

## Architecture

```
cmd/bot/main.go          — entry point: loads config, connects DB, starts Telegram polling loop
internal/config/         — reads BOT_TOKEN and DATABASE_URL from .env / env vars
internal/database/db.go  — pgxpool connection, schema creation at startup, all SQL queries
internal/bot/bot.go      — Telegram update handler; dispatches on message text
internal/models/         — plain Go structs (Player)
```

**Data flow:** `main` creates a `pgxpool.Pool` → calls `CreateTables` (idempotent, runs every startup) → enters the Telegram update loop → each message goes to `bot.HandleMessage` which reads/writes via `database.*` functions.

**DB access pattern:** raw SQL via `pgx/v5`, no ORM. All queries live in `internal/database/db.go`.

## Current state gaps to be aware of

- **Only `/start` is handled.** All other messages are silently dropped in `HandleMessage`.
- **`Player` struct vs DB schema mismatch.** `models.Player` has `XP`, `Gold`, `Strength`, `Knowledge`, `Focus` fields, but `CreateTables` only creates columns for `id`, `name`, `level`. `ReturnText` in `bot.go` shows `XP: 0, Gold: 0` as hardcoded placeholders, not DB values. When adding new game mechanics, both the DB schema and the struct's `Scan` call need updating together.
- **No migrations tooling yet.** Schema is created via raw `CREATE TABLE IF NOT EXISTS` at startup. Future schema changes (new columns, tables) must be added either to `CreateTables` or via a proper migration system (planned in stage 08).

## Language convention

User-facing strings and log messages are in Russian. Internal identifiers (variables, functions, packages) are in English.

## Game design & roadmap

- `docs/game-design/` — the game design (hero, quests, achievements, bosses, skills, streaks, currency).
- `roadmap/` — development roadmap split into stages. Start at `roadmap/00-overview.md` for current status. Stages 01–06 stay within the existing Telegram+DB setup; stages 07–09 add HTTP, tests, and migrations.
