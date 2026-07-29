package agent

import (
	"time"
)

type Agent struct {
	collector    *Collector
	sender       *Sender
	pollInterval time.Duration
	sendInterval time.Duration
	stop         chan struct{}
}

func NewAgent(c *Collector, s *Sender, pollInterval, sendInterval time.Duration) *Agent {
	return &Agent{
		collector:    c,
		sender:       s,
		pollInterval: pollInterval,
		sendInterval: sendInterval,
		stop:         make(chan struct{}),
	}
}

func (a *Agent) Run() {
	polls := 0
	for {
		select {
		case <-a.stop:
			return
		default:
		}

		gauges, pollCount := a.collector.Collect()
		polls++

		if time.Duration(polls)*a.pollInterval >= a.sendInterval {
			a.sender.Send(gauges, pollCount)
			polls = 0
		}

		select {
		case <-a.stop:
			return
		case <-time.After(a.pollInterval):
		}
	}
}

func (a *Agent) Stop() {
	close(a.stop)
}
