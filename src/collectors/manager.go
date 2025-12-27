package collectors

import (
	"fmt"
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

// CollectMetrics collects all dynamic metrics into a flat map
func (cm *CollectorManager) CollectMetrics() DynamicMetrics {
	metrics := NewDynamicMetrics()

	metrics.Merge(CollectCPUDynamic())
	metrics.Merge(CollectMemoryDynamic())
	metrics.Merge(CollectDiskDynamic())
	metrics.Merge(CollectNetworkDynamic())
	metrics.Merge(CollectContainerDynamic())

	gpuMetrics := cm.nvidia.CollectDynamic()
	for i, gpu := range gpuMetrics {
		prefix := fmt.Sprintf("nvidia_%d_", i)
		metrics.MergeWithPrefix(prefix, gpu)
	}

	// vLLM metrics (all prefixed with vllm*)
	metrics.Merge(CollectVLLM())

	// Process metrics (prefixed with proc_N_*)
	if cm.collectProcs {
		count, procs := CollectProcesses()
		metrics["pNumProcesses"] = NewMetric(count)
		for i, proc := range procs {
			prefix := fmt.Sprintf("proc_%d_", i)
			metrics.MergeWithPrefix(prefix, proc)
		}
	}

	return metrics
}

func (cm *CollectorManager) GetStaticInfo(sessionUUID uuid.UUID) StaticMetrics {
	info := NewStaticMetrics(sessionUUID)
	info.Merge(CollectCPUStatic())
	info.Merge(GetMemoryStaticInfo())
	info.Merge(CollectDiskStatic())
	info.Merge(CollectNetworkStatic())
	info.Merge(CollectContainerStatic())
	info.Merge(cm.nvidia.CollectStatic())

	return info
}

func (cm *CollectorManager) Close() {
	if cm.nvidia != nil {
		cm.nvidia.Cleanup()
	}
}
