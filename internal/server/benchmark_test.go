package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"testing"

	"go.uber.org/zap"

	"github.com/Vitaly898/metricsCollector/internal/agent"
	"github.com/Vitaly898/metricsCollector/internal/compress"
	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/hash"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

const profileIterations = 1000
const benchmarkKey = "benchmark-key"

type metricsWorkload struct {
	handler http.Handler
	sender  *agent.Sender
	store   *storage.MemStorage
	gauges  map[string]float64
}

func newMetricsWorkload() *metricsWorkload {
	store := storage.NewMemStorage()
	server := New(config.Config{Key: benchmarkKey}, store, zap.NewNop(), nil)
	gauges := make(map[string]float64, 64)
	for i := range 64 {
		gauges["Gauge"+strconv.Itoa(i)] = float64(i) + 0.5
	}
	return &metricsWorkload{
		handler: server.httpServer.Handler,
		sender:  agent.NewSender("http://localhost:8080", benchmarkKey),
		store:   store,
		gauges:  gauges,
	}
}

// run exercises the real agent encoding and server router without network or disk I/O.
func (w *metricsWorkload) run() error {
	body, err := json.Marshal(w.sender.BuildBatch(w.gauges, 1))
	if err != nil {
		return err
	}
	compressed, err := compress.Compress(body)
	if err != nil {
		return err
	}
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(compressed))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("HashSHA256", hash.Compute(body, benchmarkKey))
	rec := httptest.NewRecorder()
	w.handler.ServeHTTP(rec, req)
	if err := req.Body.Close(); err != nil {
		return err
	}
	if rec.Code != http.StatusOK {
		return fmt.Errorf("update returned status %d", rec.Code)
	}
	return nil
}

func BenchmarkMetricsPipeline(b *testing.B) {
	workload := newMetricsWorkload()
	b.ReportAllocs()
	for b.Loop() {
		if err := workload.run(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdatesHandler(b *testing.B) {
	store := storage.NewMemStorage()
	server := New(config.Config{}, store, zap.NewNop(), nil)
	body := []byte(`[{"id":"Alloc","type":"gauge","value":123.5},{"id":"PollCount","type":"counter","delta":1}]`)
	b.ReportAllocs()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatal("update failed")
		}
	}
}

// TestMemoryProfile is opt-in; normal test runs never write profiling artifacts.
func TestMemoryProfile(t *testing.T) {
	path := os.Getenv("METRICS_PROFILE")
	if path == "" {
		t.Skip("set METRICS_PROFILE to save a memory profile")
	}
	previousProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(previousProcs)
	if runtime.MemProfileRate != 1 {
		t.Fatal("run with -memprofilerate=1 for comparable allocation profiles")
	}
	workload := newMetricsWorkload()
	for range profileIterations {
		if err := workload.run(); err != nil {
			t.Fatal(err)
		}
	}
	if count, ok := workload.store.GetCounter("PollCount"); !ok || count != profileIterations {
		t.Fatalf("processed counter = %d, want %d", count, profileIterations)
	}
	runtime.GC()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// "allocs" includes all four memory sample types and defaults to alloc_space in pprof.
	writeErr := pprof.Lookup("allocs").WriteTo(file, 0)
	closeErr := file.Close()
	runtime.KeepAlive(workload)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	t.Logf("saved memory profile for %d batches of 65 metrics to %s", profileIterations, path)
}
