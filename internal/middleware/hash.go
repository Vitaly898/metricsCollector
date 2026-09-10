package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Vitaly898/metricsCollector/internal/hash"
)

const hashHeader = "HashSHA256"

type signWriter struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
	wrote  bool
}

func (w *signWriter) WriteHeader(status int) {
	if w.wrote {
		return
	}
	w.wrote = true
	w.status = status
}

// Write накапливает тело в буфер, ничего не отправляя клиенту.
// БЕЗ этого метода встроенный http.ResponseWriter.Write отправлял бы тело
// напрямую, мимо буфера, и подпись считалась бы от пустого тела.
func (w *signWriter) Write(p []byte) (int, error) {
	if !w.wrote { // handler написал тело без явного WriteHeader — по соглашению это 200 OK
		w.WriteHeader(http.StatusOK)
	}
	return w.buf.Write(p)
}

func (w *signWriter) Finish(key string) {
	if !w.wrote {
		w.status = http.StatusOK
	}
	w.Header().Set(hashHeader, hash.Compute(w.buf.Bytes(), key))
	w.ResponseWriter.WriteHeader(w.status)
	_, _ = w.ResponseWriter.Write(w.buf.Bytes())
}

func HashMiddleware(key string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			r.Body = io.NopCloser(bytes.NewReader(body))
			received := r.Header.Get(hashHeader)
			expected := hash.Compute(body, key)
			if received != expected {
				http.Error(w, "invalid signature", http.StatusBadRequest)
				return
			}
			sw := &signWriter{ResponseWriter: w}
			defer sw.Finish(key)
			next.ServeHTTP(sw, r)
		})
	}
}
