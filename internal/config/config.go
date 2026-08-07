package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr string
}

func Parse() Config {
	var cfg Config
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "адрес сервера")
	flag.Parse()

	// Переменная окружения имеет приоритет над флагом
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}

	return cfg
}
