package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/Vitaly898/metricsCollector/internal/config"
	"github.com/Vitaly898/metricsCollector/internal/hash"
	models "github.com/Vitaly898/metricsCollector/internal/model"
	"github.com/Vitaly898/metricsCollector/internal/storage"
)

func TestUnsignedJSONRequestsWithSigningKey(t *testing.T) {
	const key = "regression-key"
	initialGauge, updatedGauge := 12.5, 44.5
	initialCounter, updatedCounter := int64(7), int64(10)
	tests := []struct {
		name        string
		path        string
		body        string
		status      int
		wantGauge   float64
		wantCounter int64
		wantJSON    *models.Metrics
	}{
		{
			name: "update gauge", path: "/update/", body: `{"id":"Alloc","type":"gauge","value":44.5}`,
			status: http.StatusOK, wantGauge: updatedGauge, wantCounter: initialCounter,
			wantJSON: &models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &updatedGauge},
		},
		{
			name: "update counter", path: "/update/", body: `{"id":"PollCount","type":"counter","delta":3}`,
			status: http.StatusOK, wantGauge: initialGauge, wantCounter: updatedCounter,
			wantJSON: &models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &updatedCounter},
		},
		{
			name: "get gauge", path: "/value/", body: `{"id":"Alloc","type":"gauge"}`,
			status: http.StatusOK, wantGauge: initialGauge, wantCounter: initialCounter,
			wantJSON: &models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &initialGauge},
		},
		{
			name: "get counter", path: "/value/", body: `{"id":"PollCount","type":"counter"}`,
			status: http.StatusOK, wantGauge: initialGauge, wantCounter: initialCounter,
			wantJSON: &models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &initialCounter},
		},
		{
			name: "missing metric", path: "/value/", body: `{"id":"Missing","type":"gauge"}`,
			status: http.StatusNotFound, wantGauge: initialGauge, wantCounter: initialCounter,
		},
		{
			name: "batch", path: "/updates/",
			body:   `[{"id":"Alloc","type":"gauge","value":44.5},{"id":"PollCount","type":"counter","delta":3}]`,
			status: http.StatusOK, wantGauge: updatedGauge, wantCounter: updatedCounter,
		},
	}
	transports := []struct {
		name         string
		gzipRequest  bool
		gzipResponse bool
	}{
		{name: "JSON"},
		{name: "gzip response", gzipResponse: true},
		{name: "gzip request and response", gzipRequest: true, gzipResponse: true},
	}
	for _, transport := range transports {
		for _, tt := range tests {
			t.Run(transport.name+"/"+tt.name, func(t *testing.T) {
				store := storage.NewMemStorage()
				if err := store.UpdateGauge("Alloc", initialGauge); err != nil {
					t.Fatal(err)
				}
				if err := store.UpdateCounter("PollCount", initialCounter); err != nil {
					t.Fatal(err)
				}
				srv := New(config.Config{Key: key}, store, zap.NewNop(), nil)
				body := []byte(tt.body)
				if transport.gzipRequest {
					var compressed bytes.Buffer
					writer := gzip.NewWriter(&compressed)
					if _, err := writer.Write(body); err != nil {
						t.Fatal(err)
					}
					if err := writer.Close(); err != nil {
						t.Fatal(err)
					}
					body = compressed.Bytes()
				}
				req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json")
				if transport.gzipRequest {
					req.Header.Set("Content-Encoding", "gzip")
				}
				if transport.gzipResponse {
					req.Header.Set("Accept-Encoding", "gzip")
					req.Header.Set("Hash", "none") // Legacy header used by the course autotests.
				}
				rec := httptest.NewRecorder()
				srv.httpServer.Handler.ServeHTTP(rec, req)
				if rec.Code != tt.status {
					t.Fatalf("status = %d, want %d; body = %q", rec.Code, tt.status, rec.Body.String())
				}
				response := rec.Result()
				defer response.Body.Close()
				var reader io.Reader = response.Body
				if response.Header.Get("Content-Encoding") == "gzip" {
					decompressed, err := gzip.NewReader(response.Body)
					if err != nil {
						t.Fatal(err)
					}
					defer decompressed.Close()
					reader = decompressed
				}
				responseBody, err := io.ReadAll(reader)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := response.Header.Get("HashSHA256"), hash.Compute(responseBody, key); got != want {
					t.Errorf("response signature = %q, want %q", got, want)
				}
				if tt.wantJSON != nil {
					if got := response.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
						t.Errorf("Content-Type = %q, want application/json", got)
					}
					if transport.gzipResponse && response.Header.Get("Content-Encoding") != "gzip" {
						t.Error("JSON response was not gzip compressed")
					}
					var metric models.Metrics
					if err := json.Unmarshal(responseBody, &metric); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(metric, *tt.wantJSON) {
						t.Errorf("metric = %+v, want %+v", metric, *tt.wantJSON)
					}
				}
				if got, _ := store.GetGauge("Alloc"); got != tt.wantGauge {
					t.Errorf("Alloc = %v, want %v", got, tt.wantGauge)
				}
				if got, _ := store.GetCounter("PollCount"); got != tt.wantCounter {
					t.Errorf("PollCount = %v, want %v", got, tt.wantCounter)
				}
			})
		}
	}
}
