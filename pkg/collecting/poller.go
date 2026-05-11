package collecting

import (
	"InferenceProfiler/pkg/collecting/base"
	"context"
	"fmt"
	"sync"
	"time"
)

type poller struct {
	collector base.Collector
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	mu      sync.RWMutex
	cached  any
	cycles  int64
	totalNs int64
	minNs   int64
	maxNs   int64
}

func (p *poller) timedPoll(ctx context.Context) {
	start := time.Now()
	data := p.collector.Poll(ctx)
	elapsed := time.Since(start).Nanoseconds()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.cached = data
	p.cycles++
	p.totalNs += elapsed
	if elapsed < p.minNs || p.cycles == 1 {
		p.minNs = elapsed
	}
	if elapsed > p.maxNs {
		p.maxNs = elapsed
	}
}

func (p *poller) latest() any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cached
}

func (p *poller) startLoop(ctx context.Context, interval time.Duration) {
	ctx, p.cancel = context.WithCancel(ctx)
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()

		p.timedPoll(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.timedPoll(ctx)
			}
		}
	}()
}

func (p *poller) stop() {
	if p.cancel != nil {
		p.cancel()
		p.wg.Wait()
	}
}

func (p *poller) pollStats() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.cycles == 0 {
		return "no polls"
	}
	avgUs := float64(p.totalNs) / float64(p.cycles) / 1000.0
	minUs := float64(p.minNs) / 1000.0
	maxUs := float64(p.maxNs) / 1000.0
	return fmt.Sprintf("cycles=%8d  avg=%10.1fµs  min=%10.1fµs  max=%10.1fµs", p.cycles, avgUs, minUs, maxUs)
}
