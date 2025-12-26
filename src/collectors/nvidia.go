package collectors

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NvidiaCollector handles NVIDIA GPU metrics collection via go-nvml
type NvidiaCollector struct {
	initialized bool
	mu          sync.Mutex
}

// NewNvidiaCollector creates a new NVIDIA collector
func NewNvidiaCollector() *NvidiaCollector {
	return &NvidiaCollector{}
}

// Init initializes NVML
func (n *NvidiaCollector) Init() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.initialized {
		return nil
	}

	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("failed to initialize NVML: %s", nvml.ErrorString(ret))
	}

	n.initialized = true
	log.Println("NVIDIA NVML initialized successfully")
	return nil
}

// Collect collects GPU metrics from all devices
func (n *NvidiaCollector) Collect() []map[string]MetricValue {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.initialized {
		return nil
	}

	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) {
		log.Printf("Failed to get device count: %s", nvml.ErrorString(ret))
		return nil
	}

	gpuMetrics := make([]map[string]MetricValue, 0, count)

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if !errors.Is(ret, nvml.SUCCESS) {
			continue
		}
		gpuMetrics = append(gpuMetrics, n.collectDeviceMetrics(device, i))
	}

	return gpuMetrics
}

func (n *NvidiaCollector) collectDeviceMetrics(device nvml.Device, index int) map[string]MetricValue {
	metrics := make(map[string]MetricValue)
	metrics["nvidiaGpuIndex"] = NewMetric(int64(index))

	fetch := func(fn func() (uint32, nvml.Return)) (uint32, int64) {
		return ProbeFunction(func() (uint32, error) {
			val, ret := fn()
			if ret != nvml.SUCCESS {
				return 0, fmt.Errorf("%s", nvml.ErrorString(ret))
			}
			return val, nil
		}, 0)
	}

	util, ts := ProbeFunction(func() (nvml.Utilization, error) {
		u, ret := device.GetUtilizationRates()
		if ret != nvml.SUCCESS {
			return nvml.Utilization{}, fmt.Errorf("%s", nvml.ErrorString(ret))
		}
		return u, nil
	}, nvml.Utilization{})
	metrics["nvidiaUtilizationGpu"] = NewMetricWithTime(int64(util.Gpu), ts)
	metrics["nvidiaUtilizationMem"] = NewMetricWithTime(int64(util.Memory), ts)

	mem, ts := ProbeFunction(func() (nvml.Memory, error) {
		m, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			return nvml.Memory{}, fmt.Errorf("%s", nvml.ErrorString(ret))
		}
		return m, nil
	}, nvml.Memory{})
	metrics["nvidiaMemoryTotalMb"] = NewMetricWithTime(int64(mem.Total>>20), ts)
	metrics["nvidiaMemoryUsedMb"] = NewMetricWithTime(int64(mem.Used>>20), ts)
	metrics["nvidiaMemoryFreeMb"] = NewMetricWithTime(int64(mem.Free>>20), ts)

	bar1, ts := ProbeFunction(func() (nvml.BAR1Memory, error) {
		b, ret := device.GetBAR1MemoryInfo()
		if ret != nvml.SUCCESS {
			return nvml.BAR1Memory{}, fmt.Errorf("%s", nvml.ErrorString(ret))
		}
		return b, nil
	}, nvml.BAR1Memory{})
	metrics["nvidiaBar1UsedMb"] = NewMetricWithTime(int64(bar1.Bar1Used>>20), ts)
	metrics["nvidiaBar1FreeMb"] = NewMetricWithTime(int64(bar1.Bar1Free>>20), ts)

	temp, ts := fetch(func() (uint32, nvml.Return) { return device.GetTemperature(nvml.TEMPERATURE_GPU) })
	metrics["nvidiaTemperatureC"] = NewMetricWithTime(int64(temp), ts)

	fan, ts := fetch(device.GetFanSpeed)
	metrics["nvidiaFanSpeed"] = NewMetricWithTime(int64(fan), ts)

	power, ts := fetch(device.GetPowerUsage)
	metrics["nvidiaPowerDrawW"] = NewMetricWithTime(float64(power)/1000.0, ts)

	limit, ts := fetch(device.GetEnforcedPowerLimit)
	metrics["nvidiaPowerLimitW"] = NewMetricWithTime(float64(limit)/1000.0, ts)

	gfxClock, ts := fetch(func() (uint32, nvml.Return) { return device.GetClockInfo(nvml.CLOCK_GRAPHICS) })
	metrics["nvidiaClockGraphicsMhz"] = NewMetricWithTime(int64(gfxClock), ts)

	smClock, ts := fetch(func() (uint32, nvml.Return) { return device.GetClockInfo(nvml.CLOCK_SM) })
	metrics["nvidiaClockSmMhz"] = NewMetricWithTime(int64(smClock), ts)

	memClock, ts := fetch(func() (uint32, nvml.Return) { return device.GetClockInfo(nvml.CLOCK_MEM) })
	metrics["nvidiaClockMemMhz"] = NewMetricWithTime(int64(memClock), ts)

	pcieTx, ts := fetch(func() (uint32, nvml.Return) { return device.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES) })
	metrics["nvidiaPcieTxKbps"] = NewMetricWithTime(int64(pcieTx), ts)

	pcieRx, ts := fetch(func() (uint32, nvml.Return) { return device.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES) })
	metrics["nvidiaPcieRxKbps"] = NewMetricWithTime(int64(pcieRx), ts)

	pstate, ts := ProbeFunction(func() (nvml.Pstates, error) {
		p, ret := device.GetPerformanceState()
		if ret != nvml.SUCCESS {
			return 0, fmt.Errorf("%s", nvml.ErrorString(ret))
		}
		return p, nil
	}, nvml.Pstates(0))
	metrics["nvidiaPerfState"] = NewMetricWithTime(fmt.Sprintf("P%d", int(pstate)), ts)

	procs, ts := n.getRunningProcesses(device)
	metrics["nvidiaProcessCount"] = NewMetricWithTime(int64(len(procs)), ts)
	metrics["nvidiaProcesses"] = NewMetricWithTime(procs, ts)

	return metrics
}

func (n *NvidiaCollector) getRunningProcesses(device nvml.Device) ([]map[string]interface{}, int64) {
	ts := GetTimestamp()
	procs := make([]map[string]interface{}, 0)
	seen := make(map[uint32]bool)

	addProcess := func(pid uint32, usedMemory uint64) {
		if seen[pid] {
			return
		}
		seen[pid] = true
		procs = append(procs, map[string]interface{}{
			"pid":          pid,
			"name":         getProcessName(int(pid)),
			"usedMemoryMb": int64(usedMemory >> 20),
		})
	}

	processGetters := []func() ([]nvml.ProcessInfo, nvml.Return){
		device.GetComputeRunningProcesses,
		device.GetGraphicsRunningProcesses,
	}

	for _, getter := range processGetters {
		if list, ret := getter(); ret == nvml.SUCCESS {
			for _, p := range list {
				addProcess(p.Pid, p.UsedGpuMemory)
			}
		}
	}

	return procs, ts
}

// GetStaticInfo returns static GPU information
func (n *NvidiaCollector) GetStaticInfo() map[string]interface{} {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.initialized {
		return nil
	}

	info := make(map[string]interface{})

	getString := func(fn func() (string, nvml.Return)) string {
		if val, ret := fn(); ret == nvml.SUCCESS {
			return val
		}
		return ""
	}

	info["nvidiaDriverVersion"] = getString(nvml.SystemGetDriverVersion)

	if cudaVersion, ret := nvml.SystemGetCudaDriverVersion(); ret == nvml.SUCCESS {
		info["nvidiaCudaVersion"] = fmt.Sprintf("%d.%d", cudaVersion/1000, (cudaVersion%1000)/10)
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return info
	}

	gpus := make([]map[string]interface{}, 0, count)
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		gpu := make(map[string]interface{})
		gpu["nvidiaName"] = getString(device.GetName)
		gpu["nvidiaUuid"] = getString(device.GetUUID)

		if mem, ret := device.GetMemoryInfo(); ret == nvml.SUCCESS {
			gpu["nvidiaTotalMemoryMb"] = int64(mem.Total >> 20)
		}

		if pci, ret := device.GetPciInfo(); ret == nvml.SUCCESS {
			gpu["nvidiaPciBusId"] = pci.BusId
		}

		clockTypes := map[string]nvml.ClockType{
			"nvidiaMaxGraphicsClock": nvml.CLOCK_GRAPHICS,
			"nvidiaMaxSmClock":       nvml.CLOCK_SM,
			"nvidiaMaxMemClock":      nvml.CLOCK_MEM,
		}

		for key, clockID := range clockTypes {
			if val, ret := device.GetMaxClockInfo(clockID); ret == nvml.SUCCESS {
				gpu[key] = int64(val)
			}
		}

		gpus = append(gpus, gpu)
	}
	info["gpus"] = gpus

	return info
}

// Cleanup shuts down NVML
func (n *NvidiaCollector) Cleanup() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.initialized {
		nvml.Shutdown()
		n.initialized = false
	}
}

// getProcessName reads process name from /proc
func getProcessName(pid int) string {
	content, _ := ProbeFile(fmt.Sprintf("/proc/%d/comm", pid))
	if content != "" {
		return content
	}
	return "unknown"
}
