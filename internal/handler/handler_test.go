package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type mockStorage struct {
	gaugeCalls   map[string]float64
	counterCalls map[string]int64
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		gaugeCalls:   make(map[string]float64),
		counterCalls: make(map[string]int64),
	}
}

func (m *mockStorage) UpdateGauge(name string, value float64) {
	m.gaugeCalls[name] = value
}

func (m *mockStorage) UpdateCounter(name string, value int64) {
	m.counterCalls[name] += value
}
func (m *mockStorage) GetGauge(name string) (float64, bool) {
	val, ok := m.gaugeCalls[name]
	return val, ok
}

func (m *mockStorage) GetCounter(name string) (int64, bool) {
	val, ok := m.counterCalls[name]
	return val, ok
}

func (m *mockStorage) GetAllGauge() map[string]float64 {
	return m.gaugeCalls
}

func (m *mockStorage) GetAllCounter() map[string]int64 {
	return m.counterCalls
}

func (m *mockStorage) UpdateMetrics(metrics []models.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				m.gaugeCalls[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				m.counterCalls[metric.ID] += *metric.Delta
			}
		}
	}
	return nil
}

func TestUpdateHandler(t *testing.T) {
	type want struct {
		statusCode int
		gauge      map[string]float64
		counter    map[string]int64
	}

	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "valid gauge",
			method: http.MethodPost,
			url:    "/update/gauge/Alloc/12.5",
			want: want{
				statusCode: http.StatusOK,
				gauge:      map[string]float64{"Alloc": 12.5},
			},
		},
		{
			name:   "valid counter",
			method: http.MethodPost,
			url:    "/update/counter/PollCount/7",
			want: want{
				statusCode: http.StatusOK,
				counter:    map[string]int64{"PollCount": 7},
			},
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			url:    "/update/gauge/Alloc/12.5",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "invalid path length",
			method: http.MethodPost,
			url:    "/update/gauge/Alloc/12.5/extra",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:   "invalid metric type",
			method: http.MethodPost,
			url:    "/update/unknown/Alloc/12.5",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "empty metric name",
			method: http.MethodPost,
			url:    "/update/gauge//12.5",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:   "invalid gauge value",
			method: http.MethodPost,
			url:    "/update/gauge/Alloc/abc",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "invalid counter value",
			method: http.MethodPost,
			url:    "/update/counter/PollCount/12.5",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockStorage()
			h := NewUpdateHandler(mock)
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", h.ServeHTTP)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.statusCode {
				t.Errorf("status code = %d, want %d", res.StatusCode, tt.want.statusCode)
			}

			for name, wantVal := range tt.want.gauge {
				if got := mock.gaugeCalls[name]; got != wantVal {
					t.Errorf("gauge[%s] = %v, want %v", name, got, wantVal)
				}
			}

			for name, wantVal := range tt.want.counter {
				if got := mock.counterCalls[name]; got != wantVal {
					t.Errorf("counter[%s] = %v, want %v", name, got, wantVal)
				}
			}
		})
	}
}
