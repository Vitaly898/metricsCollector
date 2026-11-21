package audit

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type observerFunc func(context.Context, Event) error

func (f observerFunc) Notify(ctx context.Context, event Event) error {
	return f(ctx, event)
}

func TestPublisherNotifiesEveryObserver(t *testing.T) {
	want := Event{Timestamp: 12345678, Metrics: []string{"Alloc", "Frees"}, IPAddress: "192.168.0.42"}
	firstErr := errors.New("first receiver failed")
	lastErr := errors.New("last receiver failed")
	errs := []error{firstErr, nil, lastErr}
	var received []Event
	observers := make([]Observer, 0, len(errs))
	for _, err := range errs {
		observers = append(observers, observerFunc(func(_ context.Context, event Event) error {
			received = append(received, event)
			return err
		}))
	}

	err := NewPublisher(observers...).Publish(context.Background(), want)
	if !errors.Is(err, firstErr) || !errors.Is(err, lastErr) {
		t.Fatalf("Publish error = %v, want both receiver errors", err)
	}
	if len(received) != len(observers) {
		t.Fatalf("received %d events, want %d", len(received), len(observers))
	}
	for _, event := range received {
		if !reflect.DeepEqual(event, want) {
			t.Errorf("event = %+v, want %+v", event, want)
		}
	}
}

func TestPublisherWithoutObservers(t *testing.T) {
	if err := NewPublisher().Publish(context.Background(), Event{}); err != nil {
		t.Errorf("Publish error = %v, want nil", err)
	}
}
