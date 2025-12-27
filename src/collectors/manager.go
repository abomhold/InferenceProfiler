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
			log.Printf("NVIDIA collector initialization failed: %v", err)
		}
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
	if cm.cfg.Container {
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

// GetStaticInfo collects all static system information
func (cm *CollectorManager) GetStaticInfo(sessionUUID uuid.UUID) *StaticMetrics {
	m := &StaticMetrics{
		UUID: sessionUUID.String(),
	}

	// CPU static info (includes kernel info, boot time, VM ID, hostname)
	if cm.cfg.CPU {
		CollectCPUStatic(m)
	}

	// Memory static info
	if cm.cfg.Memory {
		CollectMemoryStatic(m)
	}

	// Disk static info
	if cm.cfg.Disk {
		CollectDiskStatic(m)
	}

	// Network static info
	if cm.cfg.Network {
		CollectNetworkStatic(m)
	}

	// Container static info
	if cm.cfg.Container {
		CollectContainerStatic(m)
	}

	// NVIDIA static info
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
