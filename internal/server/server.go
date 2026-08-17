package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/logger"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config, store handler.MetricsStorage) *Server {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(logger.Logger(zapLogger.Sugar()))

	r.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(store).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValueHandler(store).ServeHTTP)
	r.Get("/", handler.NewIndexHandler(store).ServeHTTP)

	r.Post("/update", handler.NewUpdateJSONHandler(store).ServeHTTP)
	r.Post("/update/", handler.NewUpdateJSONHandler(store).ServeHTTP)
	r.Post("/value", handler.NewValueJSONHandler(store).ServeHTTP)
	r.Post("/value/", handler.NewValueJSONHandler(store).ServeHTTP)

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
