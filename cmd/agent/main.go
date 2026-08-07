package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/Vitaly898/metricsCollector/internal/agent"
)

func main() {
	addr := flag.String("a", "localhost:8080", "адрес сервера")
	reportInterval := flag.Int("r", 10, "частота отправки метрик в секундах")
	pollInterval := flag.Int("p", 2, "частота опроса метрик в секундах")
	flag.Parse()
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		*addr = envAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if v, err := strconv.Atoi(envReportInterval); err == nil {
			*reportInterval = v
		}
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if v, err := strconv.Atoi(envPollInterval); err == nil {
			*pollInterval = v
		}
	}
	collector := agent.NewCollector()
	sender := agent.NewSender("https://" + *addr)
	a := agent.NewAgent(collector, sender, time.Duration(*pollInterval)*time.Second, time.Duration(*reportInterval)*time.Second)
	a.Run()

}
