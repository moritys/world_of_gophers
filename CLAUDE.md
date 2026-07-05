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

**DB access pattern:** raw SQL via `pgx/v5`, no ORM. All queries live in `internal/database/db.go`. The `player` table has `id` (Telegram user ID), `name`, and `level`.

## Build

```
go build -o bot ./cmd/bot
```

The Dockerfile builds with `CGO_ENABLED=0` for a static binary deployed on `debian:bookworm-slim`.
