package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/server"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

func main() {
	cfg := config.Parse()

	memStorage := storage.NewMemStorage()
	fileStorage := storage.NewFileStorage(memStorage, cfg.FileStoragePath, cfg.StoreInterval)
	if cfg.Restore {
		if err := fileStorage.Load(); err != nil {
			log.Printf("Failed to load metrics: %v", err)
		}
	}
	fileStorage.Start()

	srv := server.New(cfg, fileStorage)

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Printf("Failed to shutdown server: %v", err)
	}
	fileStorage.Stop()
}
