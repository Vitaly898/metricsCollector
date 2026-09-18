package agent

import (
	"fmt"
	"runtime"
	"testing"
)

func TestCollectorCollect(t *testing.T) {
	c := NewCollector()
	gauges := c.Collect()

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

func TestCollectorRandomValueInRange(t *testing.T) {
	c := NewCollector()
	gauges := c.Collect()

	rv, ok := gauges["RandomValue"]
	if !ok {
		t.Fatal("RandomValue metric not found")
	}

	if rv < 0 || rv >= 1 {
		t.Errorf("RandomValue = %v, want [0, 1)", rv)
	}
}

func TestCollectorCollectSystem(t *testing.T) {
	c := NewCollector()
	gauges := c.CollectSystem()

	for _, key := range []string{"TotalMemory", "FreeMemory", "CPUutilization1"} {
		if _, ok := gauges[key]; !ok {
			t.Errorf("expected metric %q not found in gauges", key)
		}
	}

	wantCPU := runtime.NumCPU()
	for i := 1; i <= wantCPU; i++ {
		key := fmt.Sprintf("CPUutilization%d", i)
		if _, ok := gauges[key]; !ok {
			t.Errorf("expected metric %q not found in gauges", key)
		}
	}
	if _, ok := gauges[fmt.Sprintf("CPUutilization%d", wantCPU+1)]; ok {
		t.Errorf("unexpected extra metric CPUutilization%d", wantCPU+1)
	}
}

func TestStorageSnapshotReturnsCopy(t *testing.T) {
	s := NewStorage()
	s.UpdateGauges(map[string]float64{"Alloc": 1})
	s.AddPollCount()

	gauges, pollCount := s.Snapshot()
	if gauges["Alloc"] != 1 {
		t.Errorf("Alloc = %v, want 1", gauges["Alloc"])
	}
	if pollCount != 1 {
		t.Errorf("pollCount = %d, want 1", pollCount)
	}

	gauges["Alloc"] = 999
	again, _ := s.Snapshot()
	if again["Alloc"] != 1 {
		t.Errorf("snapshot is not a copy: Alloc = %v, want 1", again["Alloc"])
	}
}
