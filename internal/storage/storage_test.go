package storage

import (
	"testing"
)

func TestUpdateGauge(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{name: "positive case: single gauge update", values: []float64{3.1}, want: 3.1},
		{name: "positive case: two gauge updates", values: []float64{3.1, 5.5}, want: 5.5},
		{name: "positive case: gauge value zero", values: []float64{0}, want: 0},
		{name: "positive case: negative gauge value", values: []float64{-100.0}, want: -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			for _, v := range tt.values {
				s.UpdateGauge("Alloc", v)
			}
			got, ok := s.GetGauge("Alloc")
			if !ok {
				t.Errorf("Metric not found")
			}
			if got != tt.want {
				t.Errorf("Test failed: %v is not equal to %v", got, tt.want)
			}
		})
	}

}

func TestUpdateCounter(t *testing.T) {
	tests := []struct {
		name   string
		values []int64
		want   int64
	}{
		{name: "positive case: single counter update", values: []int64{3}, want: 3},
		{name: "positive case: counter values are summed", values: []int64{3, 1}, want: 4},
		{name: "positive case: negative counter value", values: []int64{3, 1, -2}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			for _, v := range tt.values {
				s.UpdateCounter("PollCount", v)
			}
			got, ok := s.GetCounter("PollCount")
			if !ok {
				t.Errorf("Metric not found")
			}
			if got != tt.want {
				t.Errorf("Test failed: metric counter is incorrect")
			}
		})
	}
}
