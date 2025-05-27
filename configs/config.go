package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN      string
	ServerPort string
	LogLevel   string
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load() // если нет .env – читаем из окружения
	cfg := Config{
		DBDSN:      os.Getenv("DB_DSN"),
		ServerPort: os.Getenv("SERVER_PORT"),
		LogLevel:   os.Getenv("LOG_LEVEL"),
	}
	if cfg.DBDSN == "" {
		return cfg, fmt.Errorf("DB_DSN is required")
	}
	if cfg.ServerPort == "" {
		return cfg, fmt.Errorf("SERVER_PORT is required")
	}
	return cfg, nil
}
