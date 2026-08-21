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
		client:  &http.Client{},
	}
}

func (s *Sender) sendMetric(m models.Metrics) bool {
	url := fmt.Sprintf("%s/update", s.baseUrl)

	body, err := json.Marshal(m)
	if err != nil {
		logger.Errorw("Cannot marshal metric", "error", err)
		return false
	}

	compressed, err := compress.Compress(body)
	if err != nil {
		logger.Errorw("Cannot compress metric", "error", err)
		return false
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressed))
	if err != nil {
		logger.Errorw("Cannot create request", "error", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Errorw("Cannot send request", "error", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Errorw("Response status is not 200", "status", resp.StatusCode)
		return false
	}

	return true
}

func (s *Sender) Send(gauges map[string]float64, pollCount int64) {
	delta := pollCount - s.lastPollCount
	if delta < 0 {

		delta = pollCount
	}

	allOk := true
	for metric, val := range gauges {
		v := val
		if !s.sendMetric(models.Metrics{
			ID:    metric,
			MType: models.Gauge,
			Value: &v,
		}) {
			allOk = false
		}
	}
	if !s.sendMetric(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &delta,
	}) {
		allOk = false
	}
	if allOk {
		s.lastPollCount += delta
	}
}
