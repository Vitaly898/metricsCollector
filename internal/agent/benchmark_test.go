package agent

import (
	"strconv"
	"testing"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

var (
	benchmarkBatch  []models.Metrics
	benchmarkGauges map[string]float64
)

func benchmarkGaugeValues() map[string]float64 {
	gauges := make(map[string]float64, 64)
	for i := range 64 {
		gauges["Gauge"+strconv.Itoa(i)] = float64(i) + 0.5
	}
	return gauges
}

func BenchmarkBuildBatch(b *testing.B) {
	sender := NewSender("http://localhost:8080", "")
	gauges := benchmarkGaugeValues()
	b.ReportAllocs()
	for b.Loop() {
		benchmarkBatch = sender.BuildBatch(gauges, 10)
	}
}

func BenchmarkStorageSnapshot(b *testing.B) {
	store := NewStorage()
	store.UpdateGauges(benchmarkGaugeValues())
	b.ReportAllocs()
	for b.Loop() {
		benchmarkGauges, _ = store.Snapshot()
	}
}

func BenchmarkCollector(b *testing.B) {
	collector := NewCollector()
	b.ReportAllocs()
	for b.Loop() {
		benchmarkGauges = collector.Collect()
	}
}
