package src

import (
	"InferenceProfiler/src"
	"os"
	"strings"
	"time"

	"github.com/inference-profiler/utils"
)

// Snapshot contains all dynamic metrics from a single collection cycle.
type Snapshot struct {
	Timestamp int64 `json:"timestamp"`

	CPU       src.CPUMetrics       `json:"cpu"`
	Memory    MemoryMetrics        `json:"mem"`
	Disk      src.DiskMetrics      `json:"disk"`
	Network   NetMetrics           `json:"net"`
	Container src.ContainerMetrics `json:"containers"`
	NVIDIA    []GPUMetrics         `json:"nvidia,omitempty"`
	VLLM      VLLMMetrics          `json:"vllm,omitempty"`
	Processes *ProcessCollection   `json:"processes,omitempty"`
}

// StaticInfo contains all static system information.
type StaticInfo struct {
	UUID   string     `json:"uuid"`
	VMID   string     `json:"vId"`
	Host   HostInfo   `json:"host"`
	NVIDIA *GPUStatic `json:"nvidia,omitempty"`
}

// HostInfo contains static host information.
type HostInfo struct {
	Hostname      string           `json:"hostname"`
	Kernel        string           `json:"kernel"`
	BootTime      int64            `json:"boot_time"`
	NumProcessors int              `json:"vNumProcessors"`
	CPUType       string           `json:"vCpuType"`
	CPUCache      map[string]int64 `json:"vCpuCache"`
	KernelInfo    string           `json:"vKernelInfo"`
	MemoryTotal   int64            `json:"vMemoryTotalBytes"`
	SwapTotal     int64            `json:"vSwapTotalBytes"`
}

// Manager coordinates all metric collectors.
type Manager struct {
	collectProcesses bool
	nvml             *NVMLCollector
}

// NewManager creates a new collector manager.
// collectProcesses enables per-process metric collection (expensive).
func NewManager(collectProcesses bool) *Manager {
	m := &Manager{
		collectProcesses: collectProcesses,
		nvml:             NewNVMLCollector(),
	}
	m.nvml.Init()
	return m
}

// Collect gathers all dynamic metrics.
func (m *Manager) Collect() Snapshot {
	s := Snapshot{
		Timestamp: time.Now().UnixMilli(),
		CPU:       src.CollectCPU(),
		Memory:    CollectMemory(),
		Disk:      src.CollectDisk(),
		Network:   CollectNet(),
		Container: src.CollectContainer(),
		NVIDIA:    m.nvml.Collect(),
		VLLM:      CollectVLLM(),
	}

	if m.collectProcesses {
		procs := CollectProcesses()
		s.Processes = &procs
	}

	return s
}

// GetStatic gathers all static system information.
func (m *Manager) GetStatic(sessionUUID string) StaticInfo {
	cpuStatic := src.CollectCPUStatic()
	memStatic := CollectMemoryStatic()

	hostname, _ := os.Hostname()
	uname, _ := utils.ReadFile("/proc/version")

	return StaticInfo{
		UUID: sessionUUID,
		VMID: getVMID(),
		Host: HostInfo{
			Hostname:      hostname,
			Kernel:        uname,
			BootTime:      getBootTime(),
			NumProcessors: cpuStatic.NumProcessors,
			CPUType:       cpuStatic.Model,
			CPUCache:      cpuStatic.Cache,
			KernelInfo:    cpuStatic.KernelInfo,
			MemoryTotal:   memStatic.TotalBytes,
			SwapTotal:     memStatic.SwapTotalBytes,
		},
		NVIDIA: m.nvml.GetStatic(),
	}
}

// Close releases all collector resources.
func (m *Manager) Close() {
	if m.nvml != nil {
		m.nvml.Close()
	}
}

// getBootTime reads system boot time from /proc/stat.
func getBootTime() int64 {
	lines, _ := utils.ReadLines("/proc/stat")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				v, _ := src.parseInt(fields[1]), int64(0)
				return v
			}
		}
	}
	return 0
}

// getVMID attempts to get VM/instance ID from various sources.
func getVMID() string {
	// Try DMI product UUID
	if id, _ := utils.ReadFile("/sys/class/dmi/id/product_uuid"); id != "" && id != "None" {
		return id
	}

	// Try machine ID as fallback
	if id, _ := utils.ReadFile("/etc/machine-id"); id != "" {
		return id
	}

	return "unavailable"
}
