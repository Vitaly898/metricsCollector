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
)

func TestAgentRunSendsMetrics(t *testing.T) {
	var mu sync.Mutex
	var metrics []models.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
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
		if err := json.NewDecoder(body).Decode(&m); err == nil {
			mu.Lock()
			metrics = append(metrics, m)
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	collector := NewCollector()
	sender := NewSender(server.URL)
	a := NewAgent(collector, sender, 10*time.Millisecond, 25*time.Millisecond)

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
