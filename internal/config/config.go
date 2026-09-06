package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Addr            string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

func Parse() Config {
	var cfg Config
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	fs.IntVar(&cfg.StoreInterval, "i", 300, "interval in seconds for saving metrics to file")
	fs.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "path to file for storing metrics")
	fs.BoolVar(&cfg.Restore, "r", true, "restore metrics from file on startup")
	fs.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")
	_ = fs.Parse(os.Args[1:])

	if envAddr, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.Addr = envAddr
	}
	if envInterval, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		if v, err := strconv.Atoi(envInterval); err == nil {
			cfg.StoreInterval = v
		}
	}
	if envPath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = envPath
	}
	if envRestore, ok := os.LookupEnv("RESTORE"); ok {
		if v, err := strconv.ParseBool(envRestore); err == nil {
			cfg.Restore = v
		}
	}

	if envDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = envDSN
	}

	return cfg
}
