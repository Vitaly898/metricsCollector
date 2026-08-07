package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAgentRunSendsMetrics(t *testing.T) {
	var mu sync.Mutex
	var paths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
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

	if len(paths) == 0 {
		t.Fatal("no requests received")
	}

	hasGauge := false
	hasCounter := false
	for _, p := range paths {
		if strings.HasPrefix(p, "/update/gauge/") {
			hasGauge = true
		}
		if strings.HasPrefix(p, "/update/counter/PollCount/") {
			hasCounter = true
		}
	}

	if !hasGauge {
		t.Errorf("expected at least one gauge request, got %v", paths)
	}
	if !hasCounter {
		t.Errorf("expected at least one counter PollCount request, got %v", paths)
	}
}
