////go:build nvml

package collectors

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NVMLCollector manages NVIDIA GPU metric collection.
// Uses go-nvml bindings (requires NVIDIA drivers and libnvidia-ml.so).
type NVMLCollector struct {
	BaseCollector
	mu          sync.Mutex
	initialized bool
	available   bool
}

// NewNVMLCollector creates a new NVML collector.
func NewNVMLCollector() *NVMLCollector {
	return &NVMLCollector{}
}

// Init initializes NVML. Safe to call multiple times.
func (c *NVMLCollector) Init() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	ret := nvml.Init()
	c.initialized = true

	if ret != nvml.SUCCESS {
		c.available = false
		return fmt.Errorf("NVML init failed: %s", nvml.ErrorString(ret))
	}

	c.available = true
	return nil
}

// Collect gathers metrics from all GPUs.
// Returns empty slice if NVML unavailable.
func (c *NVMLCollector) Collect() []NvidiaMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.available {
		return nil
	}

	now := time.Now().UnixMilli()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil
	}

	metrics := make([]NvidiaMetrics, 0, count)

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		m := NvidiaMetrics{Index: i}

		// Utilization
		if util, ret := device.GetUtilizationRates(); ret == nvml.SUCCESS {
			m.UtilizationGPU = Timed[int]{Value: int(util.Gpu), Time: now}
			m.UtilizationMem = Timed[int]{Value: int(util.Memory), Time: now}
		}

		// Memory info
		if mem, ret := device.GetMemoryInfo(); ret == nvml.SUCCESS {
			m.MemoryTotalMB = Timed[int64]{Value: int64(mem.Total / 1024 / 1024), Time: now}
			m.MemoryUsedMB = Timed[int64]{Value: int64(mem.Used / 1024 / 1024), Time: now}
			m.MemoryFreeMB = Timed[int64]{Value: int64(mem.Free / 1024 / 1024), Time: now}
		}

		// BAR1 memory
		if bar1, ret := device.GetBAR1MemoryInfo(); ret == nvml.SUCCESS {
			m.BAR1UsedMB = Timed[int64]{Value: int64(bar1.Bar1Used / 1024 / 1024), Time: now}
			m.BAR1FreeMB = Timed[int64]{Value: int64(bar1.Bar1Free / 1024 / 1024), Time: now}
		}

		// Temperature
		if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
			m.TemperatureC = Timed[int]{Value: int(temp), Time: now}
		}

		// Fan speed
		if fan, ret := device.GetFanSpeed(); ret == nvml.SUCCESS {
			m.FanSpeed = Timed[int]{Value: int(fan), Time: now}
		}

		// Power
		if power, ret := device.GetPowerUsage(); ret == nvml.SUCCESS {
			m.PowerDrawW = Timed[float64]{Value: float64(power) / 1000.0, Time: now}
		}
		if limit, ret := device.GetEnforcedPowerLimit(); ret == nvml.SUCCESS {
			m.PowerLimitW = Timed[float64]{Value: float64(limit) / 1000.0, Time: now}
		}

		// Clocks
		if clock, ret := device.GetClockInfo(nvml.CLOCK_GRAPHICS); ret == nvml.SUCCESS {
			m.ClockGraphicsMHz = Timed[int]{Value: int(clock), Time: now}
		}
		if clock, ret := device.GetClockInfo(nvml.CLOCK_SM); ret == nvml.SUCCESS {
			m.ClockSMMHz = Timed[int]{Value: int(clock), Time: now}
		}
		if clock, ret := device.GetClockInfo(nvml.CLOCK_MEM); ret == nvml.SUCCESS {
			m.ClockMemMHz = Timed[int]{Value: int(clock), Time: now}
		}

		// PCIe throughput
		if tx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); ret == nvml.SUCCESS {
			m.PCIeTxKBps = Timed[int]{Value: int(tx), Time: now}
		}
		if rx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); ret == nvml.SUCCESS {
			m.PCIeRxKBps = Timed[int]{Value: int(rx), Time: now}
		}

		// Performance state
		if pstate, ret := device.GetPerformanceState(); ret == nvml.SUCCESS {
			m.PerfState = Timed[string]{Value: fmt.Sprintf("P%d", pstate), Time: now}
		}

		// ECC errors (volatile, aggregate)
		if ecc, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.VOLATILE_ECC); ret == nvml.SUCCESS {
			m.ECCSingleBit = Timed[int64]{Value: int64(ecc), Time: now}
		}
		if ecc, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_UNCORRECTED, nvml.VOLATILE_ECC); ret == nvml.SUCCESS {
			m.ECCDoubleBit = Timed[int64]{Value: int64(ecc), Time: now}
		}

		// Running processes
		procs := c.getRunningProcesses(device)
		m.ProcessCount = len(procs)
		m.Processes = procs

		metrics = append(metrics, m)
	}

	return metrics
}

// getRunningProcesses returns a list of process descriptions for a device.
func (c *NVMLCollector) getRunningProcesses(device nvml.Device) []string {
	var processes []string
	seen := make(map[uint32]bool)

	// Compute processes
	if procs, ret := device.GetComputeRunningProcesses(); ret == nvml.SUCCESS {
		for _, p := range procs {
			if seen[p.Pid] {
				continue
			}
			seen[p.Pid] = true
			name := c.getProcessName(int(p.Pid))
			memMB := p.UsedGpuMemory / 1024 / 1024
			processes = append(processes, fmt.Sprintf("%d: %s (%d MB)", p.Pid, name, memMB))
		}
	}

	// Graphics processes
	if procs, ret := device.GetGraphicsRunningProcesses(); ret == nvml.SUCCESS {
		for _, p := range procs {
			if seen[p.Pid] {
				continue
			}
			seen[p.Pid] = true
			name := c.getProcessName(int(p.Pid))
			memMB := p.UsedGpuMemory / 1024 / 1024
			processes = append(processes, fmt.Sprintf("%d: %s (%d MB)", p.Pid, name, memMB))
		}
	}

	return processes
}

// GetStatic returns static GPU information.
func (c *NVMLCollector) GetStatic() *NvidiaStatic {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.available {
		return nil
	}

	static := &NvidiaStatic{}

	// Driver version
	if ver, ret := nvml.SystemGetDriverVersion(); ret == nvml.SUCCESS {
		static.DriverVersion = ver
	}

	// CUDA version
	if ver, ret := nvml.SystemGetCudaDriverVersion(); ret == nvml.SUCCESS {
		static.CUDAVersion = ver
	}

	// Per-GPU static info
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return static
	}

	static.GPUs = make([]NvidiaStaticInfo, 0, count)

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		info := NvidiaStaticInfo{}

		// Name
		if name, ret := device.GetName(); ret == nvml.SUCCESS {
			info.Name = name
		}

		// UUID
		if uuid, ret := device.GetUUID(); ret == nvml.SUCCESS {
			info.UUID = uuid
		}

		// Total memory
		if mem, ret := device.GetMemoryInfo(); ret == nvml.SUCCESS {
			info.TotalMemoryMB = int64(mem.Total / 1024 / 1024)
		}

		// PCI bus ID
		if pci, ret := device.GetPciInfo(); ret == nvml.SUCCESS {
			// busId is a fixed-size byte array, convert to string
			busID := string(pci.BusId[:])
			// Trim null bytes
			if idx := strings.IndexByte(busID, 0); idx >= 0 {
				busID = busID[:idx]
			}
			info.PCIBusID = busID
		}

		// Max clocks
		if clock, ret := device.GetMaxClockInfo(nvml.CLOCK_GRAPHICS); ret == nvml.SUCCESS {
			info.MaxGraphicsClock = int(clock)
		}
		if clock, ret := device.GetMaxClockInfo(nvml.CLOCK_SM); ret == nvml.SUCCESS {
			info.MaxSMClock = int(clock)
		}
		if clock, ret := device.GetMaxClockInfo(nvml.CLOCK_MEM); ret == nvml.SUCCESS {
			info.MaxMemClock = int(clock)
		}

		static.GPUs = append(static.GPUs, info)
	}

	return static
}

// Close shuts down NVML.
func (c *NVMLCollector) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized && c.available {
		nvml.Shutdown()
	}
	c.initialized = false
	c.available = false
}

// getProcessName reads the process name from /proc/[pid]/comm.
func (c *NVMLCollector) getProcessName(pid int) string {
	content, _ := c.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if content == "" {
		return "unknown"
	}
	return strings.TrimSpace(content)
}

// Available returns true if NVML is available.
func (c *NVMLCollector) Available() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.available
}
