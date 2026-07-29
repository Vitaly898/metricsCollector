package storage

import "sync"

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

func (m *MemStorage) UpdateGauge(name string, val float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauge[name] = val
}

func (m *MemStorage) UpdateCounter(name string, val int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counter[name] += val
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
