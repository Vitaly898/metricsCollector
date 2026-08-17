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
}

func Parse() Config {
	var cfg Config
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	fs.IntVar(&cfg.StoreInterval, "i", 300, "interval in seconds for saving metrics to file")
	fs.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "path to file for storing metrics")
	fs.BoolVar(&cfg.Restore, "r", true, "restore metrics from file on startup")
	_ = fs.Parse(os.Args[1:])

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		if v, err := strconv.Atoi(envInterval); err == nil {
			cfg.StoreInterval = v
		}
	}
	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		cfg.FileStoragePath = envPath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if v, err := strconv.ParseBool(envRestore); err == nil {
			cfg.Restore = v
		}
	}

	return cfg
}
