package handler

import models "github.com/Vitaly898/metricsCollector/internal/model"

type MetricsStorage interface {
	Pinger
	UpdateGauge(name string, val float64) error
	UpdateCounter(name string, val int64) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauge() map[string]float64
	GetAllCounter() map[string]int64
	UpdateMetrics(metrics []models.Metrics) error
}
