package agent

import "sync"

// Storage хранит последние gauge-метрики и число опросов агента.
// Методы безопасны для параллельных вызовов. Создаётся через NewStorage.
type Storage struct {
	mu        sync.Mutex
	gauges    map[string]float64
	pollCount int64
}

// NewStorage создаёт пустое хранилище агента с нулевым числом опросов.
func NewStorage() *Storage {
	return &Storage{gauges: make(map[string]float64)}
}

// UpdateGauges заменяет значения переданных метрик, сохраняя остальные метрики.
func (s *Storage) UpdateGauges(m map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range m {
		s.gauges[k] = v
	}
}

// AddPollCount увеличивает число опросов на единицу.
func (s *Storage) AddPollCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pollCount++
}

// Snapshot возвращает копию gauge-метрик и текущее число опросов.
// Изменения возвращённой map не влияют на хранилище.
func (s *Storage) Snapshot() (map[string]float64, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		out[k] = v
	}
	return out, s.pollCount
}
