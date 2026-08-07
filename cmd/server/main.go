 package main

  import (
      "log"

      "github.com/Vitaly898/metricsCollector/internal/config"
      "github.com/Vitaly898/metricsCollector/internal/server"
      "github.com/Vitaly898/metricsCollector/internal/storage"
  )

  func main() {
      cfg := config.Parse()
      memStorage := storage.NewMemStorage()
      srv := server.New(cfg, memStorage)

      if err := srv.Run(); err != nil {
          log.Fatalf("Сервер не запустить: %v", err)
      }
  }