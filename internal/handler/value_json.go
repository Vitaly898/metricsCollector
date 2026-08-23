package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type ValueJSONHandler struct {
	storage MetricsStorage
}

func NewValueJSONHandler(s MetricsStorage) *ValueJSONHandler {
	return &ValueJSONHandler{storage: s}
}
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
