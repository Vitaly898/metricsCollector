package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	models "github.com/Vitaly898/metricsCollector/internal/model"
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

func (s *Sender) sendMetric(metricType string, metric string, val string) bool {
	url := fmt.Sprintf("%s/update/%s/%s/%s", s.baseUrl, metricType, metric, val)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("Cannot create request %v", err)
		return false
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Server problem %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Server response status code is not 200 %d", resp.StatusCode)
		return false
	}

	return true
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
		if !s.sendMetric(models.Gauge, metric, strconv.FormatFloat(val, 'f', -1, 64)) {
			allOk = false
		}
	}

	if !s.sendMetric(models.Counter, "PollCount", strconv.FormatInt(delta, 10)) {
		allOk = false
	}

	if allOk {
		s.lastPollCount += delta
	}
}
