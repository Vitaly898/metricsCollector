package handler

import (
	"fmt"
	"net/http"

	"github.com/Vitaly898/metricsCollector/internal/storage"
)

type IndexHandler struct {
	storage storage.MetricsStorage
}

func NewIndexHandler(s storage.MetricsStorage) *IndexHandler{
	return  &IndexHandler{storage: s}
}
func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/html;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"<html><body><ul>")
	for name,val := range h.storage.GetAllGauge(){
		fmt.Fprintf(w,"<li>%s: %v</li>\n",name,val)
	}
	for name,val := range h.storage.GetAllCounter(){
		fmt.Fprintf(w,"<li>%s: %v</li>\n",name,val)
	}
	fmt.Fprintf(w,"</ul></body></html>")
}