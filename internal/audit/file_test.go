package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestFileObserverAppendsJSONLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	want := []Event{
		{Timestamp: 12345678, Metrics: []string{"Alloc"}, IPAddress: "192.168.0.42"},
		{Timestamp: 12345679, Metrics: []string{"Frees", "PollCount"}, IPAddress: "2001:db8::1"},
	}
	// Reopen the receiver to check that existing audit records survive a restart.
	for _, event := range want {
		observer, err := NewFileObserver(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := observer.Notify(context.Background(), event); err != nil {
			observer.Close()
			t.Fatal(err)
		}
		if err := observer.Close(); err != nil {
			t.Fatal(err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Fatal("audit record must end with a newline")
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(lines), len(want))
	}
	for i, line := range lines {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &fields); err != nil {
			t.Fatal(err)
		}
		if len(fields) != 3 || fields["ts"] == nil || fields["metrics"] == nil || fields["ip_address"] == nil {
			t.Fatalf("unexpected audit JSON fields: %s", line)
		}
		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(event, want[i]) {
			t.Errorf("event = %+v, want %+v", event, want[i])
		}
	}
}

func TestFileObserverConcurrentRequests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	observer, err := NewFileObserver(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { observer.Close() })
	publisher := NewPublisher(observer)
	const count = 100
	var wg sync.WaitGroup
	for i := range count {
		wg.Go(func() {
			event := Event{Timestamp: int64(i), Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"}
			if err := publisher.Publish(context.Background(), event); err != nil {
				t.Errorf("Publish: %v", err)
			}
		})
	}
	wg.Wait()

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	seen := make(map[int64]bool)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("corrupted audit line %q: %v", scanner.Text(), err)
		}
		if seen[event.Timestamp] {
			t.Errorf("duplicate event %d", event.Timestamp)
		}
		seen[event.Timestamp] = true
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(seen) != count {
		t.Errorf("got %d events, want %d", len(seen), count)
	}
}

func TestFileObserverErrors(t *testing.T) {
	if _, err := NewFileObserver(filepath.Join(t.TempDir(), "missing", "audit.jsonl")); err == nil {
		t.Fatal("expected an error opening a file in a missing directory")
	}
	observer, err := NewFileObserver(filepath.Join(t.TempDir(), "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := observer.Notify(context.Background(), Event{}); err == nil {
		t.Fatal("expected an error writing to a closed audit file")
	}
}
