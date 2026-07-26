package main

import (
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

func main() {
	addr:= flag.String("a","localhost:8080","адрес сервера")
	flag.Parse()
	memStorage := storage.NewMemStorage()
	updateHandler := handler.NewUpdateHandler(memStorage)
	valueHandler := handler.NewValueHandler(memStorage)
	indexHandler := handler.NewIndexHandler(memStorage)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", updateHandler.ServeHTTP)
	r.Get("/value/{type}/{name}", valueHandler.ServeHTTP)
	r.Get("/", indexHandler.ServeHTTP)

	err := http.ListenAndServe(*addr, r)
	if err != nil {
		panic(err)
	}
}