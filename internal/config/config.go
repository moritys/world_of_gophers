package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token  string
	DBURL string
}

func ParseConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	return Config{
		Token:  os.Getenv("BOT_TOKEN"),
		DBURL: os.Getenv("DATABASE_URL"),
	}
}
