package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestHTTPObserverPostsEvent(t *testing.T) {
	want := Event{Timestamp: 12345678, Metrics: []string{"Alloc", "Frees"}, IPAddress: "192.168.0.42"}
	requests := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.RequestURI() != "/audit?source=metrics" {
			t.Errorf("request = %s %s, want POST /audit?source=metrics", r.Method, r.URL.RequestURI())
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read audit request: %v", err)
		}
		requests <- body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	observer, err := NewHTTPObserver(server.URL + "/audit?source=metrics")
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.Notify(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	var got Event
	if err := json.Unmarshal(<-requests, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("event = %+v, want %+v", got, want)
	}
}

func TestPublisherDeliversToFileAndHTTP(t *testing.T) {
	for _, status := range []int{http.StatusNoContent, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			requests := make(chan Event, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var event Event
				if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
					t.Errorf("decode HTTP event: %v", err)
				}
				requests <- event
				w.WriteHeader(status)
			}))
			defer server.Close()
			remote, err := NewHTTPObserver(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "audit.jsonl")
			file, err := NewFileObserver(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			want := Event{Timestamp: 12345678, Metrics: []string{"Alloc"}, IPAddress: "192.168.0.42"}
			err = NewPublisher(remote, file).Publish(context.Background(), want)
			if (err != nil) != (status == http.StatusInternalServerError) {
				t.Fatalf("Publish error = %v, HTTP receiver status = %d", err, status)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var got Event
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("file event = %+v, want %+v", got, want)
			}
			if received := <-requests; !reflect.DeepEqual(received, want) {
				t.Errorf("HTTP event = %+v, want %+v", received, want)
			}
		})
	}
}

func TestHTTPObserverResponseStatus(t *testing.T) {
	for _, status := range []int{200, 204, 299, 300, 400, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			}))
			defer server.Close()
			observer, err := NewHTTPObserver(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			err = observer.Notify(context.Background(), Event{})
			wantErr := status >= 300
			if (err != nil) != wantErr {
				t.Errorf("Notify error = %v, want error: %v", err, wantErr)
			}
		})
	}
}

func TestHTTPObserverDoesNotFollowRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Error("audit request followed a redirect")
			return
		}
		http.Redirect(w, r, "/redirected", http.StatusFound)
	}))
	defer server.Close()
	observer, err := NewHTTPObserver(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.Notify(context.Background(), Event{}); err == nil {
		t.Fatal("expected an error for a redirect response")
	}
}

func TestHTTPObserverTimeoutAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer server.Close()
	observer, err := NewHTTPObserver(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if observer.client.Timeout != 2*time.Second {
		t.Fatalf("HTTP timeout = %v, want 2s", observer.client.Timeout)
	}
	observer.client.Timeout = 50 * time.Millisecond
	if err := observer.Notify(context.Background(), Event{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Notify error = %v, want deadline exceeded", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := observer.Notify(ctx, Event{}); !errors.Is(err, context.Canceled) {
		t.Errorf("Notify error = %v, want context canceled", err)
	}
}

func TestHTTPObserverValidatesURL(t *testing.T) {
	for _, address := range []string{"", "/audit", "localhost:8080/audit", "ftp://example.com/audit", "http://", "http://:8080/audit", "http://example.com:invalid"} {
		t.Run(address, func(t *testing.T) {
			if _, err := NewHTTPObserver(address); err == nil {
				t.Errorf("expected an error for URL %q", address)
			}
		})
	}
	for _, address := range []string{"http://localhost:8080/audit", "https://example.com/audit", "http://[::1]:8080/audit"} {
		if _, err := NewHTTPObserver(address); err != nil {
			t.Errorf("valid URL %q: %v", address, err)
		}
	}
}
