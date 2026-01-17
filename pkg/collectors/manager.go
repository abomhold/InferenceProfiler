package collectors

import (
	"encoding/json"
	"log"
	"sync"

	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/formatting"
	"InferenceProfiler/pkg/probing"
)

type Manager struct {
	collectors []Collector
	static     *StaticMetrics
	staticJSON formatting.Record
	mu         sync.RWMutex
}

func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		collectors: make([]Collector, 0, 8),
	}

	if cfg.EnableVM {
		m.collectors = append(m.collectors,
			NewCPUCollector(),
			NewMemoryCollector(),
			NewDiskCollector(),
			NewNetworkCollector(),
		)
	}

	if cfg.EnableContainer {
		if c := NewContainerCollector(); c != nil {
			m.collectors = append(m.collectors, c)
		}
	}

	if cfg.EnableProcess {
		m.collectors = append(m.collectors, NewProcessCollector())
	}

	if cfg.EnableNvidia {
		if c := NewNvidiaCollector(cfg.CollectGPUProcesses); c != nil {
			m.collectors = append(m.collectors, c)
		}
	}

	if cfg.EnableVLLM {
		m.collectors = append(m.collectors, NewVLLMCollector())
	}

	log.Printf("Initialized %d collectors", len(m.collectors))
	return m
}

func (m *Manager) CollectStatic(s *StaticMetrics) {
	for _, c := range m.collectors {
		c.CollectStatic(s)
	}

	m.mu.Lock()
	m.static = s
	m.staticJSON = structToRecord(s)
	m.mu.Unlock()
}

func (m *Manager) CollectDynamic(d *DynamicMetrics) formatting.Record {
	d.Timestamp = probing.GetTimestamp()
	for _, c := range m.collectors {
		c.CollectDynamic(d)
	}

	record := structToRecord(d)

	m.mu.RLock()
	for k, v := range m.staticJSON {
		record[k] = v
	}
	m.mu.RUnlock()

	return record
}

func (m *Manager) Close() {
	for _, c := range m.collectors {
		if err := c.Close(); err != nil {
			log.Printf("Error closing collector %s: %v", c.Name(), err)
		}
	}
}

func (m *Manager) CollectorNames() []string {
	names := make([]string, len(m.collectors))
	for i, c := range m.collectors {
		names[i] = c.Name()
	}
	return names
}

func (m *Manager) GetStatic() *StaticMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.static
}

func (m *Manager) GetStaticRecord() formatting.Record {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.staticJSON
}

func structToRecord(v interface{}) formatting.Record {
	data, _ := json.Marshal(v)
	var result formatting.Record
	json.Unmarshal(data, &result)
	return result
}
