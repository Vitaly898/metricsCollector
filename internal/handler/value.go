package handler

import (
	models "github.com/Vitaly898/metricsCollector/internal/model"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

// ValueHandler обслуживает GET /value/{type}/{name}, возвращая значение как текст.
type ValueHandler struct {
	storage MetricsStorage
}

// NewValueHandler создаёт обработчик чтения текстового значения из s.
func NewValueHandler(s MetricsStorage) *ValueHandler {
	return &ValueHandler{storage: s}
}

// ServeHTTP читает тип и имя из параметров маршрута chi.
// При успехе возвращает HTTP 200 и text/plain; неизвестный тип или отсутствующая
// метрика приводят к HTTP 404.
func (h *ValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	var value string

	switch metricType {
	case models.Counter:
		v, ok := h.storage.GetCounter(metricName)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		value = strconv.FormatInt(v, 10)

	case models.Gauge:
		v, ok := h.storage.GetGauge(metricName)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		value = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}
