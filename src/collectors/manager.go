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

	// VM-level metrics (all prefixed with v*)
	metrics.Merge(CollectCPU())
	metrics.Merge(CollectMemory())
	metrics.Merge(CollectDisk())
	metrics.Merge(CollectNetwork())

	// Container metrics (all prefixed with c*)
	metrics.Merge(CollectContainer())

	// NVIDIA GPU metrics (prefixed with nvidia_N_*)
	gpuMetrics := cm.nvidia.Collect()
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

// GetStaticInfo collects all static system information into a flat map
func (cm *CollectorManager) GetStaticInfo(sessionUUID uuid.UUID) StaticInfo {
	info := NewStaticInfo(sessionUUID)

	// VM identification
	info["vId"] = GetVMID()

	// Host information (v* prefix)
	info["vHostname"] = GetHostname()
	info["vBootTime"] = GetBootTime()

	numCPUs, cpuType, cpuCache, kernelInfo := GetCPUStaticInfo()
	info["vNumProcessors"] = numCPUs
	info["vCpuType"] = cpuType
	info["vKernelInfo"] = kernelInfo

	// CPU cache (vCpuCache_L1d, vCpuCache_L2, etc.)
	for level, size := range cpuCache {
		info["vCpuCache_"+level] = size
	}

	memTotal, swapTotal := GetMemoryStaticInfo()
	info["vMemoryTotalBytes"] = memTotal
	info["vSwapTotalBytes"] = swapTotal

	// NVIDIA static info
	nvidiaStatic := cm.nvidia.GetStaticInfo()
	if nvidiaStatic != nil {
		if driver, ok := nvidiaStatic["nvidiaDriverVersion"].(string); ok {
			info["nvidiaDriverVersion"] = driver
		}
		if cuda, ok := nvidiaStatic["nvidiaCudaVersion"].(string); ok {
			info["nvidiaCudaVersion"] = cuda
		}

		// Per-GPU static info (nvidia_N_* prefix)
		if gpus, ok := nvidiaStatic["gpus"].([]map[string]interface{}); ok {
			for i, gpu := range gpus {
				prefix := fmt.Sprintf("nvidia_%d_", i)
				for k, v := range gpu {
					info[prefix+k] = v
				}
			}
		}
	}

	return info
}

// Close cleans up all collectors
func (cm *CollectorManager) Close() {
	if cm.nvidia != nil {
		cm.nvidia.Cleanup()
	}
}
