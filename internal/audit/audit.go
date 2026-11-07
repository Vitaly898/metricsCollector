package audit

import (
	"context"
	"errors"
)

type Event struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Observer interface {
	Notify(context.Context, Event) error
}

// Publisher's observers are registered at startup and remain unchanged while serving requests.
type Publisher struct {
	observers []Observer
}

func NewPublisher(observers ...Observer) *Publisher {
	return &Publisher{observers: append([]Observer(nil), observers...)}
}

func (p *Publisher) Publish(ctx context.Context, event Event) error {
	var errs []error
	for _, observer := range p.observers {
		if err := observer.Notify(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
