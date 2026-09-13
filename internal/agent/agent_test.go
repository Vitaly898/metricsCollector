package agent

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	models "github.com/Vitaly898/metricsCollector/internal/model"
	"sync/atomic"
)

func TestAgentRunSendsMetrics(t *testing.T) {
	var mu sync.Mutex
	var metrics []models.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch []models.Metrics
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
		if err := json.NewDecoder(body).Decode(&batch); err == nil {
			mu.Lock()
			metrics = append(metrics, batch...)
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	collector := NewCollector()
	sender := NewSender(server.URL, "")
	a := NewAgent(collector, sender, 10*time.Millisecond, 25*time.Millisecond, 2)

	go a.Run()
	time.Sleep(80 * time.Millisecond)
	a.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(metrics) == 0 {
		t.Fatal("no requests received")
	}

	hasGauge := false
	hasCounter := false
	for _, m := range metrics {
		if m.MType == models.Gauge {
			hasGauge = true
		}
		if m.MType == models.Counter && m.ID == "PollCount" {
			hasCounter = true
		}
	}

	if !hasGauge {
		t.Errorf("expected at least one gauge request, got %v", metrics)
	}
	if !hasCounter {
		t.Errorf("expected at least one counter PollCount request, got %v", metrics)
	}
}

func TestAgentRateLimitCapsConcurrentRequests(t *testing.T) {
	const rateLimit = 2

	var current atomic.Int64
	var maxSeen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := current.Add(1)
		for {
			max := maxSeen.Load()
			if cur <= max || maxSeen.CompareAndSwap(max, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		current.Add(-1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	collector := NewCollector()
	sender := NewSender(server.URL, "")
	a := NewAgent(collector, sender, 5*time.Millisecond, 10*time.Millisecond, rateLimit)

	go a.Run()
	time.Sleep(300 * time.Millisecond)
	a.Stop()

	if got := maxSeen.Load(); got > rateLimit {
		t.Errorf("max concurrent requests = %d, want <= %d", got, rateLimit)
	}
	if maxSeen.Load() == 0 {
		t.Error("no requests received")
	}
}
