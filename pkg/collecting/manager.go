package collecting

import (
	"InferenceProfiler/pkg/collecting/base"
	"context"
	"fmt"
	"log"
	"time"

	"InferenceProfiler/pkg/collecting/container"
	"InferenceProfiler/pkg/collecting/nvidia"
	"InferenceProfiler/pkg/collecting/process"
	"InferenceProfiler/pkg/collecting/vllm"
	"InferenceProfiler/pkg/collecting/vm"
	"InferenceProfiler/pkg/utils"
)

type InitResult struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

type Manager struct {
	cfg         *utils.Config
	pollers     []*poller
	initResults []InitResult
}

func NewManager(cfg *utils.Config) *Manager {
	m := &Manager{cfg: cfg}

	m.tryInit(vm.New(), cfg.DisableVM, cfg)
	m.tryInit(container.New(), cfg.DisableContainer, cfg)
	m.tryInit(process.New(), cfg.DisableProcess, cfg)
	m.tryInit(nvidia.New(), cfg.DisableNvidia, cfg)
	m.tryInit(vllm.New(), cfg.DisableVLLM, cfg)

	log.Printf("manager: initialized %d collectors", len(m.pollers))
	return m
}

func (m *Manager) tryInit(c base.Collector, disabled bool, cfg *utils.Config) {
	r := InitResult{Name: c.Name(), Enabled: !disabled}

	if disabled {
		m.initResults = append(m.initResults, r)
		return
	}

	if err := c.Init(cfg); err != nil {
		r.Error = err.Error()
		utils.Debugf("manager: %s init failed: %v", c.Name(), err)
		m.initResults = append(m.initResults, r)
		return
	}

	r.Available = true
	m.initResults = append(m.initResults, r)
	m.pollers = append(m.pollers, &poller{collector: c})
}

func (m *Manager) writeStatic(w base.Writer) {
	for _, p := range m.pollers {
		if s := p.collector.Static(); s != nil {
			w.Static(p.collector.Name(), s)
		}
	}
	w.Flush()
}

func (m *Manager) Snapshot(ctx context.Context, w base.Writer) int {
	m.writeStatic(w)

	type result struct {
		name string
		data any
	}

	ch := make(chan result, len(m.pollers))
	for _, p := range m.pollers {
		go func() {
			ch <- result{name: p.collector.Name(), data: p.collector.Poll(ctx)}
		}()
	}

	for range m.pollers {
		if r := <-ch; r.data != nil {
			w.Dynamic(r.name, r.data)
		}
	}

	w.Flush()
	return len(m.pollers)
}

func (m *Manager) Continuous(ctx context.Context, w base.Writer) (int, error) {
	interval := time.Duration(m.cfg.Interval) * time.Millisecond

	for _, p := range m.pollers {
		p.startLoop(ctx, interval)
	}
	defer func() {
		for _, p := range m.pollers {
			p.stop()
		}
	}()

	m.writeStatic(w)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	count := 0
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			elapsed := time.Since(start)
			log.Printf("manager: collected %d records in %v", count, elapsed)
			for _, p := range m.pollers {
				log.Printf("%-12s %s", p.collector.Name()+":", p.pollStats())
			}
			return count, nil

		case <-ticker.C:
			t := utils.DebugTimer()
			for _, p := range m.pollers {
				if data := p.latest(); data != nil {
					w.Dynamic(p.collector.Name(), data)
				}
			}
			if err := w.Flush(); err != nil {
				log.Printf("manager: flush error: %v", err)
			}
			count++
			utils.DebugDuration("manager", fmt.Sprintf("tick #%d", count), t)
		}
	}
}

func (m *Manager) Close() error {
	for _, p := range m.pollers {
		p.stop()
		p.collector.Close()
	}
	return nil
}

func (m *Manager) InitResults() []InitResult { return m.initResults }
func (m *Manager) Config() *utils.Config     { return m.cfg }

func (m *Manager) StaticData() map[string]any {
	out := make(map[string]any, len(m.pollers))
	for _, p := range m.pollers {
		if s := p.collector.Static(); s != nil {
			out[p.collector.Name()] = s
		}
	}
	return out
}

func (m *Manager) SnapshotTick() map[string]any {
	type result struct {
		name string
		data any
	}
	ctx := context.Background()
	ch := make(chan result, len(m.pollers))
	for _, p := range m.pollers {
		go func() {
			ch <- result{name: p.collector.Name(), data: p.collector.Poll(ctx)}
		}()
	}
	out := make(map[string]any, len(m.pollers))
	for range m.pollers {
		if r := <-ch; r.data != nil {
			out[r.name] = r.data
		}
	}
	return out
}
