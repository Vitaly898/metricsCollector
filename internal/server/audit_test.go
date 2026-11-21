package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/Vitaly898/metricsCollector/internal/audit"
	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/handler"
	"github.com/Vitaly898/metricsCollector/internal/hash"
	models "github.com/Vitaly898/metricsCollector/internal/model"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

type auditObserverFunc func(context.Context, audit.Event) error

func (f auditObserverFunc) Notify(ctx context.Context, event audit.Event) error {
	return f(ctx, event)
}

func TestSuccessfulUpdatesPublishAudit(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		body    string
		address string
		wantIP  string
		names   []string
	}{
		{
			name: "gauge URL", path: "/update/gauge/Alloc/12.5",
			address: "192.168.0.42:1234", wantIP: "192.168.0.42", names: []string{"Alloc"},
		},
		{
			name: "counter URL IPv6", path: "/update/counter/PollCount/7",
			address: "[2001:db8::1]:1234", wantIP: "2001:db8::1", names: []string{"PollCount"},
		},
		{
			name: "JSON gauge", path: "/update", body: `{"id":"Alloc","type":"gauge","value":12.5}`,
			address: "127.0.0.1:1234", wantIP: "127.0.0.1", names: []string{"Alloc"},
		},
		{
			name: "JSON counter trailing slash", path: "/update/", body: `{"id":"PollCount","type":"counter","delta":7}`,
			address: "127.0.0.1", wantIP: "127.0.0.1", names: []string{"PollCount"},
		},
		{
			name: "batch", path: "/updates/",
			body:    `[{"id":"Alloc","type":"gauge","value":12.5},{"id":"Frees","type":"counter","delta":7},{"id":"Alloc","type":"gauge","value":13}]`,
			address: "[::1]:1234", wantIP: "::1", names: []string{"Alloc", "Frees", "Alloc"},
		},
		{
			name: "empty batch", path: "/updates/", body: `[]`,
			address: "127.0.0.1:1234", wantIP: "127.0.0.1", names: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			var events []audit.Event
			observer := auditObserverFunc(func(_ context.Context, event audit.Event) error {
				for _, name := range event.Metrics {
					_, gaugeExists := store.GetGauge(name)
					_, counterExists := store.GetCounter(name)
					if !gaugeExists && !counterExists {
						t.Errorf("metric %q was not saved before publishing the event", name)
					}
				}
				events = append(events, event)
				return nil
			})
			srv := New(config.Config{}, store, zap.NewNop(), audit.NewPublisher(observer))
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.RemoteAddr = tt.address
			req.Header.Set("X-Forwarded-For", "203.0.113.1")
			req.Header.Set("X-Real-IP", "203.0.113.2")
			rec := httptest.NewRecorder()
			before := time.Now().Unix()
			srv.httpServer.Handler.ServeHTTP(rec, req)
			after := time.Now().Unix()

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if len(events) != 1 {
				t.Fatalf("got %d events, want one per request", len(events))
			}
			event := events[0]
			if !reflect.DeepEqual(event.Metrics, tt.names) || event.IPAddress != tt.wantIP {
				t.Errorf("event = %+v, want metrics %v and IP %q", event, tt.names, tt.wantIP)
			}
			if event.Timestamp < before || event.Timestamp > after {
				t.Errorf("timestamp = %d, want between %d and %d", event.Timestamp, before, after)
			}
		})
	}
}

type failingStorage struct {
	handler.MetricsStorage
}

func (f failingStorage) UpdateGauge(string, float64) error {
	return errors.New("storage unavailable")
}

func (f failingStorage) UpdateCounter(string, int64) error {
	return errors.New("storage unavailable")
}

func (f failingStorage) UpdateMetrics([]models.Metrics) error {
	return errors.New("storage unavailable")
}

func TestFailedUpdatesDoNotPublishAudit(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		storeFails bool
		status     int
	}{
		{name: "bad gauge", path: "/update/gauge/Alloc/abc", status: 400},
		{name: "bad counter", path: "/update/counter/PollCount/1.5", status: 400},
		{name: "unknown type", path: "/update/unknown/Alloc/1", status: 400},
		{name: "missing name", path: "/update/gauge//1", status: 404},
		{name: "malformed JSON", path: "/update", body: `{`, status: 400},
		{name: "empty ID", path: "/update", body: `{"type":"gauge","value":1}`, status: 400},
		{name: "missing value", path: "/update", body: `{"id":"Alloc","type":"gauge"}`, status: 400},
		{name: "missing delta", path: "/update/", body: `{"id":"PollCount","type":"counter"}`, status: 400},
		{name: "unknown JSON type", path: "/update/", body: `{"id":"Alloc","type":"unknown","value":1}`, status: 400},
		{name: "malformed batch", path: "/updates/", body: `{`, status: 400},
		{name: "invalid batch metric", path: "/updates/", body: `[{"id":"Alloc","type":"gauge","value":1},{"id":"Frees","type":"counter"}]`, status: 400},
		{name: "gauge storage error", path: "/update/gauge/Alloc/1", storeFails: true, status: 500},
		{name: "counter storage error", path: "/update/counter/PollCount/1", storeFails: true, status: 500},
		{name: "JSON gauge storage error", path: "/update", body: `{"id":"Alloc","type":"gauge","value":1}`, storeFails: true, status: 500},
		{name: "JSON counter storage error", path: "/update/", body: `{"id":"PollCount","type":"counter","delta":1}`, storeFails: true, status: 500},
		{name: "batch storage error", path: "/updates/", body: `[{"id":"Alloc","type":"gauge","value":1}]`, storeFails: true, status: 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var store handler.MetricsStorage = storage.NewMemStorage()
			if tt.storeFails {
				store = failingStorage{MetricsStorage: store}
			}
			observer := auditObserverFunc(func(_ context.Context, event audit.Event) error {
				t.Errorf("unexpected event for unsuccessful update: %+v", event)
				return nil
			})
			srv := New(config.Config{}, store, zap.NewNop(), audit.NewPublisher(observer))
			rec := httptest.NewRecorder()
			srv.httpServer.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body)))
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
		})
	}
}

func TestReadRequestsDoNotPublishAudit(t *testing.T) {
	observer := auditObserverFunc(func(_ context.Context, event audit.Event) error {
		t.Errorf("unexpected event for a read request: %+v", event)
		return nil
	})
	store := storage.NewMemStorage()
	if err := store.UpdateGauge("Alloc", 12.5); err != nil {
		t.Fatal(err)
	}
	srv := New(config.Config{}, store, zap.NewNop(), audit.NewPublisher(observer))
	for _, path := range []string{"/", "/ping", "/value/gauge/Alloc"} {
		rec := httptest.NewRecorder()
		srv.httpServer.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, rec.Code)
		}
	}
	for _, path := range []string{"/value", "/value/"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"id":"Alloc","type":"gauge"}`))
		srv.httpServer.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("POST %s status = %d, want 200", path, rec.Code)
		}
	}
}

func TestAuditDisabledOrUnavailable(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		name := "disabled"
		var publisher handler.AuditPublisher
		if enabled {
			name = "unavailable"
			publisher = audit.NewPublisher(auditObserverFunc(func(context.Context, audit.Event) error {
				return errors.New("receiver unavailable")
			}))
		}
		t.Run(name, func(t *testing.T) {
			store := storage.NewMemStorage()
			srv := New(config.Config{}, store, zap.NewNop(), publisher)
			rec := httptest.NewRecorder()
			srv.httpServer.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/7", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rec.Code)
			}
			if got, ok := store.GetCounter("PollCount"); !ok || got != 7 {
				t.Errorf("PollCount = %d, exists: %v, want 7", got, ok)
			}
		})
	}
}

func TestAuditWithGzipAndHash(t *testing.T) {
	const key = "test-key"
	body := []byte(`[{"id":"Alloc","type":"gauge","value":12.5}]`)
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	var events []audit.Event
	observer := auditObserverFunc(func(_ context.Context, event audit.Event) error {
		events = append(events, event)
		return nil
	})
	srv := New(config.Config{Key: key}, storage.NewMemStorage(), zap.NewNop(), audit.NewPublisher(observer))
	for _, valid := range []bool{true, false} {
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(compressed.Bytes()))
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		signature := hash.Compute(body, key)
		wantStatus := http.StatusOK
		if !valid {
			signature = "invalid"
			wantStatus = http.StatusBadRequest
		}
		req.Header.Set("HashSHA256", signature)
		rec := httptest.NewRecorder()
		srv.httpServer.Handler.ServeHTTP(rec, req)
		if rec.Code != wantStatus {
			t.Errorf("valid signature: %v, status = %d, want %d", valid, rec.Code, wantStatus)
		}
		if len(events) != 1 {
			t.Fatalf("got %d events, want one for the valid request", len(events))
		}
	}
	if !reflect.DeepEqual(events[0].Metrics, []string{"Alloc"}) {
		t.Errorf("event metrics = %v, want [Alloc]", events[0].Metrics)
	}
}
