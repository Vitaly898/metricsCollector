package handler

import (
	"fmt"
	"net/http"
	"html/template"
)

type IndexHandler struct {
	storage MetricsStorage
}

func NewIndexHandler(s MetricsStorage) *IndexHandler{
	return  &IndexHandler{storage: s}
}
type metric struct{
	Name string
	Value string
}

var indexTmpl = template.Must(template.New("index").Parse(      `<html><body><ul>
  {{range .}}<li>{{.Name}}: {{.Value}}</li>
  {{end}}</ul></body></html>`))

  func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	metrics := make([]metric,0)
	for name,val := range h.storage.GetAllGauge(){
		metrics = append(metrics, metric{Name:name, Value: fmt.Sprintf("%v",val)} )
	}
	for name,val := range h.storage.GetAllCounter(){
		metrics = append(metrics, metric{Name:name, Value: fmt.Sprintf("%v",val)})
	}
	      w.Header().Set("Content-Type", "text/html;charset=utf-8")
      w.WriteHeader(http.StatusOK)
      if err := indexTmpl.Execute(w, metrics); err != nil {
          http.Error(w, err.Error(), http.StatusInternalServerError)
      }
}