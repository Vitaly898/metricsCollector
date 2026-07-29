package handler


import(
	"net/http"
	"strconv"
	models "github.com/Vitaly898/metricsCollector/internal/model"
	"github.com/go-chi/chi/v5"
)


type ValueHandler struct {
	storage MetricsStorage
}

func NewValueHandler(s MetricsStorage) *ValueHandler{
	return &ValueHandler{storage:s}
}

func (h *ValueHandler)ServeHTTP(w http.ResponseWriter, r *http.Request){
	metricType := chi.URLParam(r,"type")
	metricName := chi.URLParam(r,"name")

	var value string

	switch metricType {
	case models.Counter:
		v,ok := h.storage.GetCounter(metricName)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		value = strconv.FormatInt(v,10)
	
	case models.Gauge:
	v,ok := h.storage.GetGauge(metricName)
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		return
	}
	value = strconv.FormatFloat(v,'f',-1,64)
default:
	w.WriteHeader(http.StatusNotFound)
	return
}
w.Header().Set("Content-Type","text/plain")
w.WriteHeader(http.StatusOK)
w.Write([]byte(value))
}