// Package models определяет типы метрик и их представление в JSON API.
package models

const (
	// Counter обозначает накопительную целочисленную метрику.
	Counter = "counter"
	// Gauge обозначает метрику с заменяемым значением float64.
	Gauge = "gauge"
)

// Metrics представляет одну gauge- или counter-метрику в JSON-запросах и ответах.
type Metrics struct {
	// ID содержит имя метрики.
	ID string `json:"id"`
	// MType задаёт тип: Counter или Gauge.
	MType string `json:"type"`
	// Delta содержит приращение counter в запросе или текущее значение в ответе.
	Delta *int64 `json:"delta,omitempty"`
	// Value содержит значение gauge; nil означает отсутствие значения.
	Value *float64 `json:"value,omitempty"`
	// Hash сохраняет необязательное поле hash из JSON; HTTP-подпись передаётся в заголовке HashSHA256.
	Hash string `json:"hash,omitempty"`
}
