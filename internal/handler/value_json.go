package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

// ValueJSONHandler обслуживает POST /value и /value/, возвращая метрику как JSON.
type ValueJSONHandler struct {
	storage MetricsStorage
}

// NewValueJSONHandler создаёт обработчик чтения JSON-метрики из s.
func NewValueJSONHandler(s MetricsStorage) *ValueJSONHandler {
	return &ValueJSONHandler{storage: s}
}

// ServeHTTP принимает JSON с id и type и возвращает HTTP 200 с текущим значением.
// Неверный JSON или тип приводит к HTTP 400, отсутствие значения — к HTTP 404.
func (h *ValueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch m.MType {
	case models.Counter:
		v, ok := h.storage.GetCounter(m.ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Delta = &v
	case models.Gauge:
		v, ok := h.storage.GetGauge(m.ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Value = &v
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)

}
