package agent

import (
	"time"
)

type Agent struct {
	collector    *Collector
	sender       *Sender
	pollInterval time.Duration
	sendInterval time.Duration
}

func NewAgent(c *Collector, s *Sender, pollInterval, sendInterval time.Duration) *Agent {
	return &Agent{
		collector:    c,
		sender:       s,
		pollInterval: pollInterval,
		sendInterval: sendInterval,
	}
}

func (a *Agent) Run() {
	polls := 0
	for {
		gauges, pollCount := a.collector.Collect()
		polls++

		if time.Duration(polls)*a.pollInterval >= a.sendInterval {
			a.sender.Send(gauges, pollCount)
			polls = 0
		}

		time.Sleep(a.pollInterval)
	}

}
