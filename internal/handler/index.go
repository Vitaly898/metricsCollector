package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

type IndexHandler struct {
	storage MetricsStorage
}

func NewIndexHandler(s MetricsStorage) *IndexHandler {
	return &IndexHandler{storage: s}
}

type metric struct {
	Name  string
	Value string
}

var indexTmpl = template.Must(template.New("index").Parse(`<html><body><ul>
  {{range .}}<li>{{.Name}}: {{.Value}}</li>
  {{end}}</ul></body></html>`))

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics := make([]metric, 0)
	for name, val := range h.storage.GetAllGauge() {
		metrics = append(metrics, metric{Name: name, Value: fmt.Sprintf("%v", val)})
	}
	for name, val := range h.storage.GetAllCounter() {
		metrics = append(metrics, metric{Name: name, Value: fmt.Sprintf("%v", val)})
	}
	var buf bytes.Buffer
	if err := indexTmpl.Execute(&buf, metrics); err != nil {
		logger.Errorw("Failed to render index page", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
