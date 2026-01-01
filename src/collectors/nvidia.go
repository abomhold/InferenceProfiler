package collectors

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NvidiaCollector handles NVIDIA GPU metrics collection
type NvidiaCollector struct {
	initialized  bool
	collectProcs bool
}

// NewNvidiaCollector creates a new NVIDIA collector
func NewNvidiaCollector(collectProcs bool) *NvidiaCollector {
	n := &NvidiaCollector{
		collectProcs: collectProcs,
	}
	err := n.Init()
	if err != nil {
		panic(err)
	}
	return n
}

// Init initializes the NVML library
func (n *NvidiaCollector) Init() error {
	if n.initialized {
		return nil
	}
	if ret := nvml.Init(); !errors.Is(ret, nvml.SUCCESS) {
		return fmt.Errorf("failed to initialize NVML: %s", nvml.ErrorString(ret))
	}
	n.initialized = true
	return nil
}

// Cleanup shuts down NVML
func (n *NvidiaCollector) Cleanup() {
	if n.initialized {
		nvml.Shutdown()
		n.initialized = false
	}
}

// CollectNvidiaStatic populates static GPU information
func (n *NvidiaCollector) CollectNvidiaStatic(m *StaticMetrics) {
	if !n.initialized {
		return
	}

	if driverVersion, ret := nvml.SystemGetDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvidiaDriverVersion = driverVersion
	}

	if cv, ret := nvml.SystemGetCudaDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvidiaCudaVersion = fmt.Sprintf("%d.%d", cv/1000, (cv%1000)/10)
	}

	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) {
		return
	}

	var gpus []NvidiaGPUStatic
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if !errors.Is(ret, nvml.SUCCESS) {
			continue
		}

		gpu := NvidiaGPUStatic{Index: i}

		if name, ret := device.GetName(); errors.Is(ret, nvml.SUCCESS) {
			gpu.Name = name
		}
		if uuid, ret := device.GetUUID(); errors.Is(ret, nvml.SUCCESS) {
			gpu.UUID = uuid
		}
		if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
			gpu.TotalMemoryMb = int64(mem.Total >> 20)
		}

		gpus = append(gpus, gpu)
	}

	if len(gpus) > 0 {
		if data, err := json.Marshal(gpus); err == nil {
			m.NvidiaGPUsJSON = string(data)
		}
	}
}

// CollectNvidiaDynamic populates dynamic GPU metrics into the slice
func (n *NvidiaCollector) CollectNvidiaDynamic(m *DynamicMetrics) {
	if !n.initialized {
		return
	}

	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) {
		return
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if !errors.Is(ret, nvml.SUCCESS) {
			continue
		}
		gpu := n.collectDeviceDynamic(device, i)
		m.NvidiaGPUs = append(m.NvidiaGPUs, gpu)
	}
}

func (n *NvidiaCollector) collectDeviceDynamic(device nvml.Device, index int) NvidiaGPUDynamic {
	gpu := NvidiaGPUDynamic{Index: index}

	// Utilization
	if util, ret := device.GetUtilizationRates(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.UtilizationGPU = int64(util.Gpu)
		gpu.UtilizationGPUT = ts
		gpu.UtilizationMem = int64(util.Memory)
		gpu.UtilizationMemT = ts
	}

	// Memory
	if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.MemoryUsedMb = int64(mem.Used >> 20)
		gpu.MemoryUsedMbT = ts
		gpu.MemoryFreeMb = int64(mem.Free >> 20)
		gpu.MemoryFreeMbT = ts
	}

	// BAR1 Memory
	if bar, ret := device.GetBAR1MemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.Bar1UsedMb = int64(bar.Bar1Used >> 20)
		gpu.Bar1UsedMbT = ts
	}

	// Temperature
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.TemperatureC = int64(temp)
		gpu.TemperatureCT = ts
	}

	// Fan speed
	if fan, ret := device.GetFanSpeed(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.FanSpeed = int64(fan)
		gpu.FanSpeedT = ts
	}

	// Clock speeds
	if clock, ret := device.GetClockInfo(nvml.CLOCK_GRAPHICS); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ClockGraphicsMhz = int64(clock)
		gpu.ClockGraphicsMhzT = ts
	}
	if clock, ret := device.GetClockInfo(nvml.CLOCK_SM); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ClockSmMhz = int64(clock)
		gpu.ClockSmMhzT = ts
	}
	if clock, ret := device.GetClockInfo(nvml.CLOCK_MEM); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ClockMemMhz = int64(clock)
		gpu.ClockMemMhzT = ts
	}

	// PCIe throughput
	if tx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PcieTxKbps = int64(tx)
		gpu.PcieTxKbpsT = ts
	}
	if rx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PcieRxKbps = int64(rx)
		gpu.PcieRxKbpsT = ts
	}

	// Power
	if pwr, ret := device.GetPowerUsage(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PowerDrawW = float64(pwr) / 1000.0
		gpu.PowerDrawWT = ts
	}

	// Performance state
	if pstate, ret := device.GetPerformanceState(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PerfState = fmt.Sprintf("P%d", int(pstate))
		gpu.PerfStateT = ts
	}

	// Running processes - only if enabled
	if n.collectProcs {
		procs, ts := n.getRunningProcesses(device)
		gpu.ProcessCount = int64(len(procs))
		gpu.ProcessCountT = ts
		if len(procs) > 0 {
			if data, err := json.Marshal(procs); err == nil {
				gpu.ProcessesJSON = string(data)
			}
		}
	}

	return gpu
}

func (n *NvidiaCollector) getRunningProcesses(device nvml.Device) ([]GPUProcess, int64) {
	seen := make(map[uint32]bool)
	var procs []GPUProcess

	for _, getter := range []func() ([]nvml.ProcessInfo, nvml.Return){
		device.GetComputeRunningProcesses,
		device.GetGraphicsRunningProcesses,
	} {
		if list, ret := getter(); errors.Is(ret, nvml.SUCCESS) {
			for _, p := range list {
				if !seen[p.Pid] {
					seen[p.Pid] = true
					procs = append(procs, GPUProcess{
						PID:          p.Pid,
						Name:         getProcessName(int(p.Pid)),
						UsedMemoryMb: int64(p.UsedGpuMemory >> 20),
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
