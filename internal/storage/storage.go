package storage

import (
	"context"
	"sync"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type MemStorage struct {
	mu      sync.Mutex
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// Ping для in-memory хранилища всегда успешен: хранилище работает, пока работает процесс.
func (m *MemStorage) Ping(_ context.Context) error {
	return nil
}

func (m *MemStorage) UpdateGauge(name string, val float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauge[name] = val
	return nil
}

func (m *MemStorage) UpdateCounter(name string, val int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counter[name] += val
	return nil
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.gauge[name]
	return val, ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.counter[name]
	return val, ok
}

func (m *MemStorage) GetAllGauge() map[string]float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make(map[string]float64, len(m.gauge))
	for k, v := range m.gauge {
		res[k] = v
	}
	return res
}

func (m *MemStorage) GetAllCounter() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make(map[string]int64, len(m.counter))
	for k, v := range m.counter {
		res[k] = v
	}
	return res
}

func (m *MemStorage) UpdateMetrics(metrics []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				m.gauge[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				m.counter[metric.ID] += *metric.Delta
			}
		}
	}
	return nil
}

func (m *MemStorage) getAllMetricsLocked() []models.Metrics {
	metrics := make([]models.Metrics, 0, len(m.gauge)+len(m.counter))
	for name, val := range m.gauge {
		v := val
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		})
	}
	for name, val := range m.counter {
		v := val
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &v,
		})
	}
	return metrics
}
