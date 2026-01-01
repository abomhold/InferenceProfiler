package collectors

import (
	"log"
	"sync" // Added sync package

	"github.com/google/uuid"
)

// CollectorConfig controls which collectors are enabled
type CollectorConfig struct {
	CPU         bool
	Memory      bool
	Disk        bool
	Network     bool
	Container   bool
	Nvidia      bool
	NvidiaProcs bool
	VLLM        bool
	Processes   bool
}

// CollectorManager aggregates all collectors
type CollectorManager struct {
	cfg                      CollectorConfig
	nvidia                   *NvidiaCollector
	dynamicMetricsCollectors []func(*DynamicMetrics)
	staticMetricsCollectors  []func(*StaticMetrics)
}

// NewCollectorManager creates a new collector manager
func NewCollectorManager(cfg CollectorConfig) *CollectorManager {
	cm := &CollectorManager{cfg: cfg}

	if cm.cfg.CPU {
		cm.staticMetricsCollectors = append(cm.staticMetricsCollectors, CollectCPUStatic)
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, CollectCPUDynamic)
	}
	if cm.cfg.Memory {
		cm.staticMetricsCollectors = append(cm.staticMetricsCollectors, CollectMemoryStatic)
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, CollectMemoryDynamic)
	}
	if cm.cfg.Disk {
		cm.staticMetricsCollectors = append(cm.staticMetricsCollectors, CollectDiskStatic)
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, CollectDiskDynamic)
	}
	if cm.cfg.Network {
		cm.staticMetricsCollectors = append(cm.staticMetricsCollectors, CollectNetworkStatic)
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, CollectNetworkDynamic)
	}
	if cm.cfg.Container && isCgroupDir() {
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, CollectContainerDynamic)
		cm.staticMetricsCollectors = append(cm.staticMetricsCollectors, CollectContainerStatic)
	} else if cm.cfg.Container && !isCgroupDir() {
		log.Printf("WARNING: Container collector is enabled but no cgroup directory")
	}
	if cm.cfg.Processes {
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, CollectProcessesDynamic)
	}
	if cfg.Nvidia {
		cm.nvidia = NewNvidiaCollector(cfg.NvidiaProcs)
		if err := cm.nvidia.Init(); err != nil {
			log.Printf("WARNING: NVIDIA collector initialization failed: %v", err)
		}
		cm.staticMetricsCollectors = append(cm.staticMetricsCollectors, cm.nvidia.CollectNvidiaStatic)
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, cm.nvidia.CollectNvidiaDynamic)
	}
	if cfg.VLLM {
		cm.dynamicMetricsCollectors = append(cm.dynamicMetricsCollectors, CollectVLLMDynamic)
	}

	return cm
}

// CollectDynamicMetrics collects all dynamic metrics
func (cm *CollectorManager) CollectDynamicMetrics() *DynamicMetrics {
	m := &DynamicMetrics{
		Timestamp: GetTimestamp(),
	}

	var wg sync.WaitGroup

	for _, collector := range cm.dynamicMetricsCollectors {
		wg.Add(1)
		// Launch goroutine
		go func(c func(*DynamicMetrics)) {
			defer wg.Done()
			c(m)
		}(collector)
	}

	// Wait for all collectors to finish populating 'm'
	wg.Wait()

	return m
}

// CollectStaticMetrics collects all static system information
func (cm *CollectorManager) CollectStaticMetrics(sessionUUID uuid.UUID) *StaticMetrics {
	m := &StaticMetrics{
		UUID: sessionUUID.String(),
	}
	// Static metrics are usually collected sequentially as they are done once at startup
	// and often require order or are fast enough not to need complex concurrency.
	for _, collector := range cm.staticMetricsCollectors {
		collector(m)
	}
	return m
}

// Close cleans up all collectors
func (cm *CollectorManager) Close() {
	if cm.nvidia != nil {
		cm.nvidia.Cleanup()
	}
}
