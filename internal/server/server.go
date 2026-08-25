package server

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/logger"
	"github.com/Vitaly898/metricsCollector/internal/middleware"
)

type Server struct {
	httpServer *http.Server
}

// db — пул соединений с PostgreSQL; может быть nil, если сервер запущен без БД.
func New(cfg config.Config, store handler.MetricsStorage, zapLogger *zap.Logger, db *sql.DB) *Server {
	r := chi.NewRouter()
	r.Use(logger.Logger(zapLogger.Sugar()))
	r.Use(middleware.GzipMiddleware)

	r.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(store).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValueHandler(store).ServeHTTP)
	r.Get("/", handler.NewIndexHandler(store).ServeHTTP)

	r.Post("/update", handler.NewUpdateJSONHandler(store).ServeHTTP)
	r.Post("/update/", handler.NewUpdateJSONHandler(store).ServeHTTP)
	r.Post("/value", handler.NewValueJSONHandler(store).ServeHTTP)
	r.Post("/value/", handler.NewValueJSONHandler(store).ServeHTTP)

	// Проверка соединения с БД.
	r.Get("/ping", handler.NewPingHandler(db).ServeHTTP)

	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: r,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
