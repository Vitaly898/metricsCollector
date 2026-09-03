package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubPinger struct {
	err error
}

func (s stubPinger) Ping(_ context.Context) error {
	return s.err
}

func TestPingHandlerOK(t *testing.T) {
	h := NewPingHandler(stubPinger{})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPingHandlerError(t *testing.T) {
	h := NewPingHandler(stubPinger{err: errors.New("connection refused")})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if body := rec.Body.String(); body != http.StatusText(http.StatusInternalServerError)+"\n" {
		t.Errorf("body = %q, want neutral %q", body, http.StatusText(http.StatusInternalServerError))
	}
}
