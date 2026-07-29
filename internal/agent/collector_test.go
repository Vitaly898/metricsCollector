package agent

import (
	"testing"
)

func TestCollectorCollect(t *testing.T) {
	c := NewCollector()
	gauges, pollCount := c.Collect()

	if pollCount != 1 {
		t.Errorf("pollCount = %d, want 1", pollCount)
	}

	expectedKeys := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}

	if len(gauges) != len(expectedKeys) {
		t.Errorf("gauges count = %d, want %d", len(gauges), len(expectedKeys))
	}

	for _, key := range expectedKeys {
		if _, ok := gauges[key]; !ok {
			t.Errorf("expected metric %q not found in gauges", key)
		}
	}
}

func TestCollectorPollCountIncrements(t *testing.T) {
	c := NewCollector()

	for i := 1; i <= 5; i++ {
		_, pollCount := c.Collect()
		if pollCount != int64(i) {
			t.Errorf("pollCount = %d, want %d", pollCount, i)
		}
	}
}

func TestCollectorRandomValueInRange(t *testing.T) {
	c := NewCollector()
	gauges, _ := c.Collect()

	rv, ok := gauges["RandomValue"]
	if !ok {
		t.Fatal("RandomValue metric not found")
	}

	if rv < 0 || rv >= 1 {
		t.Errorf("RandomValue = %v, want [0, 1)", rv)
	}
}
