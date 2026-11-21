// Package config читает настройки сервера из флагов и переменных окружения.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Config содержит настройки HTTP-сервера, хранения метрик, подписей и аудита.
type Config struct {
	// Addr задаёт адрес прослушивания HTTP: флаг -a, переменная ADDRESS.
	Addr string
	// StoreInterval задаёт период сохранения в секундах; 0 включает синхронную запись.
	// Настраивается через -i и STORE_INTERVAL.
	StoreInterval int
	// FileStoragePath задаёт путь к JSON-файлу метрик: -f, FILE_STORAGE_PATH.
	FileStoragePath string
	// Restore включает загрузку метрик из файла при запуске: -r, RESTORE.
	Restore bool
	// DatabaseDSN задаёт подключение PostgreSQL: -d, DATABASE_DSN; пустое значение выбирает файл.
	DatabaseDSN string
	// Key задаёт ключ HMAC: -k, KEY; пустое значение отключает HashMiddleware.
	Key string
	// AuditFile задаёт файл аудита: --audit-file, AUDIT_FILE; пустое значение отключает приёмник.
	AuditFile string
	// AuditURL задаёт HTTP-приёмник: --audit-url, AUDIT_URL; пустое значение отключает приёмник.
	AuditURL string
}

// Parse разбирает os.Args и затем применяет переменные окружения поверх флагов.
// Некорректные значения STORE_INTERVAL и RESTORE игнорируются.
func Parse() Config {
	var cfg Config
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&cfg.Key, "k", "", "key for hash signing")
	fs.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	fs.IntVar(&cfg.StoreInterval, "i", 300, "interval in seconds for saving metrics to file")
	fs.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "path to file for storing metrics")
	fs.BoolVar(&cfg.Restore, "r", true, "restore metrics from file on startup")
	fs.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")
	fs.StringVar(&cfg.AuditFile, "audit-file", "", "path to audit log file (empty disables file audit)")
	fs.StringVar(&cfg.AuditURL, "audit-url", "", "audit receiver URL (empty disables HTTP audit)")
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
	if envKey, ok := os.LookupEnv("KEY"); ok {
		cfg.Key = envKey
	}
	if envFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		cfg.AuditFile = envFile
	}
	if envURL, ok := os.LookupEnv("AUDIT_URL"); ok {
		cfg.AuditURL = envURL
	}

	return cfg
}
