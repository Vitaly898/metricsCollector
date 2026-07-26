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
	GetGauge(name string)(float64,bool)
	GetCounter(name string)(int64,bool)
	GetAllGauge() map[string]float64
	GetAllCounter() map[string]int64
}

func (m *MemStorage) UpdateGauge(name string, val float64) {
	m.gauge[name] = val
}
func (m *MemStorage) UpdateCounter(name string, val int64) {
	m.counter[name] += val
}
func (m*MemStorage) GetGauge(name string)(float64,bool){
	val,ok:= m.gauge[name]
	return val,ok
}
func (m*MemStorage)GetCounter(name string)(int64,bool){
	val,ok := m.counter[name]
	return val,ok
}
func (m*MemStorage) GetAllGauge()map[string]float64{
	res := make(map[string]float64, len(m.gauge))
	for k,v := range m.gauge{
		res[k]=v
	}
	return res
}
func (m*MemStorage)GetAllCounter()map[string]int64{
	res := make(map[string]int64,len(m.counter))
	for k,v := range m.counter {
		res[k]=v
	}
	return  res
}
