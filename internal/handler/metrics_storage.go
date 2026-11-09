// Package handler реализует HTTP-эндпоинты записи, чтения и проверки доступности метрик.
package handler

import models "github.com/Vitaly898/metricsCollector/internal/model"

// MetricsStorage задаёт операции хранения, используемые HTTP-обработчиками.
// Реализация должна поддерживать параллельное выполнение HTTP-запросов.
type MetricsStorage interface {
	// Pinger проверяет доступность хранилища.
	Pinger
	// UpdateGauge заменяет значение gauge.
	UpdateGauge(name string, val float64) error
	// UpdateCounter прибавляет val к counter.
	UpdateCounter(name string, val int64) error
	// GetGauge возвращает gauge и признак успешного получения значения.
	GetGauge(name string) (float64, bool)
	// GetCounter возвращает counter и признак успешного получения значения.
	GetCounter(name string) (int64, bool)
	// GetAllGauge возвращает снимок всех gauge-метрик.
	GetAllGauge() map[string]float64
	// GetAllCounter возвращает снимок всех counter-метрик.
	GetAllCounter() map[string]int64
	// UpdateMetrics применяет пакет обновлений gauge и counter.
	UpdateMetrics(metrics []models.Metrics) error
}
