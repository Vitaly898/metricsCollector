package agent

import "sync"

type Storage struct {
	mu        sync.RWMutex
	gauges    map[string]float64
	pollCount int64
}

func NewStorage() *Storage {
	return &Storage{gauges: make(map[string]float64)}
}

func (s *Storage) UpdateGauges(m map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range m {
		s.gauges[k] = v
	}
}

func (s *Storage) AddPollCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pollCount++
}

func (s *Storage) Snapshot() (map[string]float64, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		out[k] = v
	}
	return out, s.pollCount
}
