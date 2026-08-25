package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	// Драйвер PostgreSQL импортируется «вслепую» (с подчёркиванием):
	// нам нужен только side-эффект импорта — драйвер регистрирует себя
	// в database/sql под именем "pgx". Напрямую пакет stdlib не используем.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/server"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = zapLogger.Sync() }()

	cfg := config.Parse()

	memStorage := storage.NewMemStorage()
	fileStorage := storage.NewFileStorage(memStorage, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
	fileStorage.Start()

	// Подключение к БД. sql.Open НЕ устанавливает соединение —
	// он лишь создаёт пул, который подключится лениво при первом запросе.
	// Реальную проверку связи делает хендлер /ping через PingContext.
	// Если DSN не задан — работаем без БД (nil), как в предыдущих итерациях.
	var db *sql.DB
	if cfg.DatabaseDSN != "" {
		var err error
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			zapLogger.Fatal("Failed to open database", zap.Error(err))
		}
		defer func() { _ = db.Close() }() // закрываем пул при завершении программы
	}

	srv := server.New(cfg, fileStorage, zapLogger, db)

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		zapLogger.Error("Failed to shutdown server", zap.Error(err))
	}
	fileStorage.Stop()
}
