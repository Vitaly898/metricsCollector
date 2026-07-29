package agent

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestSenderSendsPollCountDelta(t *testing.T) {
	var totalPollCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Парсим /update/counter/PollCount/{value} и суммируем на стороне "сервера".
		path := strings.TrimPrefix(r.URL.Path, "/update/")
		parts := strings.Split(path, "/")
		if len(parts) == 3 && parts[0] == "counter" && parts[1] == "PollCount" {
			if v, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				totalPollCount += v
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSender(server.URL)

	sender.Send(map[string]float64{"Alloc": 1}, 3)
	if totalPollCount != 3 {
		t.Errorf("after first send totalPollCount = %d, want 3", totalPollCount)
	}

	sender.Send(map[string]float64{"Alloc": 2}, 5)
	if totalPollCount != 5 {
		t.Errorf("after second send totalPollCount = %d, want 5", totalPollCount)
	}

	// Имитируем перезапуск агента: новый Sender.
	newSender := NewSender(server.URL)
	newSender.Send(map[string]float64{"Alloc": 3}, 2)
	if totalPollCount != 7 {
		t.Errorf("after restart send totalPollCount = %d, want 7", totalPollCount)
	}
}
