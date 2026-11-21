package audit

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

type discardObserver struct{}

func (discardObserver) Notify(context.Context, Event) error { return nil }

func BenchmarkPublisher(b *testing.B) {
	for _, count := range []int{0, 1, 2, 8} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			observers := make([]Observer, count)
			for i := range observers {
				observers[i] = discardObserver{}
			}
			publisher := NewPublisher(observers...)
			event := Event{Timestamp: 12345678, Metrics: []string{"Alloc", "Frees"}, IPAddress: "192.168.0.42"}
			b.ReportAllocs()
			for b.Loop() {
				if err := publisher.Publish(context.Background(), event); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkFileObserver(b *testing.B) {
	observer, err := NewFileObserver(filepath.Join(b.TempDir(), "audit.jsonl"))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { observer.Close() })
	event := Event{Timestamp: 12345678, Metrics: []string{"Alloc", "Frees"}, IPAddress: "192.168.0.42"}
	b.ReportAllocs()
	for b.Loop() {
		if err := observer.Notify(context.Background(), event); err != nil {
			b.Fatal(err)
		}
	}
}
