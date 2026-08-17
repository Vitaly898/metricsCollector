package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	models "github.com/Vitaly898/metricsCollector/internal/model"
	"log"
	"net/http"
)

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
		log.Printf("Cannor marshal metrics: %v", err)
		return false
	}

	compressed, err := compress(body)
	if err != nil {
		log.Printf("Cannor compress metrics: %v", err)
		return false
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressed))
	if err != nil {
		log.Printf("Cannor create request: %v", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Cannor send request: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Response status is not 200: %v", resp.StatusCode)
		return false
	}

	return true
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s *Sender) Send(gauges map[string]float64, pollCount int64) {
	delta := pollCount - s.lastPollCount
	if delta < 0 {
		// Счетчик сбросился (например, агент перезапустился).
		// Отправляем абсолютное накопленное значение с нового старта.
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
