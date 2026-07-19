package main

import (
	"net/http"

	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

func main() {
	memStorage := storage.NewMemStorage()
	updateHandler := handler.NewUpdateHandler(memStorage)
	http.Handle(`/update/`, updateHandler)
	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
