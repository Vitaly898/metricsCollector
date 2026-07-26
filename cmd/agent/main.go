package main

import (
	"flag"
	"time"

	"github.com/Vitaly898/metricsCollector/internal/agent"
)

func main() {
	addr := flag.String("a","localhost:8080","адрес сервера")
	reportInterval:= flag.Int("r",10,"частота отправки метрик в секундах")
	pollInterval := flag.Int("p",2,"частота опроса метрик в секундах")
	collector := agent.NewCollector()
	sender := agent.NewSender("http://" + *addr)
	a := agent.NewAgent(collector, sender, time.Duration(*pollInterval)*time.Second,time.Duration(*reportInterval)*time.Second )
	a.Run()

}