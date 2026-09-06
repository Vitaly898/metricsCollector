package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type UpdateJSONHandler struct {
	storage MetricsStorage
}

func NewUpdateJSONHandler(s MetricsStorage) *UpdateJSONHandler {
	return &UpdateJSONHandler{storage: s}
}

func (h *UpdateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if m.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch m.MType {
	case models.Counter:
		if m.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateCounter(m.ID, *m.Delta); err != nil {
			logger.Errorw("Failed to update counter", "id", m.ID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		v, _ := h.storage.GetCounter(m.ID)
		*m.Delta = v
	case models.Gauge:
		if m.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateGauge(m.ID, *m.Value); err != nil {
			logger.Errorw("Failed to update gauge", "id", m.ID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)

}
