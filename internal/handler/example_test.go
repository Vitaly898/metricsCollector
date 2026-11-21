package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

// ExampleUpdateHandler_ServeHTTP показывает увеличение counter через параметры URL.
func ExampleUpdateHandler_ServeHTTP() {
	store := storage.NewMemStorage()
	if err := store.UpdateCounter("PollCount", 2); err != nil {
		panic(err)
	}
	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(store, nil).ServeHTTP)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/3", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	value, _ := store.GetCounter("PollCount")
	fmt.Println(rec.Code)
	fmt.Println(value)

	// Output:
	// 200
	// 5
}

// ExampleUpdateJSONHandler_ServeHTTP показывает запись gauge через JSON API.
func ExampleUpdateJSONHandler_ServeHTTP() {
	store := storage.NewMemStorage()
	h := handler.NewUpdateJSONHandler(store, nil)
	body := strings.NewReader(`{"id":"Alloc","type":"gauge","value":12.5}`)
	req := httptest.NewRequest(http.MethodPost, "/update/", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Print(rec.Body.String())

	// Output:
	// 200
	// application/json
	// {"id":"Alloc","type":"gauge","value":12.5}
}

// ExampleUpdatesHandler_ServeHTTP показывает запись gauge и counter одним запросом.
func ExampleUpdatesHandler_ServeHTTP() {
	store := storage.NewMemStorage()
	h := handler.NewUpdatesHandler(store, nil)
	body := strings.NewReader(`[
		{"id":"Alloc","type":"gauge","value":12.5},
		{"id":"PollCount","type":"counter","delta":3}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/updates/", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	gauge, _ := store.GetGauge("Alloc")
	counter, _ := store.GetCounter("PollCount")
	fmt.Println(rec.Code)
	fmt.Println(gauge, counter)

	// Output:
	// 200
	// 12.5 3
}

// ExampleValueHandler_ServeHTTP показывает чтение значения метрики как текста.
func ExampleValueHandler_ServeHTTP() {
	store := storage.NewMemStorage()
	if err := store.UpdateGauge("Alloc", 12.5); err != nil {
		panic(err)
	}
	router := chi.NewRouter()
	router.Get("/value/{type}/{name}", handler.NewValueHandler(store).ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(rec.Body.String())

	// Output:
	// 200
	// text/plain
	// 12.5
}

// ExampleValueJSONHandler_ServeHTTP показывает чтение counter по id и type.
func ExampleValueJSONHandler_ServeHTTP() {
	store := storage.NewMemStorage()
	if err := store.UpdateCounter("PollCount", 7); err != nil {
		panic(err)
	}
	h := handler.NewValueJSONHandler(store)
	body := strings.NewReader(`{"id":"PollCount","type":"counter"}`)
	req := httptest.NewRequest(http.MethodPost, "/value/", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Print(rec.Body.String())

	// Output:
	// 200
	// application/json
	// {"id":"PollCount","type":"counter","delta":7}
}

// ExamplePingHandler_ServeHTTP показывает проверку доступности хранилища в памяти.
func ExamplePingHandler_ServeHTTP() {
	h := handler.NewPingHandler(storage.NewMemStorage())
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.Len())

	// Output:
	// 200
	// 0
}

// ExampleIndexHandler_ServeHTTP показывает HTML-страницу со значениями метрик.
func ExampleIndexHandler_ServeHTTP() {
	store := storage.NewMemStorage()
	if err := store.UpdateGauge("Alloc", 12.5); err != nil {
		panic(err)
	}
	h := handler.NewIndexHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Content-Type"))
	fmt.Println(strings.TrimSpace(rec.Body.String()))

	// Output:
	// 200
	// text/html;charset=utf-8
	// <html><body><ul>
	//   <li>Alloc: 12.5</li>
	//   </ul></body></html>
}
