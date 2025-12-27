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
	Nvidia      bool
	NvidiaProcs bool // Collect GPU processes (requires Nvidia=true)
	VLLM        bool
	Processes   bool
}

// CollectorManager aggregates all collectors
type CollectorManager struct {
	cfg    CollectorConfig
	nvidia *NvidiaCollector
}

// NewCollectorManager creates a new collector manager
func NewCollectorManager(cfg CollectorConfig) *CollectorManager {
	cm := &CollectorManager{cfg: cfg}

	if cfg.Nvidia {
		cm.nvidia = NewNvidiaCollector(cfg.NvidiaProcs)
		if err := cm.nvidia.Init(); err != nil {
			log.Printf("WARNGING: NVIDIA collector initialization failed: %v", err)
		}
	}

	if cfg.Container && !isCgroupDir() {
		log.Println("WARNING: Cgroup directory not found")
	}

	return cm
}

// CollectMetrics collects all dynamic metrics
func (cm *CollectorManager) CollectMetrics() *DynamicMetrics {
	m := &DynamicMetrics{
		Timestamp: GetTimestamp(),
	}

	// VM-level metrics
	if cm.cfg.CPU {
		CollectCPUDynamic(m)
	}
	if cm.cfg.Memory {
		CollectMemoryDynamic(m)
	}
	if cm.cfg.Disk {
		CollectDiskDynamic(m)
	}
	if cm.cfg.Network {
		CollectNetworkDynamic(m)
	}

	// Container metrics
	if cm.cfg.Container && isCgroupDir() {
		CollectContainerDynamic(m)
	}

	// NVIDIA GPU metrics
	if cm.cfg.Nvidia && cm.nvidia != nil {
		cm.nvidia.CollectDynamic(m)
	}

	// vLLM metrics
	if cm.cfg.VLLM {
		CollectVLLMDynamic(m)
	}

	// Process metrics
	if cm.cfg.Processes {
		CollectProcessesDynamic(m)
	}

	return m
}

// GetStaticMetrics collects all static system information
func (cm *CollectorManager) GetStaticMetrics(sessionUUID uuid.UUID) *StaticMetrics {
	m := &StaticMetrics{
		UUID: sessionUUID.String(),
	}

	if cm.cfg.CPU {
		CollectCPUStatic(m)
	}
	if cm.cfg.Memory {
		CollectMemoryStatic(m)
	}
	if cm.cfg.Disk {
		CollectDiskStatic(m)
	}
	if cm.cfg.Network {
		CollectNetworkStatic(m)
	}
	if cm.cfg.Container && isCgroupDir() {
		CollectContainerStatic(m)
	}
	if cm.cfg.Nvidia && cm.nvidia != nil {
		cm.nvidia.CollectStatic(m)
	}

	return m
}

// Close cleans up all collectors
func (cm *CollectorManager) Close() {
	if cm.nvidia != nil {
		cm.nvidia.Cleanup()
	}
}
