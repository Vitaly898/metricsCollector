package agent

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type recordedRequest struct {
	method      string
	path        string
	contentType string
}

func TestSenderSendsGaugeRequest(t *testing.T) {
	var mu sync.Mutex
	var reqs []recordedRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		reqs = append(reqs, recordedRequest{
			method:      r.Method,
			path:        r.URL.Path,
			contentType: r.Header.Get("Content-Type"),
		})
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
	if gauge.path != "/update/gauge/Alloc/123.5" {
		t.Errorf("gauge request path = %q, want /update/gauge/Alloc/123.5", gauge.path)
	}
	if gauge.contentType != "text/plain" {
		t.Errorf("gauge Content-Type = %q, want text/plain", gauge.contentType)
	}
}

func TestSenderSendsCounterRequest(t *testing.T) {
	var mu sync.Mutex
	var reqs []recordedRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		reqs = append(reqs, recordedRequest{
			method:      r.Method,
			path:        r.URL.Path,
			contentType: r.Header.Get("Content-Type"),
		})
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
	if counter.path != "/update/counter/PollCount/7" {
		t.Errorf("counter request path = %q, want /update/counter/PollCount/7", counter.path)
	}
	if counter.contentType != "text/plain" {
		t.Errorf("counter Content-Type = %q, want text/plain", counter.contentType)
	}
}

func TestSenderSendsPollCountDelta(t *testing.T) {
	var mu sync.Mutex
	var totalPollCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/update/")
		parts := strings.Split(path, "/")
		if len(parts) == 3 && parts[0] == "counter" && parts[1] == "PollCount" {
			if v, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				mu.Lock()
				totalPollCount += v
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

	// Имитируем перезапуск агента: новый Sender.
	newSender := NewSender(server.URL)
	newSender.Send(map[string]float64{"Alloc": 3}, 2)
	mu.Lock()
	if totalPollCount != 7 {
		t.Errorf("after restart send totalPollCount = %d, want 7", totalPollCount)
	}
	mu.Unlock()
}
