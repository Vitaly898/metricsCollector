package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Vitaly898/metricsCollector/internal/storage"
)

type UpdateHandler struct {
	storage storage.MetricsStorage
}

func NewUpdateHandler(s storage.MetricsStorage) *UpdateHandler {
	return &UpdateHandler{storage: s}
}
func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if (len(parts)) != 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if parts[1] != "counter" && parts[1] != "gauge" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if parts[2] == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metric := parts[1]
	metricName := parts[2]
	metricValue := parts[3]
	switch metric {
	case "counter":
		met, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(metricName, met)
	case "gauge":
		met, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(metricName, met)
	}

	w.WriteHeader(http.StatusOK)

}
