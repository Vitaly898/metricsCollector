package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type recordedRequest struct {
	method      string
	path        string
	contentType string
	metric      models.Metrics
}

func recordRequest(t *testing.T, r *http.Request) recordedRequest {
	t.Helper()
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		t.Errorf("cannot decode request body: %v", err)
	}
	return recordedRequest{
		method:      r.Method,
		path:        r.URL.Path,
		contentType: r.Header.Get("Content-Type"),
		metric:      m,
	}
}

func TestSenderSendsGaugeRequest(t *testing.T) {
	var mu sync.Mutex
	var reqs []recordedRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recordRequest(t, r)
		mu.Lock()
		reqs = append(reqs, rec)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSender(server.URL)
	sender.Send(map[string]float64{"Alloc": 123.5}, 1)

	mu.Lock()
	got := append([]recordedRequest(nil), reqs...)
	mu.Unlock()

	if len(got) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(got))
	}

	gauge := got[0]
	if gauge.method != http.MethodPost {
		t.Errorf("gauge request method = %q, want POST", gauge.method)
	}
	if gauge.path != "/update" {
		t.Errorf("gauge request path = %q, want /update", gauge.path)
	}
	if gauge.contentType != "application/json" {
		t.Errorf("gauge Content-Type = %q, want application/json", gauge.contentType)
	}
	if gauge.metric.ID != "Alloc" || gauge.metric.MType != models.Gauge {
		t.Errorf("gauge metric = %+v, want id=Alloc type=gauge", gauge.metric)
	}
	if gauge.metric.Value == nil || *gauge.metric.Value != 123.5 {
		t.Errorf("gauge value = %v, want 123.5", gauge.metric.Value)
	}
}

func TestSenderSendsCounterRequest(t *testing.T) {
	var mu sync.Mutex
	var reqs []recordedRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recordRequest(t, r)
		mu.Lock()
		reqs = append(reqs, rec)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSender(server.URL)
	sender.Send(map[string]float64{"Alloc": 1}, 7)

	mu.Lock()
	got := append([]recordedRequest(nil), reqs...)
	mu.Unlock()

	if len(got) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(got))
	}

	counter := got[1]
	if counter.method != http.MethodPost {
		t.Errorf("counter request method = %q, want POST", counter.method)
	}
	if counter.path != "/update" {
		t.Errorf("counter request path = %q, want /update", counter.path)
	}
	if counter.contentType != "application/json" {
		t.Errorf("counter Content-Type = %q, want application/json", counter.contentType)
	}
	if counter.metric.ID != "PollCount" || counter.metric.MType != models.Counter {
		t.Errorf("counter metric = %+v, want id=PollCount type=counter", counter.metric)
	}
	if counter.metric.Delta == nil || *counter.metric.Delta != 7 {
		t.Errorf("counter delta = %v, want 7", counter.metric.Delta)
	}
}

func TestSenderSendsPollCountDelta(t *testing.T) {
	var mu sync.Mutex
	var totalPollCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Errorf("cannot decode request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if m.MType == models.Counter && m.ID == "PollCount" && m.Delta != nil {
			mu.Lock()
			totalPollCount += *m.Delta
			mu.Unlock()
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

	// Имитируем перезапуск агента: новый Sender.
	newSender := NewSender(server.URL)
	newSender.Send(map[string]float64{"Alloc": 3}, 2)
	mu.Lock()
	if totalPollCount != 7 {
		t.Errorf("after restart send totalPollCount = %d, want 7", totalPollCount)
	}
	mu.Unlock()
}
