package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

// UpdateHandler обслуживает POST /update/{type}/{name}/{value}.
// Gauge заменяется новым значением, counter увеличивается на переданное приращение.
type UpdateHandler struct {
	storage MetricsStorage
	audit   AuditPublisher
}

// NewUpdateHandler создаёт обработчик записи в s; nil publisher отключает аудит.
func NewUpdateHandler(s MetricsStorage, publisher AuditPublisher) *UpdateHandler {
	return &UpdateHandler{storage: s, audit: publisher}
}

// ServeHTTP читает параметры маршрута chi и обновляет одну метрику.
// Возвращает HTTP 200 при успехе, 400 при неверном типе или значении,
// 404 при пустом имени и 500 при ошибке записи. Аудит выполняется после успешной записи.
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

	publishAudit(h.audit, r, []string{metricName})
	w.WriteHeader(http.StatusOK)
}
