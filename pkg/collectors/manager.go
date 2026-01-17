package collectors

import (
	"InferenceProfiler/pkg/collectors/vm"
	"log"
	"sync"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

// Manager manages all collectors and coordinates metric collection.
type Manager struct {
	collectors []types.Collector
	static     types.Record
	mu         sync.RWMutex
}

// NewManager creates a new collector manager based on configuration.
func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		collectors: make([]types.Collector, 0),
	}

	// Initialize VM collectors
	if cfg.EnableVM {
		m.collectors = append(m.collectors,
			vm.New(),
			vm.New(),
			vm.New(),
			vm.New(),
		)
	}

	// Initialize container collector
	if cfg.EnableContainer {
		if c := New(); c != nil {
			m.collectors = append(m.collectors, c)
		}
	}

	// Initialize process collector
	if cfg.EnableProcess {
		m.collectors = append(m.collectors, New())
	}

	// Initialize NVIDIA collector
	if cfg.EnableNvidia {
		if n := New(cfg.CollectGPUProcesses); n != nil {
			m.collectors = append(m.collectors, n)
		}
	}

	// Initialize vLLM collector
	if cfg.EnableVLLM {
		m.collectors = append(m.collectors, New())
	}

	log.Printf("Initialized %d collectors", len(m.collectors))
	return m
}

// CollectStatic collects static metrics from all collectors.
func (m *Manager) CollectStatic(base *BaseStatic) {
	records := make([]types.Record, 0, len(m.collectors)+1)
	records = append(records, base.ToRecord())

	for _, collector := range m.collectors {
		if data := collector.CollectStatic(); data != nil {
			records = append(records, data)
		}
	}

	m.mu.Lock()
	m.static = types.MergeRecords(records...)
	m.mu.Unlock()
}

// CollectDynamic collects dynamic metrics from all collectors.
func (m *Manager) CollectDynamic(base *BaseDynamic) types.Record {
	base.Timestamp = probing.GetTimestamp()

	records := make([]types.Record, 0, len(m.collectors)+2)

	m.mu.RLock()
	records = append(records, m.static)
	m.mu.RUnlock()

	records = append(records, base.ToRecord())

	for _, collector := range m.collectors {
		if data := collector.CollectDynamic(); data != nil {
			records = append(records, data)
		}
	}

	return types.MergeRecords(records...)
}

// Close releases resources from all collectors.
func (m *Manager) Close() {
	for _, collector := range m.collectors {
		if err := collector.Close(); err != nil {
			log.Printf("Error closing collector %s: %v", collector.Name(), err)
		}
	}
}

// GetCollectorNames returns the names of all active collectors.
func (m *Manager) GetCollectorNames() []string {
	names := make([]string, len(m.collectors))
	for i, c := range m.collectors {
		names[i] = c.Name()
	}
	return names
}
