package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token string
	DBURL string
}

func ParseConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env не найден, используем переменные окружения")
	}

	return Config{
		Token: os.Getenv("BOT_TOKEN"),
		DBURL: os.Getenv("DATABASE_URL"),
	}
}
