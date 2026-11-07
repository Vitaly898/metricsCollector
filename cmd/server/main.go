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

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Vitaly898/metricsCollector/internal/audit"
	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/repository"
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

	var observers []audit.Observer
	if cfg.AuditURL != "" {
		observer, err := audit.NewHTTPObserver(cfg.AuditURL)
		if err != nil {
			zapLogger.Fatal("Failed to init HTTP audit", zap.Error(err))
		}
		observers = append(observers, observer)
	}
	if cfg.AuditFile != "" {
		observer, err := audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			zapLogger.Fatal("Failed to init file audit", zap.Error(err))
		}
		defer func() {
			if err := observer.Close(); err != nil {
				zapLogger.Error("Failed to close audit file", zap.Error(err))
			}
		}()
		observers = append(observers, observer)
	}
	var publisher handler.AuditPublisher
	if len(observers) > 0 {
		publisher = audit.NewPublisher(observers...)
	}

	var (
		db    *sql.DB
		store handler.MetricsStorage
	)

	if cfg.DatabaseDSN != "" {
		var err error
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			zapLogger.Fatal("Failed to open database", zap.Error(err))
		}
		defer func() { _ = db.Close() }()

		store, err = repository.NewPostgresStorage(db)
		if err != nil {
			zapLogger.Fatal("Failed to init database storage", zap.Error(err))
		}
	} else {
		memStorage := storage.NewMemStorage()
		fileStorage := storage.NewFileStorage(memStorage, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
		fileStorage.Start()
		defer func() { fileStorage.Stop() }()

		store = fileStorage
	}

	srv := server.New(cfg, store, zapLogger, publisher)

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
}
