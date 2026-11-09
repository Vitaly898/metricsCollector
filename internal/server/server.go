// Package server собирает маршруты и middleware HTTP-сервера метрик.
package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/logger"
	"github.com/Vitaly898/metricsCollector/internal/middleware"
)

// Server управляет HTTP-сервером приёма и чтения метрик.
type Server struct {
	httpServer *http.Server
}

// New создаёт сервер с маршрутами метрик, gzip, логированием и необязательной подписью.
// store и zapLogger должны быть заданы; nil publisher отключает публикацию аудита.
// Прослушивание адреса cfg.Addr начинается после вызова Run.
func New(cfg config.Config, store handler.MetricsStorage, zapLogger *zap.Logger, publisher handler.AuditPublisher) *Server {
	r := chi.NewRouter()
	r.Use(logger.Logger(zapLogger.Sugar()))
	r.Use(middleware.GzipMiddleware)
	if cfg.Key != "" {
		r.Use(middleware.HashMiddleware(cfg.Key))
	}

	r.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(store, publisher).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValueHandler(store).ServeHTTP)
	r.Get("/", handler.NewIndexHandler(store).ServeHTTP)

	r.Post("/update", handler.NewUpdateJSONHandler(store, publisher).ServeHTTP)
	r.Post("/update/", handler.NewUpdateJSONHandler(store, publisher).ServeHTTP)
	r.Post("/value", handler.NewValueJSONHandler(store).ServeHTTP)
	r.Post("/value/", handler.NewValueJSONHandler(store).ServeHTTP)

	r.Post("/updates/", handler.NewUpdatesHandler(store, publisher).ServeHTTP)

	r.Get("/ping", handler.NewPingHandler(store).ServeHTTP)

	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: r,
		},
	}
}

// Run начинает приём HTTP-запросов и блокирует выполнение до остановки сервера.
// При штатной остановке возвращает http.ErrServerClosed.
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

// Stop завершает работу HTTP-сервера, ожидая активные запросы в пределах ctx.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
