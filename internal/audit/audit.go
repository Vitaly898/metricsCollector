// Package audit рассылает события аудита файловым и HTTP-приёмникам через наблюдателей.
package audit

import (
	"context"
	"errors"
)

// Event описывает успешно обработанный запрос с метриками.
type Event struct {
	// Timestamp содержит Unix-время события в секундах.
	Timestamp int64 `json:"ts"`
	// Metrics содержит имена метрик в порядке запроса, включая повторяющиеся.
	Metrics []string `json:"metrics"`
	// IPAddress содержит IP клиента без номера порта.
	IPAddress string `json:"ip_address"`
}

// Observer принимает события аудита и доставляет их конкретному приёмнику.
type Observer interface {
	// Notify доставляет одно событие и возвращает ошибку доставки, если она возникла.
	Notify(context.Context, Event) error
}

// Publisher синхронно уведомляет наблюдателей, зарегистрированных при создании.
// Параллельные вызовы Publish требуют потокобезопасных наблюдателей.
type Publisher struct {
	observers []Observer
}

// NewPublisher создаёт издателя с копией списка наблюдателей.
func NewPublisher(observers ...Observer) *Publisher {
	return &Publisher{observers: append([]Observer(nil), observers...)}
}

// Publish уведомляет всех наблюдателей и объединяет ошибки через errors.Join.
// Ошибка одного приёмника не прерывает уведомление остальных.
func (p *Publisher) Publish(ctx context.Context, event Event) error {
	var errs []error
	for _, observer := range p.observers {
		if err := observer.Notify(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
