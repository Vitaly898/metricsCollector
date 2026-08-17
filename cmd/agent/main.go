package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/Vitaly898/metricsCollector/internal/agent"
)

func main() {
	addr := flag.String("a", "localhost:8080", "server address")
	reportInterval := flag.Int("r", 10, "metrics sending interval in seconds")
	pollInterval := flag.Int("p", 2, "metrics polling interval in seconds")
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
	sender := agent.NewSender("http://" + *addr)
	a := agent.NewAgent(collector, sender, time.Duration(*pollInterval)*time.Second, time.Duration(*reportInterval)*time.Second)
	a.Run()

}
