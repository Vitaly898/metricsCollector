package models

import "testing"

func TestMetricsReset(t *testing.T) {
	delta, value := int64(12), 3.5
	metric := Metrics{
		ID: "Alloc", MType: Gauge, Delta: &delta, Value: &value, Hash: "hash",
	}
	metric.Reset()
	if metric.ID != "" || metric.MType != "" || metric.Hash != "" {
		t.Fatalf("scalar fields were not reset: %+v", metric)
	}
	if metric.Delta != &delta || metric.Value != &value || delta != 0 || value != 0 {
		t.Fatal("pointer fields were replaced or their values were not reset")
	}
	delta, value = 8, 2.5
	metric.Reset()
	if delta != 0 || value != 0 {
		t.Fatal("repeated reset did not reset pointed values")
	}
}

func TestMetricsResetNil(t *testing.T) {
	var metric *Metrics
	metric.Reset()
	metric = &Metrics{ID: "PollCount", MType: Counter}
	metric.Reset()
	if *metric != (Metrics{}) {
		t.Fatalf("reset allocated nil pointers or kept scalar values: %+v", metric)
	}
}
