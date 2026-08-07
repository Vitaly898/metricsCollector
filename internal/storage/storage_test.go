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
		{name: "позитивный кейс- обновление метрики один вызов", values: []float64{3.1}, want: 3.1},
		{name: "позитивный кейс- обновление метрики два вызова", values: []float64{3.1, 5.5}, want: 5.5},
		{name: "позитивный кейс- обновление метрики два вызова", values: []float64{0}, want: 0},
		{name: "позитивный кейс- обновление метрики два вызова", values: []float64{-100.0}, want: -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			for _, v := range tt.values {
				s.UpdateGauge("Alloc", v)
			}
			got,ok := s.GetGauge("Alloc");
			if !ok{
				t.Errorf("Метрика не найдена")
			}
			if  got != tt.want {
				t.Errorf("Тест упал %v не равен %v", got, tt.want)
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
		{name: "позитивный кейс- обновление метрики один вызов", values: []int64{3}, want: 3},
		{name: "позитивный кейс- обновление метрики числа суммируются", values: []int64{3, 1}, want: 4},
		{name: "позитивный кейс- обновление метрики отрицательное значение метрики", values: []int64{3, 1, -2}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			for _, v := range tt.values {
				s.UpdateCounter("PollCount", v)
			}
			got,ok := s.GetCounter("PollCount")
			if !ok{
				t.Errorf("Метрика не найдена")
			}
			if ; got != tt.want {
				t.Errorf("Тест упал счетчик метрики не верный")
			}
		})
	}
}
