package src

import (
	"fmt"
	"sync"

	"github.com/inference-profiler/utils"
)

// GPUMetrics contains per-GPU measurements from NVML.
type GPUMetrics struct {
	// Identification
	Index int `json:"gGpuIndex"`

	// Utilization (percent)
	UtilizationGPU exporter.Timed[int] `json:"gUtilizationGpu"`
	UtilizationMem exporter.Timed[int] `json:"gUtilizationMem"`

	// Memory (MB)
	MemoryTotalMB exporter.Timed[int64] `json:"gMemoryTotalMb"`
	MemoryUsedMB  exporter.Timed[int64] `json:"gMemoryUsedMb"`
	MemoryFreeMB  exporter.Timed[int64] `json:"gMemoryFreeMb"`
	BAR1UsedMB    exporter.Timed[int64] `json:"gBar1UsedMb"`
	BAR1FreeMB    exporter.Timed[int64] `json:"gBar1FreeMb"`

	// Thermal
	TemperatureC exporter.Timed[int] `json:"gTemperatureC"`
	FanSpeed     exporter.Timed[int] `json:"gFanSpeed"`

	// Power (watts)
	PowerDrawW  exporter.Timed[float64] `json:"gPowerDrawW"`
	PowerLimitW exporter.Timed[float64] `json:"gPowerLimitW"`

	// Clocks (MHz)
	ClockGraphicsMHz exporter.Timed[int] `json:"gClockGraphicsMhz"`
	ClockSMMHz       exporter.Timed[int] `json:"gClockSmMhz"`
	ClockMemMHz      exporter.Timed[int] `json:"gClockMemMhz"`

	// PCIe (KB/s)
	PCIeTxKBps exporter.Timed[int] `json:"gPcieTxKbps"`
	PCIeRxKBps exporter.Timed[int] `json:"gPcieRxKbps"`

	// State
	PerfState exporter.Timed[string] `json:"gPerfState"`

	// ECC errors
	ECCSingleBit exporter.Timed[int64] `json:"gEccSingleBitErrors"`
	ECCDoubleBit exporter.Timed[int64] `json:"gEccDoubleBitErrors"`

	// Processes
	ProcessCount int      `json:"gProcessCount"`
	Processes    []string `json:"gProcesses"` // "PID: name (VRAM MB)"
}

// GPUStatic contains static GPU information.
type GPUStatic struct {
	DriverVersion string          `json:"gDriverVersion"`
	CUDAVersion   int             `json:"gCudaVersion"`
	GPUs          []GPUStaticInfo `json:"gpus"`
}

// GPUStaticInfo contains static info for a single GPU.
type GPUStaticInfo struct {
	Name             string `json:"gName"`
	UUID             string `json:"gUuid"`
	TotalMemoryMB    int64  `json:"gTotalMemoryMb"`
	PCIBusID         string `json:"gPciBusId"`
	MaxGraphicsClock int    `json:"gMaxGraphicsClock"`
	MaxSMClock       int    `json:"gMaxSmClock"`
	MaxMemClock      int    `json:"gMaxMemClock"`
}

// NVMLCollector manages NVIDIA GPU metric collection.
// Uses go-nvml bindings (requires NVIDIA drivers).
type NVMLCollector struct {
	mu          sync.Mutex
	initialized bool
	available   bool
}

// NewNVMLCollector creates a new NVML collector.
// Call Init() before collecting metrics.
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

	// NVML initialization happens through CGO bindings
	// This is a stub - actual implementation requires go-nvml
	c.initialized = true
	c.available = false // Will be set true if NVML loads successfully

	return nil
}

// Collect gathers metrics from all GPUs.
// Returns empty slice if NVML unavailable.
func (c *NVMLCollector) Collect() []GPUMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.available {
		return nil
	}

	// Stub - actual implementation uses go-nvml
	return nil
}

// GetStatic returns static GPU information.
func (c *NVMLCollector) GetStatic() *GPUStatic {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.available {
		return nil
	}

	// Stub - actual implementation uses go-nvml
	return nil
}

// Close shuts down NVML.
func (c *NVMLCollector) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		// nvml.Shutdown()
		c.initialized = false
		c.available = false
	}
}

// getProcessName reads the process name from /proc/[pid]/comm.
func getProcessName(pid int) string {
	content, _ := utils.readFile(fmt.Sprintf("/proc/%d/comm", pid))
	if content == "" {
		return "unknown"
	}
	return content
}

// Available returns true if NVML is available.
func (c *NVMLCollector) Available() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.available
}
