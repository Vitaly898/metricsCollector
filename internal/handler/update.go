package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type UpdateHandler struct {
	storage MetricsStorage
}

func NewUpdateHandler(s MetricsStorage) *UpdateHandler {
	return &UpdateHandler{storage: s}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case models.Counter:
		met, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateCounter(metricName, met); err != nil {
			logger.Errorw("Failed to update counter", "name", metricName, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case models.Gauge:
		met, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateGauge(metricName, met); err != nil {
			logger.Errorw("Failed to update gauge", "name", metricName, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
