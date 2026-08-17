package handler

type MetricsStorage interface {
	UpdateGauge(name string, val float64)
	UpdateCounter(name string, val int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauge() map[string]float64
	GetAllCounter() map[string]int64
}
