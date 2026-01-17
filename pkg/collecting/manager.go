package collecting

import (
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/metrics"
	"InferenceProfiler/pkg/probing"
	"log"
	"sync"
)

type Manager struct {
	collectors []Collector
	static     metrics.Record
	mu         sync.RWMutex
}

func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		collectors: make([]Collector, 0),
	}

	// Initialize VM collectors
	if cfg.EnableVM {
		m.collectors = append(m.collectors,
			NewCPU(),
			NewMemory(),
			NewDisk(),
			NewNetwork(),
		)
	}

	// Initialize container collector
	if cfg.EnableContainer {
		if c := NewContainer(); c != nil {
			m.collectors = append(m.collectors, c)
		}
	}

	// Initialize process collector
	if cfg.EnableProcess {
		m.collectors = append(m.collectors, NewProcess())
	}

	// Initialize NVIDIA collector
	if cfg.EnableNvidia {
		if n := NewNvidia(cfg.CollectGPUProcesses); n != nil {
			m.collectors = append(m.collectors, n)
		}
	}

	// Initialize vLLM collector
	if cfg.EnableVLLM {
		m.collectors = append(m.collectors, NewVLLM())
	}

	log.Printf("Initialized %d collectors", len(m.collectors))
	return m
}

func (m *Manager) CollectStatic(base *metrics.BaseStatic) {
	staticMetrics := make([]any, 0, len(m.collectors)+1)
	staticMetrics = append(staticMetrics, base)

	for _, collector := range m.collectors {
		if data := collector.CollectStatic(); data != nil {
			staticMetrics = append(staticMetrics, data)
		}
	}

	m.mu.Lock()
	m.static = metrics.Merge(staticMetrics...)
	m.mu.Unlock()
}

func (m *Manager) CollectDynamic(base *metrics.BaseDynamic) metrics.Record {
	base.Timestamp = probing.GetTimestamp()
	dynamicMetrics := make([]any, 0, len(m.collectors)+1)
	dynamicMetrics = append(dynamicMetrics, base)

	for _, collector := range m.collectors {
		if data := collector.CollectDynamic(); data != nil {
			dynamicMetrics = append(dynamicMetrics, data)
		}
	}

	m.mu.RLock()
	static := m.static
	m.mu.RUnlock()

	// Merge static and dynamic metrics
	allMetrics := make([]any, 0, len(dynamicMetrics)+1)
	allMetrics = append(allMetrics, static)
	allMetrics = append(allMetrics, dynamicMetrics...)

	return metrics.Merge(allMetrics...)
}

func (m *Manager) Close() {
	for _, collector := range m.collectors {
		if err := collector.Close(); err != nil {
			log.Printf("Error closing collector %s: %v", collector.Name(), err)
		}
	}
}

func (m *Manager) GetCollectorNames() []string {
	names := make([]string, len(m.collectors))
	for i, c := range m.collectors {
		names[i] = c.Name()
	}
	return names
}
