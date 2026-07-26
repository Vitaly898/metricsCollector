package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type Sender struct {
	baseUrl string
	client  *http.Client
}

func NewSender(baseUrl string) *Sender {
	return &Sender{
		baseUrl: baseUrl,
		client:  &http.Client{},
	}
}

func (s *Sender) SendMetric(metricType string, metric string, val string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", s.baseUrl, metricType, metric, val)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("Cannot create request")
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Server problem")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Server response status code is not 200")
		return
	}
}

func (s *Sender) Send(gauges map[string]float64, pollCount int64) {
	for metric, val := range gauges {
		s.SendMetric(models.Gauge, metric, strconv.FormatFloat(val, 'f', -1, 64))
	}
	s.SendMetric(models.Counter, "PollCount", strconv.FormatInt(pollCount, 10))
}
