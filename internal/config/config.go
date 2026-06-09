package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token  string
	DBPath string
}

func ParseConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	return Config{
		Token:  os.Getenv("BOT_TOKEN"),
		DBPath: os.Getenv("DB_PATH"),
	}
}
