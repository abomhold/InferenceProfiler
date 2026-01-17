package collectors

import (
	"InferenceProfiler/src/collectors/types"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NvidiaCollector handles NVIDIA GPU metrics collection via NVML
type NvidiaCollector struct {
	initialized  bool
	collectProcs bool
}

// NewNvidiaCollector creates a new NVIDIA collector instance
// Returns nil if NVML initialization fails
func NewNvidiaCollector(collectProcs bool) *NvidiaCollector {
	n := &NvidiaCollector{
		collectProcs: collectProcs,
	}
	if err := n.init(); err != nil {
		log.Printf("WARNING: NVIDIA collector disabled: %v", err)
		return nil
	}
	return n
}

func (n *NvidiaCollector) Name() string {
	return "NVIDIA"
}

func (n *NvidiaCollector) init() error {
	if n.initialized {
		return nil
	}
	if ret := nvml.Init(); !errors.Is(ret, nvml.SUCCESS) {
		return fmt.Errorf("failed to initialize NVML: %s", nvml.ErrorString(ret))
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

// =============================================================================
// Static Collection
// =============================================================================

func (n *NvidiaCollector) CollectStatic(m *types.StaticMetrics) {
	if !n.initialized {
		return
	}

	// System-level info
	if driverVersion, ret := nvml.SystemGetDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvidiaDriverVersion = driverVersion
	}
	if cudaVersion, ret := nvml.SystemGetCudaDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvidiaCudaVersion = fmt.Sprintf("%d.%d", cudaVersion/1000, (cudaVersion%1000)/10)
	}
	if nvmlVersion, ret := nvml.SystemGetNVMLVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvmlVersion = nvmlVersion
	}

	// Device enumeration
	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) {
		return
	}
	m.NvidiaGPUCount = count

	var gpus []types.NvidiaGPUStatic
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if !errors.Is(ret, nvml.SUCCESS) {
			continue
		}
		gpus = append(gpus, n.collectDeviceStatic(device, i))
	}

	if len(gpus) > 0 {
		if data, err := json.Marshal(gpus); err == nil {
			m.NvidiaGPUsJSON = string(data)
		}
	}
}

func (n *NvidiaCollector) collectDeviceStatic(device nvml.Device, index int) types.NvidiaGPUStatic {
	gpu := types.NvidiaGPUStatic{Index: index}

	// Basic identification
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

	// Architecture
	if arch, ret := device.GetArchitecture(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Architecture = archToString(arch)
	}
	if major, minor, ret := device.GetCudaComputeCapability(); errors.Is(ret, nvml.SUCCESS) {
		gpu.CudaCapabilityMajor = major
		gpu.CudaCapabilityMinor = minor
	}

	// Memory
	if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MemoryTotalBytes = int64(mem.Total)
	}
	if bar1, ret := device.GetBAR1MemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Bar1TotalBytes = int64(bar1.Bar1Total)
	}
	if busWidth, ret := device.GetMemoryBusWidth(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MemoryBusWidthBits = int(busWidth)
	}

	// Compute
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

	// PCI
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

	// Power
	if defaultLimit, ret := device.GetPowerManagementDefaultLimit(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerDefaultLimitMw = int(defaultLimit)
	}
	if minLimit, maxLimit, ret := device.GetPowerManagementLimitConstraints(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerMinLimitMw = int(minLimit)
		gpu.PowerMaxLimitMw = int(maxLimit)
	}

	// Firmware
	if vbios, ret := device.GetVbiosVersion(); errors.Is(ret, nvml.SUCCESS) {
		gpu.VbiosVersion = vbios
	}
	if inforomImg, ret := device.GetInforomImageVersion(); errors.Is(ret, nvml.SUCCESS) {
		gpu.InforomImageVersion = inforomImg
	}
	if inforomOem, ret := device.GetInforomVersion(nvml.INFOROM_OEM); errors.Is(ret, nvml.SUCCESS) {
		gpu.InforomOemVersion = inforomOem
	}

	// Thermal
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

	// Configuration
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

	// MIG
	if migCurrent, _, ret := device.GetMigMode(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MigModeEnabled = migCurrent == nvml.DEVICE_MIG_ENABLE
	}

	// Encoder
	if cap, ret := device.GetEncoderCapacity(nvml.ENCODER_QUERY_H264); errors.Is(ret, nvml.SUCCESS) {
		gpu.EncoderCapacityH264 = int(cap)
	}
	if cap, ret := device.GetEncoderCapacity(nvml.ENCODER_QUERY_HEVC); errors.Is(ret, nvml.SUCCESS) {
		gpu.EncoderCapacityHEVC = int(cap)
	}
	if cap, ret := device.GetEncoderCapacity(nvml.ENCODER_QUERY_AV1); errors.Is(ret, nvml.SUCCESS) {
		gpu.EncoderCapacityAV1 = int(cap)
	}

	// NVLink count
	nvlinkCount := 0
	for link := 0; link < MaxNvLinks; link++ {
		if _, ret := device.GetNvLinkState(link); errors.Is(ret, nvml.SUCCESS) {
			nvlinkCount++
		} else {
			break
		}
	}
	gpu.NvLinkCount = nvlinkCount

	return gpu
}

// =============================================================================
// Dynamic Collection
// =============================================================================

func (n *NvidiaCollector) CollectDynamic(m *types.DynamicMetrics) {
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

func (n *NvidiaCollector) collectDeviceDynamic(device nvml.Device, index int) types.NvidiaGPUDynamic {
	gpu := types.NvidiaGPUDynamic{Index: index}

	// Utilization
	if util, ret := device.GetUtilizationRates(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.UtilizationGPU = int64(util.Gpu)
		gpu.UtilizationGPUT = ts
		gpu.UtilizationMemory = int64(util.Memory)
		gpu.UtilizationMemoryT = ts
	}

	if util, period, ret := device.GetEncoderUtilization(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.UtilizationEncoder = int64(util)
		gpu.UtilizationEncoderT = ts
		gpu.EncoderSamplingPeriodUs = int64(period)
	}

	if util, period, ret := device.GetDecoderUtilization(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.UtilizationDecoder = int64(util)
		gpu.UtilizationDecoderT = ts
		gpu.DecoderSamplingPeriodUs = int64(period)
	}

	if util, _, ret := device.GetJpgUtilization(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.UtilizationJpeg = int64(util)
		gpu.UtilizationJpegT = ts
	}

	if util, _, ret := device.GetOfaUtilization(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.UtilizationOfa = int64(util)
		gpu.UtilizationOfaT = ts
	}

	// Memory
	if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.MemoryUsedBytes = int64(mem.Used)
		gpu.MemoryUsedBytesT = ts
		gpu.MemoryFreeBytes = int64(mem.Free)
		gpu.MemoryFreeBytesT = ts
		gpu.MemoryTotalBytes = int64(mem.Total)
	}

	if mem, ret := device.GetMemoryInfo_v2(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.MemoryReservedBytes = int64(mem.Reserved)
		gpu.MemoryReservedBytesT = ts
	}

	if bar1, ret := device.GetBAR1MemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.Bar1UsedBytes = int64(bar1.Bar1Used)
		gpu.Bar1UsedBytesT = ts
		gpu.Bar1FreeBytes = int64(bar1.Bar1Free)
		gpu.Bar1FreeBytesT = ts
		gpu.Bar1TotalBytes = int64(bar1.Bar1Total)
	}

	// Temperature
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.TemperatureGpuC = int64(temp)
		gpu.TemperatureGpuCT = ts
	}

	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_COUNT); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.TemperatureMemoryC = int64(temp)
		gpu.TemperatureMemoryCT = ts
	}

	// Fan
	if fan, ret := device.GetFanSpeed(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.FanSpeedPercent = int64(fan)
		gpu.FanSpeedPercentT = ts
	}

	if numFans, ret := device.GetNumFans(); errors.Is(ret, nvml.SUCCESS) && numFans > 1 {
		var speeds []int
		for f := 0; f < int(numFans); f++ {
			if speed, ret := device.GetFanSpeed_v2(f); errors.Is(ret, nvml.SUCCESS) {
				speeds = append(speeds, int(speed))
			}
		}
		if len(speeds) > 0 {
			if data, err := json.Marshal(speeds); err == nil {
				gpu.FanSpeedsJSON = string(data)
			}
		}
	}

	// Clocks
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
		gpu.ClockMemoryMhz = int64(clock)
		gpu.ClockMemoryMhzT = ts
	}
	if clock, ret := device.GetClockInfo(nvml.CLOCK_VIDEO); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ClockVideoMhz = int64(clock)
		gpu.ClockVideoMhzT = ts
	}

	if gfxClock, ret := device.GetApplicationsClock(nvml.CLOCK_GRAPHICS); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.AppClockGraphicsMhz = int64(gfxClock)
		gpu.AppClocksT = ts
	}
	if memClock, ret := device.GetApplicationsClock(nvml.CLOCK_MEM); errors.Is(ret, nvml.SUCCESS) {
		gpu.AppClockMemoryMhz = int64(memClock)
	}

	// Performance state
	if pstate, ret := device.GetPerformanceState(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PerformanceState = int(pstate)
		gpu.PerformanceStateT = ts
	}

	// Power
	if power, ret := device.GetPowerUsage(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PowerUsageMw = int64(power)
		gpu.PowerUsageMwT = ts
	}
	if limit, ret := device.GetPowerManagementLimit(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PowerLimitMw = int64(limit)
		gpu.PowerLimitMwT = ts
	}
	if enforced, ret := device.GetEnforcedPowerLimit(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PowerEnforcedLimitMw = int64(enforced)
		gpu.PowerEnforcedLimitMwT = ts
	}
	if energy, ret := device.GetTotalEnergyConsumption(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.EnergyConsumptionMj = int64(energy)
		gpu.EnergyConsumptionMjT = ts
	}

	// PCIe
	if tx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PcieTxBytesPerSec = int64(tx) * 1000 // NVML returns KB/s
		gpu.PcieTxBytesPerSecT = ts
	}
	if rx, ret := device.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PcieRxBytesPerSec = int64(rx) * 1000
		gpu.PcieRxBytesPerSecT = ts
	}
	if gen, ret := device.GetCurrPcieLinkGeneration(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PcieCurrentLinkGen = int(gen)
		gpu.PcieCurrentLinkGenT = ts
	}
	if width, ret := device.GetCurrPcieLinkWidth(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PcieCurrentLinkWidth = int(width)
		gpu.PcieCurrentLinkWidthT = ts
	}
	if replay, ret := device.GetPcieReplayCounter(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.PcieReplayCounter = int64(replay)
		gpu.PcieReplayCounterT = ts
	}

	// Throttling
	if reasons, ret := device.GetCurrentClocksEventReasons(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ClocksEventReasons = reasons
		gpu.ClocksEventReasonsT = ts
		gpu.ThrottleReasonsActive = decodeThrottleReasons(reasons)
	}

	n.collectViolationStatus(device, &gpu)
	n.collectEccErrors(device, &gpu)
	n.collectEncoderStats(device, &gpu)
	n.collectNvLinkMetrics(device, &gpu)

	if n.collectProcs {
		n.collectGPUProcesses(device, &gpu)
	}

	return gpu
}

func (n *NvidiaCollector) collectViolationStatus(device nvml.Device, gpu *types.NvidiaGPUDynamic) {
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_POWER); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ViolationPowerNs = int64(viol.ViolationTime)
		gpu.ViolationPowerNsT = ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_THERMAL); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ViolationThermalNs = int64(viol.ViolationTime)
		gpu.ViolationThermalNsT = ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_RELIABILITY); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ViolationReliabilityNs = int64(viol.ViolationTime)
		gpu.ViolationReliabilityNsT = ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_BOARD_LIMIT); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ViolationBoardLimitNs = int64(viol.ViolationTime)
		gpu.ViolationBoardLimitNsT = ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_LOW_UTILIZATION); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ViolationLowUtilNs = int64(viol.ViolationTime)
		gpu.ViolationLowUtilNsT = ts
	}
	if viol, ret := device.GetViolationStatus(nvml.PERF_POLICY_SYNC_BOOST); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.ViolationSyncBoostNs = int64(viol.ViolationTime)
		gpu.ViolationSyncBoostNsT = ts
	}
}

func (n *NvidiaCollector) collectEccErrors(device nvml.Device, gpu *types.NvidiaGPUDynamic) {
	if count, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.VOLATILE_ECC); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.EccVolatileSbe = int64(count)
		gpu.EccVolatileSbeT = ts
	}
	if count, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_UNCORRECTED, nvml.VOLATILE_ECC); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.EccVolatileDbe = int64(count)
		gpu.EccVolatileDbeT = ts
	}
	if count, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.AGGREGATE_ECC); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.EccAggregateSbe = int64(count)
		gpu.EccAggregateSbeT = ts
	}
	if count, ret := device.GetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_UNCORRECTED, nvml.AGGREGATE_ECC); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.EccAggregateDbe = int64(count)
		gpu.EccAggregateDbeT = ts
	}

	if sbe, ret := device.GetRetiredPages(nvml.PAGE_RETIREMENT_CAUSE_MULTIPLE_SINGLE_BIT_ECC_ERRORS); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.RetiredPagesSbe = int64(len(sbe))
		gpu.RetiredPagesT = ts
	}
	if dbe, ret := device.GetRetiredPages(nvml.PAGE_RETIREMENT_CAUSE_DOUBLE_BIT_ECC_ERROR); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.RetiredPagesDbe = int64(len(dbe))
		gpu.RetiredPagesT = ts
	}
	if pending, ret := device.GetRetiredPagesPendingStatus(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.RetiredPending = pending == nvml.FEATURE_ENABLED
		gpu.RetiredPendingT = ts
	}

	if correctable, uncorrectable, pending, failure, ret := device.GetRemappedRows(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.RemappedRowsCorrectable = int64(correctable)
		gpu.RemappedRowsUncorrectable = int64(uncorrectable)
		gpu.RemappedRowsPending = pending
		gpu.RemappedRowsFailure = failure
		gpu.RemappedRowsT = ts
	}
}

func (n *NvidiaCollector) collectEncoderStats(device nvml.Device, gpu *types.NvidiaGPUDynamic) {
	if sessionCount, avgFps, avgLatency, ret := device.GetEncoderStats(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.EncoderSessionCount = int(sessionCount)
		gpu.EncoderAvgFps = int(avgFps)
		gpu.EncoderAvgLatencyUs = int(avgLatency)
		gpu.EncoderStatsT = ts
	}

	if stats, ret := device.GetFBCStats(); errors.Is(ret, nvml.SUCCESS) {
		ts := GetTimestamp()
		gpu.FbcSessionCount = int(stats.SessionsCount)
		gpu.FbcAvgFps = int(stats.AverageFPS)
		gpu.FbcAvgLatencyUs = int(stats.AverageLatency)
		gpu.FbcStatsT = ts
	}
}

func (n *NvidiaCollector) collectNvLinkMetrics(device nvml.Device, gpu *types.NvidiaGPUDynamic) {
	var bandwidths []types.NvLinkBandwidth
	var linkErrors []types.NvLinkErrors

	for link := 0; link < MaxNvLinks; link++ {
		state, ret := device.GetNvLinkState(link)
		if !errors.Is(ret, nvml.SUCCESS) || state != nvml.FEATURE_ENABLED {
			break
		}

		bw := types.NvLinkBandwidth{Link: link}
		if rx, tx, ret := device.GetNvLinkUtilizationCounter(link, 0); errors.Is(ret, nvml.SUCCESS) {
			bw.TxBytes = int64(tx)
			bw.RxBytes = int64(rx)
		}
		bandwidths = append(bandwidths, bw)

		errs := types.NvLinkErrors{Link: link}
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
		if data, err := json.Marshal(bandwidths); err == nil {
			gpu.NvLinkBandwidthJSON = string(data)
		}
	}
	if len(linkErrors) > 0 {
		if data, err := json.Marshal(linkErrors); err == nil {
			gpu.NvLinkErrorsJSON = string(data)
		}
	}
}

func (n *NvidiaCollector) collectGPUProcesses(device nvml.Device, gpu *types.NvidiaGPUDynamic) {
	seen := make(map[uint32]bool)
	var procs []types.GPUProcess

	if list, ret := device.GetComputeRunningProcesses(); errors.Is(ret, nvml.SUCCESS) {
		for _, p := range list {
			if !seen[p.Pid] {
				seen[p.Pid] = true
				procs = append(procs, types.GPUProcess{
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
				procs = append(procs, types.GPUProcess{
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
				procs = append(procs, types.GPUProcess{
					PID:             p.Pid,
					Name:            getProcessName(int(p.Pid)),
					UsedMemoryBytes: int64(p.UsedGpuMemory),
					Type:            "mps",
				})
			}
		}
	}

	gpu.ProcessCount = int64(len(procs))
	gpu.ProcessCountT = GetTimestamp()

	if len(procs) > 0 {
		if data, err := json.Marshal(procs); err == nil {
			gpu.ProcessesJSON = string(data)
		}
	}

	if samples, ret := device.GetProcessUtilization(0); errors.Is(ret, nvml.SUCCESS) {
		var result []types.GPUProcessUtilization
		for _, s := range samples {
			result = append(result, types.GPUProcessUtilization{
				PID:         s.Pid,
				SmUtil:      int(s.SmUtil),
				MemUtil:     int(s.MemUtil),
				EncUtil:     int(s.EncUtil),
				DecUtil:     int(s.DecUtil),
				TimestampUs: int64(s.TimeStamp),
			})
		}
		if len(result) > 0 {
			if data, err := json.Marshal(result); err == nil {
				gpu.ProcessUtilizationJSON = string(data)
			}
		}
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func getProcessName(pid int) string {
	if content, _ := ProbeFile(fmt.Sprintf("/proc/%d/comm", pid)); content != "" {
		return content
	}
	return "unknown"
}

func decodeThrottleReasons(reasons uint64) []string {
	const (
		reasonGpuIdle              uint64 = 0x0000000000000001
		reasonAppClocksSetting     uint64 = 0x0000000000000002
		reasonSwPowerCap           uint64 = 0x0000000000000004
		reasonHwSlowdown           uint64 = 0x0000000000000008
		reasonSyncBoost            uint64 = 0x0000000000000010
		reasonSwThermalSlowdown    uint64 = 0x0000000000000020
		reasonHwThermalSlowdown    uint64 = 0x0000000000000040
		reasonHwPowerBrakeSlowdown uint64 = 0x0000000000000080
		reasonDisplayClockSetting  uint64 = 0x0000000000000100
	)

	reasonMap := map[uint64]string{
		reasonGpuIdle:              "GpuIdle",
		reasonAppClocksSetting:     "AppClocksSetting",
		reasonSwPowerCap:           "SwPowerCap",
		reasonHwSlowdown:           "HwSlowdown",
		reasonSyncBoost:            "SyncBoost",
		reasonSwThermalSlowdown:    "SwThermalSlowdown",
		reasonHwThermalSlowdown:    "HwThermalSlowdown",
		reasonHwPowerBrakeSlowdown: "HwPowerBrakeSlowdown",
		reasonDisplayClockSetting:  "DisplayClockSetting",
	}

	var active []string
	for mask, name := range reasonMap {
		if reasons&mask != 0 {
			active = append(active, name)
		}
	}
	return active
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
	switch brand {
	case nvml.BRAND_UNKNOWN:
		return "Unknown"
	case nvml.BRAND_QUADRO:
		return "Quadro"
	case nvml.BRAND_TESLA:
		return "Tesla"
	case nvml.BRAND_NVS:
		return "NVS"
	case nvml.BRAND_GRID:
		return "GRID"
	case nvml.BRAND_GEFORCE:
		return "GeForce"
	case nvml.BRAND_TITAN:
		return "Titan"
	case nvml.BRAND_NVIDIA_VAPPS:
		return "vApps"
	case nvml.BRAND_NVIDIA_VPC:
		return "VPC"
	case nvml.BRAND_NVIDIA_VCS:
		return "VCS"
	case nvml.BRAND_NVIDIA_VWS:
		return "VWS"
	case nvml.BRAND_NVIDIA_CLOUD_GAMING:
		return "CloudGaming"
	case nvml.BRAND_QUADRO_RTX:
		return "QuadroRTX"
	case nvml.BRAND_NVIDIA_RTX:
		return "NvidiaRTX"
	case nvml.BRAND_NVIDIA:
		return "Nvidia"
	case nvml.BRAND_GEFORCE_RTX:
		return "GeForceRTX"
	case nvml.BRAND_TITAN_RTX:
		return "TitanRTX"
	default:
		return fmt.Sprintf("Unknown(%d)", brand)
	}
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
