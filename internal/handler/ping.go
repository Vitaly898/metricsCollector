package handler

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const pingTimeout = 2 * time.Second

var logger = zap.Must(zap.NewProduction()).Sugar()

// Pinger определяет проверку доступности хранилища метрик.
type Pinger interface {
	// Ping возвращает ошибку, если хранилище недоступно.
	Ping(ctx context.Context) error
}

// PingHandler обслуживает GET /ping для проверки доступности хранилища.
type PingHandler struct {
	pinger Pinger
}

// NewPingHandler создаёт обработчик проверки переданного pinger.
func NewPingHandler(pinger Pinger) *PingHandler {
	return &PingHandler{pinger: pinger}
}

// ServeHTTP проверяет хранилище с таймаутом две секунды.
// Возвращает HTTP 200 при успехе или HTTP 500 при ошибке проверки.
func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		logger.Errorw("Storage ping failed", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
