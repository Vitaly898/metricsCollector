package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

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

	srv := server.New(cfg, fileStorage, zapLogger)

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
