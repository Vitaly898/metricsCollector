package storage

import (
	"testing"
)

func TestUpdateStorage(t *testing.T){
	tests := []struct {
		name string
		values []float64
		want float64
	}{
		{name: "позитивный кейс- обновление метрики один вызов",values:[]float64{3.1}, want: 3.1 },
		{name: "позитивный кейс- обновление метрики два вызова",values: []float64{3.1, 5.5},want:5.5 },
		{name: "позитивный кейс- обновление метрики два вызова",values: []float64{0},want:0 },
		{name: "позитивный кейс- обновление метрики два вызова",values: []float64{-100.0},want:-100 },


	}

	for _,tt := range tests {
		t.Run(tt.name,func(t *testing.T){
			s := NewMemStorage()
			for _,v := range tt.values{
				s.UpdateGauge("Alloc",v)
			}
			if got:= s.gauge["Alloc"]; got !=tt.want{
				t.Errorf("Тест упал %v не равен %v", got,tt.want)
			}
		})
	}

}