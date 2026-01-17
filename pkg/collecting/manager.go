package collecting

import (
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"log"
	"sync"
)

type Manager struct {
	collectors []Collector
	static     *StaticMetrics
	staticJSON exporting.Record
	concurrent bool
	mu         sync.RWMutex
}

func NewManager(cfg *utils.Config) *Manager {
	m := &Manager{
		collectors: make([]Collector, 0, 8),
		concurrent: cfg.Concurrent,
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
		m.collectors = append(m.collectors, NewProcessCollector(cfg.Concurrent))
	}

	if cfg.EnableNvidia {
		if c := NewNvidiaCollector(cfg.CollectGPUProcesses); c != nil {
			m.collectors = append(m.collectors, c)
		}
	}

	if cfg.EnableVLLM {
		m.collectors = append(m.collectors, NewVLLMCollector())
	}

	mode := "sequential"
	if m.concurrent {
		mode = "concurrent"
	}
	log.Printf("Initialized %d collectors (%s)", len(m.collectors), mode)
	return m
}

func (m *Manager) CollectStatic(s *StaticMetrics) {
	if m.concurrent {
		m.collectStaticConcurrent(s)
	} else {
		for _, c := range m.collectors {
			c.CollectStatic(s)
		}
	}

	m.mu.Lock()
	m.static = s
	m.staticJSON = structToRecord(s)
	m.mu.Unlock()
}

func (m *Manager) collectStaticConcurrent(s *StaticMetrics) {
	var wg sync.WaitGroup
	wg.Add(len(m.collectors))
	for _, c := range m.collectors {
		go func(col Collector) {
			defer wg.Done()
			col.CollectStatic(s)
		}(c)
	}
	wg.Wait()
}

func (m *Manager) CollectDynamic(d *DynamicMetrics) exporting.Record {
	d.Timestamp = utils.GetTimestamp()

	if m.concurrent {
		m.collectDynamicConcurrent(d)
	} else {
		for _, c := range m.collectors {
			c.CollectDynamic(d)
		}
	}

	record := structToRecord(d)

	m.mu.RLock()
	for k, v := range m.staticJSON {
		record[k] = v
	}
	m.mu.RUnlock()

	return record
}

func (m *Manager) collectDynamicConcurrent(d *DynamicMetrics) {
	var wg sync.WaitGroup
	wg.Add(len(m.collectors))
	for _, c := range m.collectors {
		go func(col Collector) {
			defer wg.Done()
			col.CollectDynamic(d)
		}(c)
	}
	wg.Wait()
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

func (m *Manager) GetStaticRecord() exporting.Record {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.staticJSON
}

func structToRecord(v interface{}) exporting.Record {
	data, _ := json.Marshal(v)
	var result exporting.Record
	json.Unmarshal(data, &result)

	// Handle special fields excluded by json:"-"
	switch t := v.(type) {
	case *DynamicMetrics:
		if len(t.GPUs) > 0 {
			result["_gpus"] = t.GPUs
		}
		if len(t.Processes) > 0 {
			result["_processes"] = t.Processes
		}
	}

	return result
}
