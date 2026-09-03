package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Vitaly898/metricsCollector/internal/compress"
	models "github.com/Vitaly898/metricsCollector/internal/model"
)

var logger = zap.Must(zap.NewProduction()).Sugar()

type Sender struct {
	baseUrl       string
	client        *http.Client
	lastPollCount int64
}

func NewSender(baseUrl string) *Sender {
	return &Sender{
		baseUrl: baseUrl,
		client: &http.Client{
			Transport: NewRetryRoundTripper(nil, defaultRetryIntervals),
		},
	}
}

func (s *Sender) trySendMetrics(metrics []models.Metrics) (int, error) {
	url := fmt.Sprintf("%s/updates/", s.baseUrl)

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

func (s *Sender) Send(gauges map[string]float64, pollCount int64) {
	delta := pollCount - s.lastPollCount
	if delta < 0 {
		delta = pollCount
	}

	metrics := make([]models.Metrics, 0, len(gauges)+1)
	for metric, val := range gauges {
		v := val
		metrics = append(metrics, models.Metrics{
			ID:    metric,
			MType: models.Gauge,
			Value: &v,
		})
	}

	d := delta
	metrics = append(metrics, models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &d,
	})

	if s.sendMetrics(metrics) {
		s.lastPollCount += delta
	}
}
