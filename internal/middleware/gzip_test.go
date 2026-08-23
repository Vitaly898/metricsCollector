package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddlewareCompressesJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux := http.NewServeMux()
	mux.Handle("/", GzipMiddleware(handler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding gzip, got %q", res.Header.Get("Content-Encoding"))
	}

	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatalf("cannot create gzip reader: %v", err)
	}
	defer gr.Close()

	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("cannot read decompressed body: %v", err)
	}

	if string(body) != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", string(body), `{"status":"ok"}`)
	}
}

func TestGzipMiddlewareDoesNotCompressWithoutAcceptEncoding(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.Header.Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding, got %q", res.Header.Get("Content-Encoding"))
	}

	body, _ := io.ReadAll(res.Body)
	if string(body) != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", string(body), `{"status":"ok"}`)
	}
}

func TestGzipMiddlewareDoesNotCompressPlainText(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.Header.Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding for text/plain, got %q", res.Header.Get("Content-Encoding"))
	}
}

func TestGzipMiddlewareDecompressesRequest(t *testing.T) {
	var received string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.WriteHeader(http.StatusOK)
	})

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte(`{"id":"x"}`))
	gw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)

	if received != `{"id":"x"}` {
		t.Errorf("received body = %q, want %q", received, `{"id":"x"}`)
	}
}
