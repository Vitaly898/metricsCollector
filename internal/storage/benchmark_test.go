package storage

import (
	"strconv"
	"testing"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

var benchmarkGaugeSnapshot map[string]float64

func BenchmarkMemStorage(b *testing.B) {
	b.Run("UpdateGauge", func(b *testing.B) {
		store := NewMemStorage()
		if err := store.UpdateGauge("Alloc", 1); err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		for b.Loop() {
			if err := store.UpdateGauge("Alloc", 123.5); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("UpdateCounter", func(b *testing.B) {
		store := NewMemStorage()
		if err := store.UpdateCounter("PollCount", 1); err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		for b.Loop() {
			if err := store.UpdateCounter("PollCount", 1); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GetGauge", func(b *testing.B) {
		store := NewMemStorage()
		if err := store.UpdateGauge("Alloc", 123.5); err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		for b.Loop() {
			if value, ok := store.GetGauge("Alloc"); !ok || value != 123.5 {
				b.Fatal("unexpected gauge value")
			}
		}
	})
	b.Run("UpdateMetrics64", func(b *testing.B) {
		store := NewMemStorage()
		metrics := make([]models.Metrics, 64)
		values := make([]float64, len(metrics))
		for i := range metrics {
			values[i] = float64(i)
			metrics[i] = models.Metrics{ID: "Gauge" + strconv.Itoa(i), MType: models.Gauge, Value: &values[i]}
		}
		if err := store.UpdateMetrics(metrics); err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		for b.Loop() {
			if err := store.UpdateMetrics(metrics); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Snapshot64", func(b *testing.B) {
		store := NewMemStorage()
		for i := range 64 {
			if err := store.UpdateGauge("Gauge"+strconv.Itoa(i), float64(i)); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportAllocs()
		for b.Loop() {
			benchmarkGaugeSnapshot = store.GetAllGauge()
		}
	})
}
