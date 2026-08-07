package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData хранит сведения об ответе: код статуса и размер тела.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter оборачивает http.ResponseWriter, чтобы перехватывать
// код статуса и размер записанного тела ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

// Logger возвращает logger, логирующий сведения о запросах и ответах:
// URI, метод, длительность выполнения, код статуса и размер ответа.
func Logger(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData{status: http.StatusOK},
			}

			next.ServeHTTP(lw, r)

			logger.Infow("request handled",
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", time.Since(start),
				"status", lw.responseData.status,
				"size", lw.responseData.size,
			)
		})
	}
}
