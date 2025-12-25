package collectors

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Manager coordinates all metric collectors.
type Manager struct {
	collectProcesses bool

	cpu       *CPUCollector
	mem       *MemoryCollector
	disk      *DiskCollector
	net       *NetCollector
	container *ContainerCollector
	proc      *ProcessCollector
	nvml      *NVMLCollector
	vllm      *VLLMCollector
}

// NewManager creates a new collector manager.
// collectProcesses enables per-process metric collection (expensive).
func NewManager(collectProcesses bool) *Manager {
	m := &Manager{
		collectProcesses: collectProcesses,
		cpu:              &CPUCollector{},
		mem:              &MemoryCollector{},
		disk:             &DiskCollector{},
		net:              &NetCollector{},
		container:        &ContainerCollector{},
		proc:             &ProcessCollector{},
		nvml:             NewNVMLCollector(),
		vllm:             &VLLMCollector{},
	}
	err := m.nvml.Init()
	if err != nil {
		log.Println("Failed to initialize NVML:", err)
		m.nvml = nil
	}
	return m
}

// Collect gathers all dynamic metrics.
func (m *Manager) Collect() Snapshot {
	s := Snapshot{
		Timestamp: time.Now().UnixMilli(),
		CPU:       m.cpu.Collect(),
		Memory:    m.mem.Collect(),
		Disk:      m.disk.Collect(),
		Network:   m.net.Collect(),
		Container: m.container.Collect(),
		NVIDIA:    m.nvml.Collect(),
		VLLM:      m.vllm.Collect(),
	}

	if m.collectProcesses {
		procs := m.proc.Collect()
		s.Processes = &procs
	}

	return s
}

// GetStatic gathers all static system information.
func (m *Manager) GetStatic(sessionUUID uuid.UUID) StaticInfo {
	cpuStatic := m.cpu.GetStatic()
	memStatic := m.mem.GetStatic()

	hostname, _ := os.Hostname()
	uname, _ := m.cpu.ReadFile("/proc/version")

	return StaticInfo{
		UUID: sessionUUID,
		VMID: m.getVMID(),
		Host: HostInfo{
			Hostname:      hostname,
			Kernel:        uname,
			BootTime:      m.getBootTime(),
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
func (m *Manager) getBootTime() int64 {
	lines, _ := m.cpu.ReadLines("/proc/stat")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return m.cpu.ParseInt(fields[1])
			}
		}
	}
	return 0
}

// getVMID attempts to get VM/instance ID from various sources.
func (m *Manager) getVMID() string {
	// Try DMI product UUID
	if id, _ := m.cpu.ReadFile("/sys/class/dmi/id/product_uuid"); id != "" && id != "None" {
		return id
	}

	// Try machine ID as fallback
	if id, _ := m.cpu.ReadFile("/etc/machine-id"); id != "" {
		return id
	}

	return "unavailable"
}
