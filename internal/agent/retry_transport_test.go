package agent

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type stubRoundTripper struct {
	calls   int
	handler func(call int) (*http.Response, error)
}

func (s *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	s.calls++
	return s.handler(s.calls)
}

func okResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

func errorResponse(code int) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

func newTestRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "http://example.com/updates/", strings.NewReader("body"))
	if err != nil {
		t.Fatalf("cannot create request: %v", err)
	}
	return req
}

func TestRetryRoundTripperRetriesOnNetworkError(t *testing.T) {
	stub := &stubRoundTripper{
		handler: func(call int) (*http.Response, error) {
			if call < 3 {
				return nil, errors.New("network error")
			}
			return okResponse(), nil
		},
	}

	rt := NewRetryRoundTripper(stub, []time.Duration{0, 0, 0})
	resp, err := rt.RoundTrip(newTestRequest(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if stub.calls != 3 {
		t.Errorf("expected 3 attempts, got %d", stub.calls)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRetryRoundTripperRetriesOnServerError(t *testing.T) {
	stub := &stubRoundTripper{
		handler: func(call int) (*http.Response, error) {
			if call < 2 {
				return errorResponse(http.StatusServiceUnavailable), nil
			}
			return okResponse(), nil
		},
	}

	rt := NewRetryRoundTripper(stub, []time.Duration{0, 0, 0})
	resp, err := rt.RoundTrip(newTestRequest(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if stub.calls != 2 {
		t.Errorf("expected 2 attempts, got %d", stub.calls)
	}
}

func TestRetryRoundTripperDoesNotRetryClientError(t *testing.T) {
	stub := &stubRoundTripper{
		handler: func(call int) (*http.Response, error) {
			return errorResponse(http.StatusBadRequest), nil
		},
	}

	rt := NewRetryRoundTripper(stub, []time.Duration{0, 0, 0})
	resp, err := rt.RoundTrip(newTestRequest(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if stub.calls != 1 {
		t.Errorf("expected 1 attempt, got %d", stub.calls)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRetryRoundTripperExhaustsRetries(t *testing.T) {
	stub := &stubRoundTripper{
		handler: func(call int) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	rt := NewRetryRoundTripper(stub, []time.Duration{0, 0, 0})
	_, err := rt.RoundTrip(newTestRequest(t))
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if stub.calls != 4 {
		t.Errorf("expected 4 attempts, got %d", stub.calls)
	}
}
