package agent

import (
	"sync"
	"time"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

type sendJob struct {
	metrics   []models.Metrics
	pollCount int64
}

type Agent struct {
	collector    *Collector
	sender       *Sender
	storage      *Storage
	pollInterval time.Duration
	sendInterval time.Duration
	rateLimit    int
	jobs         chan sendJob
	stop         chan struct{}
	done         chan struct{}
	wg           sync.WaitGroup
}

func NewAgent(c *Collector, s *Sender, pollInterval, sendInterval time.Duration, rateLimit int) *Agent {
	if rateLimit < 1 {
		rateLimit = 1
	}
	return &Agent{
		collector:    c,
		sender:       s,
		storage:      NewStorage(),
		pollInterval: pollInterval,
		sendInterval: sendInterval,
		rateLimit:    rateLimit,
		jobs:         make(chan sendJob),
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
	}
}

func (a *Agent) collectLoop(collect func() map[string]float64, countPoll bool) {
	defer a.wg.Done()
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-a.stop:
			return
		case <-ticker.C:
			a.storage.UpdateGauges(collect())
			if countPoll {
				a.storage.AddPollCount()
			}
		}
	}
}

func (a *Agent) worker() {
	defer a.wg.Done()
	for job := range a.jobs {
		a.sender.SendBatch(job.metrics, job.pollCount)
	}
}

func (a *Agent) Run() {
	defer close(a.done)

	a.wg.Add(2)
	go a.collectLoop(a.collector.Collect, true)
	go a.collectLoop(a.collector.CollectSystem, false)

	a.wg.Add(a.rateLimit)
	for i := 0; i < a.rateLimit; i++ {
		go a.worker()
	}

	ticker := time.NewTicker(a.sendInterval)
	defer ticker.Stop()
	for {
		select {
		case <-a.stop:
			close(a.jobs)
			a.wg.Wait()
			return
		case <-ticker.C:
			gauges, pollCount := a.storage.Snapshot()
			a.jobs <- sendJob{metrics: a.sender.BuildBatch(gauges, pollCount), pollCount: pollCount}
		}
	}
}

func (a *Agent) Stop() {
	close(a.stop)
	<-a.done
}
