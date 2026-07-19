package storage

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

type MetricsStorage interface {
	UpdateGauge(name string, val float64)
	UpdateCounter(name string, val int64)
}

func (m *MemStorage) UpdateGauge(name string, val float64) {
	m.gauge[name] = val
}
func (m *MemStorage) UpdateCounter(name string, val int64) {
	m.counter[name] += val
}
