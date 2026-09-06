package handler

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const pingTimeout = 2 * time.Second

var logger = zap.Must(zap.NewProduction()).Sugar()

type Pinger interface {
	Ping(ctx context.Context) error
}

type PingHandler struct {
	pinger Pinger
}

func NewPingHandler(pinger Pinger) *PingHandler {
	return &PingHandler{pinger: pinger}
}

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
