package agent

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type recordedBatch struct {
	method      string
	path        string
	contentType string
	encoding    string
	metrics     []models.Metrics
}

func recordBatch(t *testing.T, r *http.Request) recordedBatch {
	t.Helper()
	body := r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("cannot create gzip reader: %v", err)
		}
		defer gz.Close()
		body = gz
	}
	var metrics []models.Metrics
	if err := json.NewDecoder(body).Decode(&metrics); err != nil {
		t.Errorf("cannot decode request body: %v", err)
	}
	return recordedBatch{
		method:      r.Method,
		path:        r.URL.Path,
		contentType: r.Header.Get("Content-Type"),
		encoding:    r.Header.Get("Content-Encoding"),
		metrics:     metrics,
	}
}

func findMetric(metrics []models.Metrics, id string, mType string) (models.Metrics, bool) {
	for _, m := range metrics {
		if m.ID == id && m.MType == mType {
			return m, true
		}
	}
	return models.Metrics{}, false
}

func TestSenderSendsBatchRequest(t *testing.T) {
	var mu sync.Mutex
	var batches []recordedBatch

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batch := recordBatch(t, r)
		mu.Lock()
		batches = append(batches, batch)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSender(server.URL)
	sender.Send(map[string]float64{"Alloc": 123.5}, 7)

	mu.Lock()
	got := append([]recordedBatch(nil), batches...)
	mu.Unlock()

	if len(got) != 1 {
		t.Fatalf("expected 1 request, got %d", len(got))
	}

	batch := got[0]
	if batch.method != http.MethodPost {
		t.Errorf("request method = %q, want POST", batch.method)
	}
	if batch.path != "/updates/" {
		t.Errorf("request path = %q, want /updates/", batch.path)
	}
	if batch.contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", batch.contentType)
	}
	if batch.encoding != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", batch.encoding)
	}

	gauge, ok := findMetric(batch.metrics, "Alloc", models.Gauge)
	if !ok {
		t.Errorf("gauge Alloc not found in batch")
	}
	if gauge.Value == nil || *gauge.Value != 123.5 {
		t.Errorf("gauge value = %v, want 123.5", gauge.Value)
	}

	counter, ok := findMetric(batch.metrics, "PollCount", models.Counter)
	if !ok {
		t.Errorf("counter PollCount not found in batch")
	}
	if counter.Delta == nil || *counter.Delta != 7 {
		t.Errorf("counter delta = %v, want 7", counter.Delta)
	}
}

func TestSenderRetriesOnServerError(t *testing.T) {
	var mu sync.Mutex
	attempts := 0
	var received []models.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attempts++
		current := attempts
		mu.Unlock()

		if current < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		batch := recordBatch(t, r)
		mu.Lock()
		received = append(received, batch.metrics...)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSender(server.URL)
	sender.Send(map[string]float64{"Alloc": 42}, 1)

	mu.Lock()
	defer mu.Unlock()

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
	m, ok := findMetric(received, "Alloc", models.Gauge)
	if !ok {
		t.Fatalf(" Alloc not found in received batch")
	}
	if m.Value == nil || *m.Value != 42 {
		t.Errorf("gauge value = %v, want 42", m.Value)
	}
}

func TestSenderSendsPollCountDelta(t *testing.T) {
	var mu sync.Mutex
	var totalPollCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		body := r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Errorf("cannot create gzip reader: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer gz.Close()
			body = gz
		}
		if err := json.NewDecoder(body).Decode(&metrics); err != nil {
			t.Errorf("cannot decode request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for _, m := range metrics {
			if m.MType == models.Counter && m.ID == "PollCount" && m.Delta != nil {
				mu.Lock()
				totalPollCount += *m.Delta
				mu.Unlock()
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSender(server.URL)

	sender.Send(map[string]float64{"Alloc": 1}, 3)
	mu.Lock()
	if totalPollCount != 3 {
		t.Errorf("after first send totalPollCount = %d, want 3", totalPollCount)
	}
	mu.Unlock()

	sender.Send(map[string]float64{"Alloc": 2}, 5)
	mu.Lock()
	if totalPollCount != 5 {
		t.Errorf("after second send totalPollCount = %d, want 5", totalPollCount)
	}
	mu.Unlock()

	newSender := NewSender(server.URL)
	newSender.Send(map[string]float64{"Alloc": 3}, 2)
	mu.Lock()
	if totalPollCount != 7 {
		t.Errorf("after restart send totalPollCount = %d, want 7", totalPollCount)
	}
	mu.Unlock()
}
