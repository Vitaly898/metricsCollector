package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vitaly898/metricsCollector/internal/hash"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("echo:" + string(body)))
}

func TestHashMiddleware(t *testing.T) {
	const key = "secret"

	t.Run("empty key test", func(t *testing.T) {
		h := HashMiddleware("")(http.HandlerFunc(okHandler))
		req := httptest.NewRequest(http.MethodPost, "/updates", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rec.Code, http.StatusOK)
		}
	})
	t.Run("валидная подпись — 200 и подписанный ответ", func(t *testing.T) {
		h := HashMiddleware(key)(http.HandlerFunc(okHandler))
		body := []byte(`[{"id":"a","type":"gauge","value":1}]`)
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
		req.Header.Set(hashHeader, hash.Compute(body, key)) // агентская сторона
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("got %d, want 200", rec.Code)
		}
		want := hash.Compute([]byte("echo:"+string(body)), key)
		if got := rec.Header().Get(hashHeader); got != want {
			t.Errorf("response hash = %q, want %q", got, want)
		}
	})

	t.Run("невалидная подпись — 400, handler не вызван", func(t *testing.T) {
		called := false
		h := HashMiddleware(key)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))
		req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
		req.Header.Set(hashHeader, "deadbeef")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", rec.Code)
		}
		if called {
			t.Error("handler must not be called on invalid hash")
		}
	})

	t.Run("заголовок отсутствует — 400", func(t *testing.T) {
		h := HashMiddleware(key)(http.HandlerFunc(okHandler))
		req := httptest.NewRequest(http.MethodPost, "/updates/", nil) // без HashSHA256
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", rec.Code)
		}
	})

}
