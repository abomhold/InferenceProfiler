package collectors

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type NvidiaCollector struct {
	initialized  bool
	collectProcs bool
	devices      []nvml.Device
}

func NewNvidiaCollector(collectProcs bool) *NvidiaCollector {
	n := &NvidiaCollector{collectProcs: collectProcs}
	if err := n.init(); err != nil {
		log.Printf("WARNING: NVIDIA collector disabled: %v", err)
		return nil
	}
	return n
}

func (n *NvidiaCollector) Name() string { return "NVIDIA" }

func (n *NvidiaCollector) init() error {
	if n.initialized {
		return nil
	}
	if ret := nvml.Init(); !errors.Is(ret, nvml.SUCCESS) {
		return fmt.Errorf("failed to initialize NVML: %s", nvml.ErrorString(ret))
	}

	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) || count == 0 {
		nvml.Shutdown()
		return fmt.Errorf("no NVIDIA devices found")
	}

	n.devices = make([]nvml.Device, count)
	for i := 0; i < count; i++ {
		n.devices[i], _ = nvml.DeviceGetHandleByIndex(i)
	}

	n.initialized = true
	return nil
}

func (n *NvidiaCollector) Close() error {
	if n.initialized {
		nvml.Shutdown()
		n.initialized = false
	}
	return nil
}

func (n *NvidiaCollector) CollectStatic(m *StaticMetrics) {
	if !n.initialized {
		return
	}

	m.NvidiaGPUCount = len(n.devices)

	if driverVersion, ret := nvml.SystemGetDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvidiaDriverVersion = driverVersion
	}
	if cudaVersion, ret := nvml.SystemGetCudaDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvidiaCudaVersion = fmt.Sprintf("%d.%d", cudaVersion/1000, (cudaVersion%1000)/10)
	}
	if nvmlVersion, ret := nvml.SystemGetNVMLVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvmlVersion = nvmlVersion
	}

	gpus := make([]GPUInfo, 0, len(n.devices))
	for i, device := range n.devices {
		gpus = append(gpus, n.collectDeviceStatic(device, i))
	}

	if len(gpus) > 0 {
		data, _ := json.Marshal(gpus)
		m.NvidiaGPUsJSON = string(data)
	}
}

func (n *NvidiaCollector) collectDeviceStatic(device nvml.Device, index int) GPUInfo {
	gpu := GPUInfo{Index: index}

	if name, ret := device.GetName(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Name = name
	}
	if uuid, ret := device.GetUUID(); errors.Is(ret, nvml.SUCCESS) {
		gpu.UUID = uuid
	}
	if serial, ret := device.GetSerial(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Serial = serial
	}
	if partNum, ret := device.GetBoardPartNumber(); errors.Is(ret, nvml.SUCCESS) {
		gpu.BoardPartNumber = partNum
	}
	if brand, ret := device.GetBrand(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Brand = brandToString(brand)
	}
	if arch, ret := device.GetArchitecture(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Architecture = archToString(arch)
	}
	if major, minor, ret := device.GetCudaComputeCapability(); errors.Is(ret, nvml.SUCCESS) {
		gpu.CudaCapabilityMajor = major
		gpu.CudaCapabilityMinor = minor
	}
	if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MemoryTotalBytes = int64(mem.Total)
	}
	if bar1, ret := device.GetBAR1MemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Bar1TotalBytes = int64(bar1.Bar1Total)
	}
	if busWidth, ret := device.GetMemoryBusWidth(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MemoryBusWidthBits = int(busWidth)
	}
	if numCores, ret := device.GetNumGpuCores(); errors.Is(ret, nvml.SUCCESS) {
		gpu.NumCores = int(numCores)
	}
	if maxGfx, ret := device.GetMaxClockInfo(nvml.CLOCK_GRAPHICS); errors.Is(ret, nvml.SUCCESS) {
		gpu.MaxClockGraphicsMhz = int(maxGfx)
	}
	if maxMem, ret := device.GetMaxClockInfo(nvml.CLOCK_MEM); errors.Is(ret, nvml.SUCCESS) {
		gpu.MaxClockMemoryMhz = int(maxMem)
	}
	if maxSm, ret := device.GetMaxClockInfo(nvml.CLOCK_SM); errors.Is(ret, nvml.SUCCESS) {
		gpu.MaxClockSmMhz = int(maxSm)
	}
	if maxVideo, ret := device.GetMaxClockInfo(nvml.CLOCK_VIDEO); errors.Is(ret, nvml.SUCCESS) {
		gpu.MaxClockVideoMhz = int(maxVideo)
	}
	if pci, ret := device.GetPciInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PciBusId = int8SliceToString(pci.BusId[:])
		gpu.PciDeviceId = pci.PciDeviceId
		gpu.PciSubsystemId = pci.PciSubSystemId
	}
	if maxGen, ret := device.GetMaxPcieLinkGeneration(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PcieMaxLinkGen = int(maxGen)
	}
	if maxWidth, ret := device.GetMaxPcieLinkWidth(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PcieMaxLinkWidth = int(maxWidth)
	}
	if defaultLimit, ret := device.GetPowerManagementDefaultLimit(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerDefaultLimitMw = int(defaultLimit)
	}
	if minLimit, maxLimit, ret := device.GetPowerManagementLimitConstraints(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerMinLimitMw = int(minLimit)
		gpu.PowerMaxLimitMw = int(maxLimit)
	}
	if vbios, ret := device.GetVbiosVersion(); errors.Is(ret, nvml.SUCCESS) {
		gpu.VbiosVersion = vbios
	}
	if inforomImg, ret := device.GetInforomImageVersion(); errors.Is(ret, nvml.SUCCESS) {
		gpu.InforomImageVersion = inforomImg
	}
	if inforomOem, ret := device.GetInforomVersion(nvml.INFOROM_OEM); errors.Is(ret, nvml.SUCCESS) {
		gpu.InforomOemVersion = inforomOem
	}
	if numFans, ret := device.GetNumFans(); errors.Is(ret, nvml.SUCCESS) {
		gpu.NumFans = int(numFans)
	}
	if temp, ret := device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SHUTDOWN); errors.Is(ret, nvml.SUCCESS) {
		gpu.TempShutdownC = int(temp)
	}
	if temp, ret := device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SLOWDOWN); errors.Is(ret, nvml.SUCCESS) {
		gpu.TempSlowdownC = int(temp)
	}
	if temp, ret := device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_GPU_MAX); errors.Is(ret, nvml.SUCCESS) {
		gpu.TempMaxOperatingC = int(temp)
	}
	if temp, ret := device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_ACOUSTIC_CURR); errors.Is(ret, nvml.SUCCESS) {
		gpu.TempTargetC = int(temp)
	}
	if eccCurrent, _, ret := device.GetEccMode(); errors.Is(ret, nvml.SUCCESS) {
		gpu.EccModeEnabled = eccCurrent == nvml.FEATURE_ENABLED
	}
	if persistence, ret := device.GetPersistenceMode(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PersistenceModeOn = persistence == nvml.FEATURE_ENABLED
	}
	if computeMode, ret := device.GetComputeMode(); errors.Is(ret, nvml.SUCCESS) {
		gpu.ComputeMode = computeModeToString(computeMode)
	}
	if multiGpu, ret := device.GetMultiGpuBoard(); errors.Is(ret, nvml.SUCCESS) {
		gpu.IsMultiGpuBoard = multiGpu != 0
	}
	if displayMode, ret := device.GetDisplayMode(); errors.Is(ret, nvml.SUCCESS) {
		gpu.DisplayModeEnabled = displayMode == nvml.FEATURE_ENABLED
	}
	if displayActive, ret := device.GetDisplayActive(); errors.Is(ret, nvml.SUCCESS) {
		gpu.DisplayActive = displayActive == nvml.FEATURE_ENABLED
	}
	if migCurrent, _, ret := device.GetMigMode(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MigModeEnabled = migCurrent == nvml.DEVICE_MIG_ENABLE
	}
	if capa, ret := device.GetEncoderCapacity(nvml.ENCODER_QUERY_H264); errors.Is(ret, nvml.SUCCESS) {
		gpu.EncoderCapacityH264 = capa
	}
	if capa, ret := device.GetEncoderCapacity(nvml.ENCODER_QUERY_HEVC); errors.Is(ret, nvml.SUCCESS) {
		gpu.EncoderCapacityHEVC = capa
	}
	if capa, ret := device.GetEncoderCapacity(nvml.ENCODER_QUERY_AV1); errors.Is(ret, nvml.SUCCESS) {
		gpu.EncoderCapacityAV1 = capa
	}

	nvlinkCount := 0
	for link := 0; link < config.MaxNvLinks; link++ {
		if _, ret := device.GetNvLinkState(link); errors.Is(ret, nvml.SUCCESS) {
			nvlinkCount++
		} else {
			break
		}
	}
	gpu.NvLinkCount = nvlinkCount

	return gpu
}

func (n *NvidiaCollector) CollectDynamic(m *DynamicMetrics) {
	if !n.initialized {
		return
	}

	gpus := make([]GPUDynamic, 0, len(n.devices))
	for i, device := range n.devices {
		gpus = append(gpus, n.collectDeviceDynamic(device, i))
	}
	m.GPUs = gpus
}

func (n *NvidiaCollector) collectDeviceDynamic(device nvml.Device, index int) GPUDynamic {
	gpu := GPUDynamic{Index: index}
	ts := time.Now().UnixNano()

	if util, ret := device.GetUtilizationRates(); errors.Is(ret, nvml.SUCCESS) {
		gpu.UtilizationGPU, gpu.UtilizationGPUT = int64(util.Gpu), ts
		gpu.UtilizationMemory, gpu.UtilizationMemoryT = int64(util.Memory), ts
	}
	if util, period, ret := device.GetEncoderUtilization(); errors.Is(ret, nvml.SUCCESS) {
		gpu.UtilizationEncoder, gpu.UtilizationEncoderT = int64(util), ts
		gpu.EncoderSamplingPeriodUs = int64(period)
	}
	if util, period, ret := device.GetDecoderUtilization(); errors.Is(ret, nvml.SUCCESS) {
		gpu.UtilizationDecoder, gpu.UtilizationDecoderT = int64(util), ts
		gpu.DecoderSamplingPeriodUs = int64(period)
	}
	if util, _, ret := device.GetJpgUtilization(); errors.Is(ret, nvml.SUCCESS) {
		gpu.UtilizationJpeg, gpu.UtilizationJpegT = int64(util), ts
	}
	if util, _, ret := device.GetOfaUtilization(); errors.Is(ret, nvml.SUCCESS) {
		gpu.UtilizationOfa, gpu.UtilizationOfaT = int64(util), ts
	}
	if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MemoryUsedBytes, gpu.MemoryUsedBytesT = int64(mem.Used), ts
		gpu.MemoryFreeBytes, gpu.MemoryFreeBytesT = int64(mem.Free), ts
		gpu.MemoryTotalBytes = int64(mem.Total)
	}
	if mem, ret := device.GetMemoryInfo_v2(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MemoryReservedBytes, gpu.MemoryReservedBytesT = int64(mem.Reserved), ts
	}
	if bar1, ret := device.GetBAR1MemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Bar1UsedBytes, gpu.Bar1UsedBytesT = int64(bar1.Bar1Used), ts
		gpu.Bar1FreeBytes, gpu.Bar1FreeBytesT = int64(bar1.Bar1Free), ts
		gpu.Bar1TotalBytes = int64(bar1.Bar1Total)
	}
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); errors.Is(ret, nvml.SUCCESS) {
		gpu.TemperatureGpuC, gpu.TemperatureGpuCT = int64(temp), ts
	}
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_COUNT); errors.Is(ret, nvml.SUCCESS) {
		gpu.TemperatureMemoryC, gpu.TemperatureMemoryCT = int64(temp), ts
	}
	if fan, ret := device.GetFanSpeed(); errors.Is(ret, nvml.SUCCESS) {
		gpu.FanSpeedPercent, gpu.FanSpeedPercentT = int64(fan), ts
	}
	if numFans, ret := device.GetNumFans(); errors.Is(ret, nvml.SUCCESS) && numFans > 1 {
		var speeds []int
		for f := 0; f < int(numFans); f++ {
			if speed, ret := device.GetFanSpeed_v2(f); errors.Is(ret, nvml.SUCCESS) {
				speeds = append(speeds, int(speed))
			}
		}
		if len(speeds) > 0 {
			data, _ := json.Marshal(speeds)
			gpu.FanSpeedsJSON = string(data)
		}
	}
	if clock, ret := device.GetClockInfo(nvml.CLOCK_GRAPHICS); errors.Is(ret, nvml.SUCCESS) {
		gpu.ClockGraphicsMhz, gpu.ClockGraphicsMhzT = int64(clock), ts
	}
	if clock, ret := device.GetClockInfo(nvml.CLOCK_SM); errors.Is(ret, nvml.SUCCESS) {
		gpu.ClockSmMhz, gpu.ClockSmMhzT = int64(clock), ts
	}
	if clock, ret := device.GetClockInfo(nvml.CLOCK_MEM); errors.Is(ret, nvml.SUCCESS) {
		gpu.ClockMemoryMhz, gpu.ClockMemoryMhzT = int64(clock), ts
	}
	if clock, ret := device.GetClockInfo(nvml.CLOCK_VIDEO); errors.Is(ret, nvml.SUCCESS) {
		gpu.ClockVideoMhz, gpu.ClockVideoMhzT = int64(clock), ts
	}
	if pstate, ret := device.GetPerformanceState(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PerformanceState, gpu.PerformanceStateT = int(pstate), ts
	}
	if power, ret := device.GetPowerUsage(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerUsageMw, gpu.PowerUsageMwT = int64(power), ts
	}
	if limit, ret := device.GetPowerManagementLimit(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerLimitMw, gpu.PowerLimitMwT = int64(limit), ts
	}
	if enforced, ret := device.GetEnforcedPowerLimit(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerEnforcedLimitMw, gpu.PowerEnforcedLimitMwT = int64(enforced), ts
	}
	if energy, ret := device.GetTotalEnergyConsumption(); errors.Is(ret, nvml.SUCCESS) {
		gpu.EnergyConsumptionMj, gpu.EnergyConsumptionMjT = int64(energy), ts
	}
	if tx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); errors.Is(ret, nvml.SUCCESS) {
		gpu.PcieTxBytesPerSec, gpu.PcieTxBytesPerSecT = int64(tx)*1000, ts
	}
	if rx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); errors.Is(ret, nvml.SUCCESS) {
		gpu.PcieRxBytesPerSec, gpu.PcieRxBytesPerSecT = int64(rx)*1000, ts
	}
	if gen, ret := device.GetCurrPcieLinkGeneration(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PcieCurrentLinkGen, gpu.PcieCurrentLinkGenT = int(gen), ts
	}
	if width, ret := device.GetCurrPcieLinkWidth(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PcieCurrentLinkWidth, gpu.PcieCurrentLinkWidthT = int(width), ts
	}
	if replay, ret := device.GetPcieReplayCounter(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PcieReplayCounter, gpu.PcieReplayCounterT = int64(replay), ts
	}
	if reasons, ret := device.GetCurrentClocksEventReasons(); errors.Is(ret, nvml.SUCCESS) {
		gpu.ClocksEventReasons, gpu.ClocksEventReasonsT = reasons, ts
	}

	n.collectViolationStatus(device, &gpu, ts)
	n.collectEccErrors(device, &gpu, ts)
	n.collectNvLinkMetrics(device, &gpu)

	if n.collectProcs {
		n.collectGPUProcesses(device, &gpu, ts)
	}

	return gpu
}

func (n *NvidiaCollector) collectViolationStatus(device nvml.Device, gpu *GPUDynamic, ts int64) {
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_POWER); errors.Is(ret, nvml.SUCCESS) {
		gpu.ViolationPowerNs, gpu.ViolationPowerNsT = int64(viol.ViolationTime), ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_THERMAL); errors.Is(ret, nvml.SUCCESS) {
		gpu.ViolationThermalNs, gpu.ViolationThermalNsT = int64(viol.ViolationTime), ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_RELIABILITY); errors.Is(ret, nvml.SUCCESS) {
		gpu.ViolationReliabilityNs, gpu.ViolationReliabilityNsT = int64(viol.ViolationTime), ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_BOARD_LIMIT); errors.Is(ret, nvml.SUCCESS) {
		gpu.ViolationBoardLimitNs, gpu.ViolationBoardLimitNsT = int64(viol.ViolationTime), ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_LOW_UTILIZATION); errors.Is(ret, nvml.SUCCESS) {
		gpu.ViolationLowUtilNs, gpu.ViolationLowUtilNsT = int64(viol.ViolationTime), ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_SYNC_BOOST); errors.Is(ret, nvml.SUCCESS) {
		gpu.ViolationSyncBoostNs, gpu.ViolationSyncBoostNsT = int64(viol.ViolationTime), ts
	}
}

func (n *NvidiaCollector) collectEccErrors(device nvml.Device, gpu *GPUDynamic, ts int64) {
	if count, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.AGGREGATE_ECC); errors.Is(ret, nvml.SUCCESS) {
		gpu.EccAggregateSbe, gpu.EccAggregateSbeT = int64(count), ts
	}
	if count, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_UNCORRECTED, nvml.AGGREGATE_ECC); errors.Is(ret, nvml.SUCCESS) {
		gpu.EccAggregateDbe, gpu.EccAggregateDbeT = int64(count), ts
	}
	if sbe, ret := device.GetRetiredPages(nvml.PAGE_RETIREMENT_CAUSE_MULTIPLE_SINGLE_BIT_ECC_ERRORS); errors.Is(ret, nvml.SUCCESS) {
		gpu.RetiredPagesSbe, gpu.RetiredPagesT = int64(len(sbe)), ts
	}
	if dbe, ret := device.GetRetiredPages(nvml.PAGE_RETIREMENT_CAUSE_DOUBLE_BIT_ECC_ERROR); errors.Is(ret, nvml.SUCCESS) {
		gpu.RetiredPagesDbe, gpu.RetiredPagesT = int64(len(dbe)), ts
	}
	if pending, ret := device.GetRetiredPagesPendingStatus(); errors.Is(ret, nvml.SUCCESS) {
		gpu.RetiredPending, gpu.RetiredPendingT = pending == nvml.FEATURE_ENABLED, ts
	}
	if correctable, uncorrectable, pending, failure, ret := device.GetRemappedRows(); errors.Is(ret, nvml.SUCCESS) {
		gpu.RemappedRowsCorrectable = int64(correctable)
		gpu.RemappedRowsUncorrectable = int64(uncorrectable)
		gpu.RemappedRowsPending = pending
		gpu.RemappedRowsFailure = failure
		gpu.RemappedRowsT = ts
	}
}

func (n *NvidiaCollector) collectNvLinkMetrics(device nvml.Device, gpu *GPUDynamic) {
	var bandwidths []NvLinkBandwidth
	var linkErrors []NvLinkErrors

	for link := 0; link < config.MaxNvLinks; link++ {
		state, ret := device.GetNvLinkState(link)
		if !errors.Is(ret, nvml.SUCCESS) || state != nvml.FEATURE_ENABLED {
			break
		}

		bw := NvLinkBandwidth{Link: link}
		if rx, tx, ret := device.GetNvLinkUtilizationCounter(link, 0); errors.Is(ret, nvml.SUCCESS) {
			bw.TxBytes = int64(tx)
			bw.RxBytes = int64(rx)
		}
		bandwidths = append(bandwidths, bw)

		errs := NvLinkErrors{Link: link}
		if crc, ret := device.GetNvLinkErrorCounter(link, nvml.NVLINK_ERROR_DL_CRC_FLIT); errors.Is(ret, nvml.SUCCESS) {
			errs.CrcErrors = int64(crc)
		}
		if ecc, ret := device.GetNvLinkErrorCounter(link, nvml.NVLINK_ERROR_DL_ECC_DATA); errors.Is(ret, nvml.SUCCESS) {
			errs.EccErrors = int64(ecc)
		}
		if replay, ret := device.GetNvLinkErrorCounter(link, nvml.NVLINK_ERROR_DL_REPLAY); errors.Is(ret, nvml.SUCCESS) {
			errs.ReplayErrors = int64(replay)
		}
		if recovery, ret := device.GetNvLinkErrorCounter(link, nvml.NVLINK_ERROR_DL_RECOVERY); errors.Is(ret, nvml.SUCCESS) {
			errs.RecoveryCount = int64(recovery)
		}
		linkErrors = append(linkErrors, errs)
	}

	if len(bandwidths) > 0 {
		data, _ := json.Marshal(bandwidths)
		gpu.NvLinkBandwidthJSON = string(data)
	}
	if len(linkErrors) > 0 {
		data, _ := json.Marshal(linkErrors)
		gpu.NvLinkErrorsJSON = string(data)
	}
}

func (n *NvidiaCollector) collectGPUProcesses(device nvml.Device, gpu *GPUDynamic, ts int64) {
	seen := make(map[uint32]bool)
	var procs []GPUProcess

	if list, ret := device.GetComputeRunningProcesses(); errors.Is(ret, nvml.SUCCESS) {
		for _, p := range list {
			if !seen[p.Pid] {
				seen[p.Pid] = true
				procs = append(procs, GPUProcess{
					PID:             p.Pid,
					Name:            getProcessName(int(p.Pid)),
					UsedMemoryBytes: int64(p.UsedGpuMemory),
					Type:            "compute",
				})
			}
		}
	}
	if list, ret := device.GetGraphicsRunningProcesses(); errors.Is(ret, nvml.SUCCESS) {
		for _, p := range list {
			if !seen[p.Pid] {
				seen[p.Pid] = true
				procs = append(procs, GPUProcess{
					PID:             p.Pid,
					Name:            getProcessName(int(p.Pid)),
					UsedMemoryBytes: int64(p.UsedGpuMemory),
					Type:            "graphics",
				})
			}
		}
	}
	if list, ret := device.GetMPSComputeRunningProcesses(); errors.Is(ret, nvml.SUCCESS) {
		for _, p := range list {
			if !seen[p.Pid] {
				seen[p.Pid] = true
				procs = append(procs, GPUProcess{
					PID:             p.Pid,
					Name:            getProcessName(int(p.Pid)),
					UsedMemoryBytes: int64(p.UsedGpuMemory),
					Type:            "mps",
				})
			}
		}
	}

	gpu.ProcessCount, gpu.ProcessCountT = int64(len(procs)), ts
	if len(procs) > 0 {
		data, _ := json.Marshal(procs)
		gpu.ProcessesJSON = string(data)
	}

	if samples, ret := device.GetProcessUtilization(0); errors.Is(ret, nvml.SUCCESS) {
		var result []GPUProcessUtilization
		for _, s := range samples {
			result = append(result, GPUProcessUtilization{
				PID:         s.Pid,
				SmUtil:      int(s.SmUtil),
				MemUtil:     int(s.MemUtil),
				EncUtil:     int(s.EncUtil),
				DecUtil:     int(s.DecUtil),
				TimestampUs: int64(s.TimeStamp),
			})
		}
		if len(result) > 0 {
			data, _ := json.Marshal(result)
			gpu.ProcessUtilizationJSON = string(data)
		}
	}
}

func getProcessName(pid int) string {
	val, _, _ := probing.File(fmt.Sprintf("/proc/%d/comm", pid))
	return val
}

func archToString(arch nvml.DeviceArchitecture) string {
	archMap := map[nvml.DeviceArchitecture]string{
		nvml.DEVICE_ARCH_KEPLER:  "Kepler",
		nvml.DEVICE_ARCH_MAXWELL: "Maxwell",
		nvml.DEVICE_ARCH_PASCAL:  "Pascal",
		nvml.DEVICE_ARCH_VOLTA:   "Volta",
		nvml.DEVICE_ARCH_TURING:  "Turing",
		nvml.DEVICE_ARCH_AMPERE:  "Ampere",
		nvml.DEVICE_ARCH_ADA:     "Ada",
		nvml.DEVICE_ARCH_HOPPER:  "Hopper",
	}
	if name, ok := archMap[arch]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", arch)
}

func brandToString(brand nvml.BrandType) string {
	brandMap := map[nvml.BrandType]string{
		nvml.BRAND_UNKNOWN:             "Unknown",
		nvml.BRAND_QUADRO:              "Quadro",
		nvml.BRAND_TESLA:               "Tesla",
		nvml.BRAND_NVS:                 "NVS",
		nvml.BRAND_GRID:                "GRID",
		nvml.BRAND_GEFORCE:             "GeForce",
		nvml.BRAND_TITAN:               "Titan",
		nvml.BRAND_NVIDIA_VAPPS:        "vApps",
		nvml.BRAND_NVIDIA_VPC:          "VPC",
		nvml.BRAND_NVIDIA_VCS:          "VCS",
		nvml.BRAND_NVIDIA_VWS:          "VWS",
		nvml.BRAND_NVIDIA_CLOUD_GAMING: "CloudGaming",
		nvml.BRAND_QUADRO_RTX:          "QuadroRTX",
		nvml.BRAND_NVIDIA_RTX:          "NvidiaRTX",
		nvml.BRAND_NVIDIA:              "Nvidia",
		nvml.BRAND_GEFORCE_RTX:         "GeForceRTX",
		nvml.BRAND_TITAN_RTX:           "TitanRTX",
	}
	if name, ok := brandMap[brand]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", brand)
}

func computeModeToString(mode nvml.ComputeMode) string {
	modeMap := map[nvml.ComputeMode]string{
		nvml.COMPUTEMODE_DEFAULT:           "Default",
		nvml.COMPUTEMODE_EXCLUSIVE_THREAD:  "ExclusiveThread",
		nvml.COMPUTEMODE_PROHIBITED:        "Prohibited",
		nvml.COMPUTEMODE_EXCLUSIVE_PROCESS: "ExclusiveProcess",
	}
	if name, ok := modeMap[mode]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", mode)
}

func int8SliceToString(b []int8) string {
	var buf []byte
	for _, c := range b {
		if c == 0 {
			break
		}
		buf = append(buf, byte(c))
	}
	return string(buf)
}
