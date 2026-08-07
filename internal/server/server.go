package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/handler"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config, store handler.MetricsStorage) *Server {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(store).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValueHandler(store).ServeHTTP)
	r.Get("/", handler.NewIndexHandler(store).ServeHTTP)

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
