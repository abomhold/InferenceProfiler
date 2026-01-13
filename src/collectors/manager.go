package collectors

import (
	"log"

	"github.com/google/uuid"
)

// CollectorConfig controls which collectors are enabled
type CollectorConfig struct {
	CPU         bool
	Memory      bool
	Disk        bool
	Network     bool
	Container   bool
	Processes   bool
	Nvidia      bool
	NvidiaProcs bool
	VLLM        bool
}

// CollectorManager aggregates and orchestrates all collectors
type CollectorManager struct {
	collectors []Collector
}

// NewCollectorManager creates a new collector manager with the specified configuration
func NewCollectorManager(cfg CollectorConfig) *CollectorManager {
	cm := &CollectorManager{
		collectors: make([]Collector, 0, 9), // Pre-allocate for typical case
	}

	// Register enabled collectors
	if cfg.CPU {
		cm.Register(NewCPUCollector())
	}
	if cfg.Memory {
		cm.Register(NewMemoryCollector())
	}
	if cfg.Disk {
		cm.Register(NewDiskCollector())
	}
	if cfg.Network {
		cm.Register(NewNetworkCollector())
	}
	if cfg.Container {
		if c := NewContainerCollector(); c != nil {
			cm.Register(c)
		}
	}
	if cfg.Processes {
		cm.Register(NewProcessCollector())
	}
	if cfg.Nvidia {
		if c := NewNvidiaCollector(cfg.NvidiaProcs); c != nil {
			cm.Register(c)
		}
	}
	if cfg.VLLM {
		cm.Register(NewVLLMCollector())
	}

	return cm
}

// Register adds a collector to the manager
func (cm *CollectorManager) Register(c Collector) {
	if c != nil {
		cm.collectors = append(cm.collectors, c)
		log.Printf("Registered collector: %s", c.Name())
	}
}

// CollectStaticMetrics collects all static system information
func (cm *CollectorManager) CollectStaticMetrics(sessionUUID uuid.UUID) *StaticMetrics {
	m := &StaticMetrics{
		UUID: sessionUUID.String(),
	}

	for _, c := range cm.collectors {
		c.CollectStatic(m)
	}

	return m
}

// CollectDynamicMetrics collects all dynamic metrics
func (cm *CollectorManager) CollectDynamicMetrics() *DynamicMetrics {
	m := &DynamicMetrics{
		Timestamp: GetTimestamp(),
	}

	for _, c := range cm.collectors {
		c.CollectDynamic(m)
	}

	return m
}

// Close cleans up all collectors
func (cm *CollectorManager) Close() {
	for _, c := range cm.collectors {
		if err := c.Close(); err != nil {
			log.Printf("Warning: error closing %s collector: %v", c.Name(), err)
		}
	}
}

// CollectorCount returns the number of registered collectors
func (cm *CollectorManager) CollectorCount() int {
	return len(cm.collectors)
}

// CollectorNames returns the names of all registered collectors
func (cm *CollectorManager) CollectorNames() []string {
	names := make([]string, len(cm.collectors))
	for i, c := range cm.collectors {
		names[i] = c.Name()
	}
	return names
}
