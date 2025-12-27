package collectors

import (
	"errors"
	"fmt"
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type NvidiaCollector struct {
	initialized bool
}

func NewNvidiaCollector() *NvidiaCollector { return &NvidiaCollector{} }

func nvmlProbe[T any](fn func() (T, nvml.Return), defaultValue T) (T, int64) {
	return ProbeFunction(func() (T, error) {
		val, ret := fn()
		if ret != nvml.SUCCESS {
			return defaultValue, fmt.Errorf("NVML error: %s", nvml.ErrorString(ret))
		}
		return val, nil
	}, defaultValue)
}

func (n *NvidiaCollector) Init() error {
	if n.initialized {
		return nil
	}
	if ret := nvml.Init(); !errors.Is(ret, nvml.SUCCESS) {
		return fmt.Errorf("failed to initialize NVML: %s", nvml.ErrorString(ret))
	}
	n.initialized = true
	log.Println("NVIDIA NVML initialized successfully")
	return nil
}

func (n *NvidiaCollector) Cleanup() {
	if n.initialized {
		nvml.Shutdown()
		n.initialized = false
	}
}

// --- Static Metrics ---

func (n *NvidiaCollector) CollectStatic() StaticMetrics {
	if !n.initialized {
		return nil
	}

	info := make(StaticMetrics)
	info["nvidiaDriverVersion"], _ = nvml.SystemGetDriverVersion()
	if cv, ret := nvml.SystemGetCudaDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		info["nvidiaCudaVersion"] = fmt.Sprintf("%d.%d", cv/1000, (cv%1000)/10)
	}

	count, _ := nvml.DeviceGetCount()
	for i := 0; i < count; i++ {
		if device, ret := nvml.DeviceGetHandleByIndex(i); errors.Is(ret, nvml.SUCCESS) {
			prefix := fmt.Sprintf("nvidia_%d_", i)
			for k, v := range n.collectDeviceStatic(device) {
				info[prefix+k] = v
			}
		}
	}
	return info
}

func (n *NvidiaCollector) collectDeviceStatic(device nvml.Device) StaticMetrics {
	name, _ := device.GetName()
	uuid, _ := device.GetUUID()
	mem, _ := device.GetMemoryInfo()

	return StaticMetrics{
		"name":          name,
		"uuid":          uuid,
		"memoryTotalMb": int64(mem.Total >> 20),
	}
}

// --- Dynamic Metrics ---

func (n *NvidiaCollector) CollectDynamic() []DynamicMetrics {
	if !n.initialized {
		return nil
	}
	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) {
		return nil
	}

	gpuMetrics := make([]DynamicMetrics, 0, count)
	for i := 0; i < count; i++ {
		if device, ret := nvml.DeviceGetHandleByIndex(i); errors.Is(ret, nvml.SUCCESS) {
			gpuMetrics = append(gpuMetrics, n.collectDeviceDynamic(device, i))
		}
	}
	return gpuMetrics
}

func (n *NvidiaCollector) collectDeviceDynamic(device nvml.Device, index int) DynamicMetrics {
	util, tUtil := nvmlProbe(device.GetUtilizationRates, nvml.Utilization{})
	mem, tMem := nvmlProbe(device.GetMemoryInfo, nvml.Memory{})
	bar, tBar := nvmlProbe(device.GetBAR1MemoryInfo, nvml.BAR1Memory{})
	pwr, tPwr := nvmlProbe(device.GetPowerUsage, uint32(0))
	pstate, tPs := nvmlProbe(device.GetPerformanceState, nvml.Pstates(0))
	procs, tProc := n.getRunningProcesses(device)

	return DynamicMetrics{
		"gpuIndex":       NewMetric(int64(index)),
		"utilizationGpu": NewMetricWithTime(int64(util.Gpu), tUtil),
		"utilizationMem": NewMetricWithTime(int64(util.Memory), tUtil),
		"memoryUsedMb":   NewMetricWithTime(int64(mem.Used>>20), tMem),
		"memoryFreeMb":   NewMetricWithTime(int64(mem.Free>>20), tMem),
		"bar1UsedMb":     NewMetricWithTime(int64(bar.Bar1Used>>20), tBar),
		"powerDrawW":     NewMetricWithTime(float64(pwr)/1000.0, tPwr),
		"perfState":      NewMetricWithTime(fmt.Sprintf("P%d", int(pstate)), tPs),
		"processCount":   NewMetricWithTime(int64(len(procs)), tProc),
		"processes":      NewMetricWithTime(procs, tProc),
	}.Merge(n.collectClockMetrics(device))
}

func (n *NvidiaCollector) collectClockMetrics(device nvml.Device) DynamicMetrics {
	m := make(DynamicMetrics)
	for key, fn := range map[string]func() (uint32, nvml.Return){
		"temperatureC":     func() (uint32, nvml.Return) { return device.GetTemperature(nvml.TEMPERATURE_GPU) },
		"fanSpeed":         func() (uint32, nvml.Return) { return device.GetFanSpeed() },
		"clockGraphicsMhz": func() (uint32, nvml.Return) { return device.GetClockInfo(nvml.CLOCK_GRAPHICS) },
		"clockSmMhz":       func() (uint32, nvml.Return) { return device.GetClockInfo(nvml.CLOCK_SM) },
		"clockMemMhz":      func() (uint32, nvml.Return) { return device.GetClockInfo(nvml.CLOCK_MEM) },
		"pcieTxKbps":       func() (uint32, nvml.Return) { return device.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES) },
		"pcieRxKbps":       func() (uint32, nvml.Return) { return device.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES) },
	} {
		val, ts := nvmlProbe(fn, uint32(0))
		m[key] = NewMetricWithTime(int64(val), ts)
	}
	return m
}

func (n *NvidiaCollector) getRunningProcesses(device nvml.Device) ([]map[string]interface{}, int64) {
	seen := make(map[uint32]bool)
	var procs []map[string]interface{}

	for _, getter := range []func() ([]nvml.ProcessInfo, nvml.Return){
		device.GetComputeRunningProcesses,
		device.GetGraphicsRunningProcesses,
	} {
		if list, ret := getter(); ret == nvml.SUCCESS {
			for _, p := range list {
				if !seen[p.Pid] {
					seen[p.Pid] = true
					procs = append(procs, map[string]interface{}{
						"pid":          p.Pid,
						"name":         getProcessName(int(p.Pid)),
						"usedMemoryMb": int64(p.UsedGpuMemory >> 20),
					})
				}
			}
		}
	}
	return procs, GetTimestamp()
}

func getProcessName(pid int) string {
	if content, _ := ProbeFile(fmt.Sprintf("/proc/%d/comm", pid)); content != "" {
		return content
	}
	return "unknown"
}
