package agent

import (
	"context"
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
	done         chan struct{}
	wg           sync.WaitGroup

	mu     sync.Mutex
	cancel context.CancelFunc
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
		jobs:         make(chan sendJob, rateLimit),
		done:         make(chan struct{}),
	}
}

func (a *Agent) collectLoop(ctx context.Context, collect func() map[string]float64, countPoll bool) {
	defer a.wg.Done()
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.storage.UpdateGauges(collect())
			if countPoll {
				a.storage.AddPollCount()
			}
		}
	}
}

func (a *Agent) worker(ctx context.Context) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-a.jobs:
			if !ok {
				return
			}
			a.sender.SendBatch(job.metrics, job.pollCount)
		}
	}
}

func (a *Agent) Run(ctx context.Context) {
	defer close(a.done)

	ctx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mu.Unlock()

	a.wg.Add(2)
	go a.collectLoop(ctx, a.collector.Collect, true)
	go a.collectLoop(ctx, a.collector.CollectSystem, false)

	a.wg.Add(a.rateLimit)
	for i := 0; i < a.rateLimit; i++ {
		go a.worker(ctx)
	}

	ticker := time.NewTicker(a.sendInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(a.jobs)
			a.wg.Wait()
			return
		case <-ticker.C:
			gauges, pollCount := a.storage.Snapshot()
			job := sendJob{metrics: a.sender.BuildBatch(gauges, pollCount), pollCount: pollCount}
			select {
			case a.jobs <- job:
			case <-ctx.Done():
			}
		}
	}
}

func (a *Agent) Stop() {
	a.mu.Lock()
	cancel := a.cancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	<-a.done
}
