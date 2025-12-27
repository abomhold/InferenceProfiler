package collectors

import (
	"log"

	"github.com/google/uuid"
)

// CollectorManager aggregates all collectors
type CollectorManager struct {
	nvidia       *NvidiaCollector
	collectProcs bool
}

// NewCollectorManager creates a new collector manager
func NewCollectorManager(collectProcesses bool) *CollectorManager {
	nvidia := NewNvidiaCollector()
	if err := nvidia.Init(); err != nil {
		log.Printf("NVIDIA collector initialization failed: %v", err)
	}

	return &CollectorManager{
		nvidia:       nvidia,
		collectProcs: collectProcesses,
	}
}

// CollectMetrics collects all dynamic metrics
func (cm *CollectorManager) CollectMetrics() *DynamicMetrics {
	m := &DynamicMetrics{
		Timestamp: GetTimestamp(),
	}

	// VM-level metrics
	CollectCPUDynamic(m)
	CollectMemoryDynamic(m)
	CollectDiskDynamic(m)
	CollectNetworkDynamic(m)

	// Container metrics
	CollectContainerDynamic(m)

	// NVIDIA GPU metrics
	cm.nvidia.CollectDynamic(m)

	// vLLM metrics
	CollectVLLMDynamic(m)

	// Process metrics
	if cm.collectProcs {
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
	CollectCPUStatic(m)

	// Memory static info
	CollectMemoryStatic(m)

	// Disk static info
	CollectDiskStatic(m)

	// Network static info
	CollectNetworkStatic(m)

	// Container static info
	CollectContainerStatic(m)

	// NVIDIA static info
	cm.nvidia.CollectStatic(m)

	return m
}

// Close cleans up all collectors
func (cm *CollectorManager) Close() {
	if cm.nvidia != nil {
		cm.nvidia.Cleanup()
	}
}
