package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type UpdatesHandler struct {
	storage MetricsStorage
	audit   AuditPublisher
}

func NewUpdatesHandler(s MetricsStorage, publisher AuditPublisher) *UpdatesHandler {
	return &UpdatesHandler{storage: s, audit: publisher}
}

func (h *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, m := range metrics {
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
		case models.Gauge:
			if m.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if err := h.storage.UpdateMetrics(metrics); err != nil {
		logger.Errorw("Failed to update metrics batch", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if h.audit != nil {
		names := make([]string, 0, len(metrics))
		for _, metric := range metrics {
			names = append(names, metric.ID)
		}
		publishAudit(h.audit, r, names)
	}
	w.WriteHeader(http.StatusOK)
}
