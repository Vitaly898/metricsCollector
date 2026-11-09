package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Vitaly898/metricsCollector/internal/hash"

	"github.com/Vitaly898/metricsCollector/internal/compress"
	models "github.com/Vitaly898/metricsCollector/internal/model"
)

// Sender отправляет пакеты метрик и хранит последнее подтверждённое число опросов.
type Sender struct {
	baseURL       string
	client        *http.Client
	mu            sync.Mutex
	lastPollCount int64
	key           string
}

// NewSender создаёт отправитель на POST baseURL/updates/ с gzip и повторами запросов.
// Непустой key включает подпись JSON-тела в заголовке HashSHA256.
func NewSender(baseURL string, key string) *Sender {
	return &Sender{
		baseURL: baseURL,
		key:     key,
		client: &http.Client{
			Transport: NewRetryRoundTripper(nil, defaultRetryIntervals),
		},
	}
}

func (s *Sender) trySendMetrics(metrics []models.Metrics) (int, error) {
	url := fmt.Sprintf("%s/updates/", s.baseURL)

	body, err := json.Marshal(metrics)
	if err != nil {
		return 0, err
	}

	compressed, err := compress.Compress(body)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressed))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	if s.key != "" {
		req.Header.Set("HashSHA256", hash.Compute(body, s.key))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (s *Sender) sendMetrics(metrics []models.Metrics) bool {
	if len(metrics) == 0 {
		return true
	}

	// Ретраи выполняются прозрачно внутри http.Client (см. retryRoundTripper).
	status, err := s.trySendMetrics(metrics)
	return err == nil && status == http.StatusOK
}

// BuildBatch формирует независимый пакет gauge и приращение counter PollCount
// относительно последней успешной отправки. При сбросе счётчика приращением
// становится pollCount. Порядок gauge-метрик в пакете не определён.
func (s *Sender) BuildBatch(gauges map[string]float64, pollCount int64) []models.Metrics {
	s.mu.Lock()
	delta := pollCount - s.lastPollCount
	if delta < 0 {
		delta = pollCount
	}
	s.mu.Unlock()

	metrics := make([]models.Metrics, 0, len(gauges)+1)
	// Each metric needs a stable pointer; one backing array replaces a heap object per gauge.
	values := make([]float64, len(gauges))
	i := 0
	for metric, val := range gauges {
		values[i] = val
		metrics = append(metrics, models.Metrics{
			ID:    metric,
			MType: models.Gauge,
			Value: &values[i],
		})
		i++
	}

	d := delta
	metrics = append(metrics, models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &d,
	})
	return metrics
}

// SendBatch отправляет готовый пакет и при успехе запоминает pollCount.
// Ошибка доставки не меняет подтверждённый счётчик. Пустой пакет считается успешным.
func (s *Sender) SendBatch(metrics []models.Metrics, pollCount int64) {
	if s.sendMetrics(metrics) {
		s.mu.Lock()
		s.lastPollCount = pollCount
		s.mu.Unlock()
	}
}

// Send формирует пакет через BuildBatch и отправляет его через SendBatch.
func (s *Sender) Send(gauges map[string]float64, pollCount int64) {
	s.SendBatch(s.BuildBatch(gauges, pollCount), pollCount)
}
