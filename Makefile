main_package_path = ./cmd/bot
binary_name = world_of_gophers

.PHONY: help
help: # список всех команд
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

.PHONY: fmt
fmt: # форматирует код
	go fmt ./...

.PHONY: vet
vet: # ищет ошибки кода
	go vet ./...

.PHONY: lint
lint: # запуск линтера
	golangci-lint run

.PHONY: test
test: # запуск тестов
	go test ./...

.PHONY: test-race
test-race: # запуск тестов с -race (поиск гонки)
	go test -race ./...

.PHONY: tidy
tidy: # форматирует файл go mod
	go mod tidy

.PHONY: check
check: fmt vet lint test-race # выполняет сразу все проверки кода

.PHONY: build
build: # собирает бинарник
	go build -o bin/$(binary_name) $(main_package_path)

.PHONY: run
run: # запуск на хосте, нужен make db-up
	DATABASE_URL=postgres://MASHA:1234@localhost:5332/wog go run $(main_package_path)

.PHONY: up
up: # поднять все контейнеры для приложения
	docker compose up -d --build

.PHONY: down
down: # остановить все контейнеры
	docker compose down

.PHONY: logs
logs: # посмотреть логи бота
	docker compose logs -f bot

.PHONY: ps
ps: # что запущено
	docker compose ps

.PHONY: db-up
db-up: # поднять только БД
	docker compose up -d db
