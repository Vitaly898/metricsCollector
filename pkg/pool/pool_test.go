package pool_test

import (
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	models "github.com/Vitaly898/metricsCollector/internal/model"
	"github.com/Vitaly898/metricsCollector/pkg/pool"
)

type object struct {
	text    string
	values  []int
	labels  map[string]int
	onReset func(*object)
}

func (o *object) Reset() {
	o.text = ""
	o.values = o.values[:0]
	clear(o.labels)
	if o.onReset != nil {
		o.onReset(o)
	}
}

func newObject() *object {
	return &object{values: make([]int, 0, 8), labels: make(map[string]int)}
}

func TestGetUsesFactoryWhenEmpty(t *testing.T) {
	var created int
	p := pool.New(func() *object {
		created++
		o := newObject()
		o.text = strconv.Itoa(created)
		return o
	})
	if created != 0 {
		t.Fatal("constructor called the factory before Get")
	}
	first, second := p.Get(), p.Get()
	if created != 2 || first == second || first.text != "1" || second.text != "2" {
		t.Fatal("Get did not create separate objects through the factory")
	}
}

func TestPutResetsState(t *testing.T) {
	p := pool.New(newObject)
	o := p.Get()
	o.text = "dirty"
	o.values = append(o.values, 1, 2, 3)
	o.labels["key"] = 1
	backing, capacity := &o.values[0], cap(o.values)
	labels := o.labels
	var resets int
	o.onReset = func(reset *object) {
		resets++
		if reset.text != "" || len(reset.values) != 0 || len(reset.labels) != 0 {
			t.Error("object state was not reset")
		}
		if cap(reset.values) != capacity || &reset.values[:capacity][0] != backing {
			t.Error("reset replaced slice storage")
		}
		if len(labels) != 0 {
			t.Error("reset did not clear the original map")
		}
	}
	p.Put(o)
	if resets != 1 {
		t.Fatalf("Reset calls = %d, want 1 before Put returns", resets)
	}
}

type gatedObject struct {
	started chan struct{}
	release chan struct{}
}

func (o *gatedObject) Reset() {
	if o.started != nil {
		close(o.started)
		<-o.release
	}
}

func TestPutPublishesAfterResetFinishes(t *testing.T) {
	p := pool.New(func() *gatedObject { return &gatedObject{} })
	o := &gatedObject{started: make(chan struct{}), release: make(chan struct{})}
	finished := make(chan struct{})
	defer func() {
		close(o.release)
		<-finished
	}()
	go func() {
		p.Put(o)
		close(finished)
	}()
	<-o.started
	if got := p.Get(); got == o {
		t.Error("Get returned an object while its Reset was still running")
	}
	select {
	case <-finished:
		t.Error("Put returned before Reset finished")
	default:
	}
}

func TestPoolWithGeneratedMetricsReset(t *testing.T) {
	p := pool.New(func() *models.Metrics {
		return &models.Metrics{Delta: new(int64), Value: new(float64)}
	})
	metric := p.Get()
	metric.ID, metric.MType, metric.Hash = "Alloc", models.Gauge, "hash"
	*metric.Delta, *metric.Value = 12, 3.5
	p.Put(metric)
	clean := p.Get()
	if clean.ID != "" || clean.MType != "" || clean.Hash != "" ||
		clean.Delta == nil || clean.Value == nil || *clean.Delta != 0 || *clean.Value != 0 {
		t.Fatalf("unexpected metric state: %+v", clean)
	}
	p.Put(clean)
}

func TestGetAfterGC(t *testing.T) {
	p := pool.New(newObject)
	o := p.Get()
	o.text = "dirty"
	p.Put(o)
	runtime.GC()
	runtime.GC()
	clean := p.Get()
	if clean == nil || clean.text != "" || clean.values == nil || clean.labels == nil {
		t.Fatal("Get did not return a ready object after GC")
	}
	p.Put(clean)
}

type labels map[string]int

func (l labels) Reset() { clear(l) }

func TestNonComparableResetter(t *testing.T) {
	p := pool.New(func() labels { return make(labels) })
	value := p.Get()
	value["key"] = 1
	p.Put(value)
	clean := p.Get()
	if clean == nil || len(clean) != 0 {
		t.Fatal("non-comparable value was not reset")
	}
	p.Put(clean)
}

func TestConcurrentGetPut(t *testing.T) {
	const workers, iterations = 32, 200
	var resets atomic.Int64
	p := pool.New(func() *object {
		o := newObject()
		o.onReset = func(*object) { resets.Add(1) }
		return o
	})
	var wg sync.WaitGroup
	for worker := range workers {
		wg.Go(func() {
			name := strconv.Itoa(worker)
			for iteration := range iterations {
				o := p.Get()
				if o.text != "" || len(o.values) != 0 || len(o.labels) != 0 {
					t.Error("Get returned dirty state")
					return
				}
				o.text = name
				o.values = append(o.values, iteration)
				o.labels["worker"] = worker
				runtime.Gosched()
				if o.text != name || len(o.values) != 1 || o.values[0] != iteration || o.labels["worker"] != worker {
					t.Error("object was shared with another borrower")
					return
				}
				p.Put(o)
			}
		})
	}
	wg.Wait()
	if got, want := resets.Load(), int64(workers*iterations); got != want {
		t.Fatalf("Reset calls = %d, want %d", got, want)
	}
}
