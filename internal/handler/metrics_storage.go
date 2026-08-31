package handler

import models "github.com/Vitaly898/metricsCollector/internal/model"

type MetricsStorage interface {
	UpdateGauge(name string, val float64)
	UpdateCounter(name string, val int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauge() map[string]float64
	GetAllCounter() map[string]int64
	UpdateMetrics(metrics []models.Metrics) error
}
