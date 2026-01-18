package collecting

import (
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type NvidiaCollector struct {
	initialized  bool
	collectProcs bool
	concurrent   bool
	devices      []nvml.Device
}

func NewNvidiaCollector(collectProcs bool, concurrent bool) *NvidiaCollector {
	n := &NvidiaCollector{
		collectProcs: collectProcs,
		concurrent:   concurrent,
	}
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

	captureStr(nvml.SystemGetDriverVersion, &m.NvidiaDriverVersion)
	if cudaVersion, ret := nvml.SystemGetCudaDriverVersion(); errors.Is(ret, nvml.SUCCESS) {
		m.NvidiaCudaVersion = fmt.Sprintf("%d.%d", cudaVersion/1000, (cudaVersion%1000)/10)
	}
	captureStr(nvml.SystemGetNVMLVersion, &m.NvmlVersion)

	gpus := make([]GPUInfo, 0, len(n.devices))
	for i, device := range n.devices {
		gpus = append(gpus, n.collectDeviceStatic(device, i))
	}
	marshalToJSON(gpus, &m.NvidiaGPUsJSON)
}

func captureStr(call func() (string, nvml.Return), dst *string) bool {
	if val, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		*dst = val
		return true
	}
	return false
}

func capture[T any](call func() (T, nvml.Return), dst *T) bool {
	if val, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		*dst = val
		return true
	}
	return false
}

func capture2[T1, T2 any](call func() (T1, T2, nvml.Return), dst1 *T1, dst2 *T2) bool {
	if val1, val2, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		*dst1, *dst2 = val1, val2
		return true
	}
	return false
}

func captureInt[T any](call func() (T, nvml.Return), dst *int, conv func(T) int) bool {
	if val, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		*dst = conv(val)
		return true
	}
	return false
}

func captureInt64[T any](call func() (T, nvml.Return), dst *int64, conv func(T) int64) bool {
	if val, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		*dst = conv(val)
		return true
	}
	return false
}

func captureTs[T any](call func() (T, nvml.Return), value *int64, timestamp *int64, conv func(T) int64) {
	if val, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		*value, *timestamp = conv(val), utils.GetTimestamp()
	}
}

func capture2Ts[T any](call func() (T, nvml.Return), v1 *int64, v2 *int64, t *int64, conv1 func(T) int64, conv2 func(T) int64) {
	if val, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		ts := utils.GetTimestamp()
		*v1, *v2, *t = conv1(val), conv2(val), ts
	}
}

func captureUtil3(call func() (uint32, uint32, nvml.Return), util *int64, utilT *int64, period *int64) {
	if u, p, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		ts := utils.GetTimestamp()
		*util, *utilT, *period = int64(u), ts, int64(p)
	}
}

func captureUtil2(call func() (uint32, uint32, nvml.Return), util *int64, utilT *int64) {
	if u, _, ret := call(); errors.Is(ret, nvml.SUCCESS) {
		*util, *utilT = int64(u), utils.GetTimestamp()
	}
}

func marshalToJSON[T any](data []T, dst *string) {
	if len(data) > 0 {
		if jsonData, err := json.Marshal(data); err == nil {
			*dst = string(jsonData)
		}
	}
}

func enumToString[T comparable](val T, mapping map[T]string, typeName string) string {
	if str, ok := mapping[val]; ok {
		return str
	}
	return fmt.Sprintf("Unknown(%v)", val)
}

func (n *NvidiaCollector) collectDeviceStatic(device nvml.Device, index int) GPUInfo {
	gpu := GPUInfo{Index: index}

	capture(device.GetName, &gpu.Name)
	capture(device.GetUUID, &gpu.UUID)
	capture(device.GetSerial, &gpu.Serial)
	capture(device.GetBoardPartNumber, &gpu.BoardPartNumber)

	if brand, ret := device.GetBrand(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Brand = brandToString(brand)
	}
	if arch, ret := device.GetArchitecture(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Architecture = archToString(arch)
	}

	capture2(device.GetCudaComputeCapability, &gpu.CudaCapabilityMajor, &gpu.CudaCapabilityMinor)

	if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.MemoryTotalBytes = int64(mem.Total)
	}
	if bar1, ret := device.GetBAR1MemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.Bar1TotalBytes = int64(bar1.Bar1Total)
	}

	captureInt(device.GetMemoryBusWidth, &gpu.MemoryBusWidthBits, func(v uint32) int { return int(v) })
	captureInt(device.GetNumGpuCores, &gpu.NumCores, func(v int) int { return v })

	for _, ct := range []struct {
		typ nvml.ClockType
		dst *int
	}{
		{nvml.CLOCK_GRAPHICS, &gpu.MaxClockGraphicsMhz},
		{nvml.CLOCK_MEM, &gpu.MaxClockMemoryMhz},
		{nvml.CLOCK_SM, &gpu.MaxClockSmMhz},
		{nvml.CLOCK_VIDEO, &gpu.MaxClockVideoMhz},
	} {
		captureInt(func() (uint32, nvml.Return) { return device.GetMaxClockInfo(ct.typ) }, ct.dst, func(v uint32) int { return int(v) })
	}

	if pci, ret := device.GetPciInfo(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PciBusId = utils.Int8SliceToString(pci.BusId[:])
		gpu.PciDeviceId = pci.PciDeviceId
		gpu.PciSubsystemId = pci.PciSubSystemId
	}
	captureInt(device.GetMaxPcieLinkGeneration, &gpu.PcieMaxLinkGen, func(v int) int { return v })
	captureInt(device.GetMaxPcieLinkWidth, &gpu.PcieMaxLinkWidth, func(v int) int { return int(v) })

	captureInt(device.GetPowerManagementDefaultLimit, &gpu.PowerDefaultLimitMw, func(v uint32) int { return int(v) })
	if minLimit, maxLimit, ret := device.GetPowerManagementLimitConstraints(); errors.Is(ret, nvml.SUCCESS) {
		gpu.PowerMinLimitMw = int(minLimit)
		gpu.PowerMaxLimitMw = int(maxLimit)
	}

	capture(device.GetVbiosVersion, &gpu.VbiosVersion)
	capture(device.GetInforomImageVersion, &gpu.InforomImageVersion)
	captureStr(func() (string, nvml.Return) { return device.GetInforomVersion(nvml.INFOROM_OEM) }, &gpu.InforomOemVersion)

	captureInt(device.GetNumFans, &gpu.NumFans, func(v int) int { return int(v) })

	for _, tt := range []struct {
		typ nvml.TemperatureThresholds
		dst *int
	}{
		{nvml.TEMPERATURE_THRESHOLD_SHUTDOWN, &gpu.TempShutdownC},
		{nvml.TEMPERATURE_THRESHOLD_SLOWDOWN, &gpu.TempSlowdownC},
		{nvml.TEMPERATURE_THRESHOLD_GPU_MAX, &gpu.TempMaxOperatingC},
		{nvml.TEMPERATURE_THRESHOLD_ACOUSTIC_CURR, &gpu.TempTargetC},
	} {
		captureInt(func() (uint32, nvml.Return) { return device.GetTemperatureThreshold(tt.typ) }, tt.dst, func(v uint32) int { return int(v) })
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

	for _, et := range []struct {
		typ nvml.EncoderType
		dst *int
	}{
		{nvml.ENCODER_QUERY_H264, &gpu.EncoderCapacityH264},
		{nvml.ENCODER_QUERY_HEVC, &gpu.EncoderCapacityHEVC},
		{nvml.ENCODER_QUERY_AV1, &gpu.EncoderCapacityAV1},
	} {
		capture(func() (int, nvml.Return) { return device.GetEncoderCapacity(et.typ) }, et.dst)
	}

	gpu.NvLinkCount = countActiveNvLinks(device)

	return gpu
}

func countActiveNvLinks(device nvml.Device) int {
	count := 0
	for link := 0; link < maxNvlinks; link++ {
		if _, ret := device.GetNvLinkState(link); !errors.Is(ret, nvml.SUCCESS) {
			break
		}
		count++
	}
	return count
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

	var wg sync.WaitGroup
	run := func(fn func()) {
		if n.concurrent {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn()
			}()
		} else {
			fn()
		}
	}

	run(func() {
		if util, ret := device.GetUtilizationRates(); errors.Is(ret, nvml.SUCCESS) {
			ts := utils.GetTimestamp()
			gpu.UtilizationGPU, gpu.UtilizationGPUT = int64(util.Gpu), ts
			gpu.UtilizationMemory, gpu.UtilizationMemoryT = int64(util.Memory), ts
		}
	})
	run(func() {
		captureUtil3(device.GetEncoderUtilization, &gpu.UtilizationEncoder, &gpu.UtilizationEncoderT, &gpu.EncoderSamplingPeriodUs)
	})
	run(func() {
		captureUtil3(device.GetDecoderUtilization, &gpu.UtilizationDecoder, &gpu.UtilizationDecoderT, &gpu.DecoderSamplingPeriodUs)
	})
	run(func() { captureUtil2(device.GetJpgUtilization, &gpu.UtilizationJpeg, &gpu.UtilizationJpegT) })
	run(func() { captureUtil2(device.GetOfaUtilization, &gpu.UtilizationOfa, &gpu.UtilizationOfaT) })

	run(func() {
		if mem, ret := device.GetMemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
			ts := utils.GetTimestamp()
			gpu.MemoryUsedBytes, gpu.MemoryUsedBytesT = int64(mem.Used), ts
			gpu.MemoryFreeBytes, gpu.MemoryFreeBytesT = int64(mem.Free), ts
			gpu.MemoryTotalBytes = int64(mem.Total)
		}
	})
	run(func() {
		if mem, ret := device.GetMemoryInfo_v2(); errors.Is(ret, nvml.SUCCESS) {
			gpu.MemoryReservedBytes, gpu.MemoryReservedBytesT = int64(mem.Reserved), utils.GetTimestamp()
		}
	})
	run(func() {
		if bar1, ret := device.GetBAR1MemoryInfo(); errors.Is(ret, nvml.SUCCESS) {
			ts := utils.GetTimestamp()
			gpu.Bar1UsedBytes, gpu.Bar1UsedBytesT = int64(bar1.Bar1Used), ts
			gpu.Bar1FreeBytes, gpu.Bar1FreeBytesT = int64(bar1.Bar1Free), ts
			gpu.Bar1TotalBytes = int64(bar1.Bar1Total)
		}
	})

	for _, tm := range []struct {
		typ nvml.TemperatureSensors
		val *int64
		ts  *int64
	}{
		{nvml.TEMPERATURE_GPU, &gpu.TemperatureGpuC, &gpu.TemperatureGpuCT},
		{nvml.TEMPERATURE_COUNT, &gpu.TemperatureMemoryC, &gpu.TemperatureMemoryCT},
	} {
		tm := tm
		run(func() {
			captureTs(func() (uint32, nvml.Return) { return device.GetTemperature(tm.typ) }, tm.val, tm.ts, func(v uint32) int64 { return int64(v) })
		})
	}

	run(func() {
		captureTs(device.GetFanSpeed, &gpu.FanSpeedPercent, &gpu.FanSpeedPercentT, func(v uint32) int64 { return int64(v) })
	})
	run(func() {
		if numFans, ret := device.GetNumFans(); errors.Is(ret, nvml.SUCCESS) && numFans > 1 {
			speeds := make([]int, 0, numFans)
			for f := 0; f < int(numFans); f++ {
				if speed, ret := device.GetFanSpeed_v2(f); errors.Is(ret, nvml.SUCCESS) {
					speeds = append(speeds, int(speed))
				}
			}
			marshalToJSON(speeds, &gpu.FanSpeedsJSON)
		}
	})

	for _, cm := range []struct {
		typ nvml.ClockType
		val *int64
		ts  *int64
	}{
		{nvml.CLOCK_GRAPHICS, &gpu.ClockGraphicsMhz, &gpu.ClockGraphicsMhzT},
		{nvml.CLOCK_SM, &gpu.ClockSmMhz, &gpu.ClockSmMhzT},
		{nvml.CLOCK_MEM, &gpu.ClockMemoryMhz, &gpu.ClockMemoryMhzT},
		{nvml.CLOCK_VIDEO, &gpu.ClockVideoMhz, &gpu.ClockVideoMhzT},
	} {
		cm := cm
		run(func() {
			captureTs(func() (uint32, nvml.Return) { return device.GetClockInfo(cm.typ) }, cm.val, cm.ts, func(v uint32) int64 { return int64(v) })
		})
	}

	run(func() {
		if pstate, ret := device.GetPerformanceState(); errors.Is(ret, nvml.SUCCESS) {
			gpu.PerformanceState, gpu.PerformanceStateT = int(pstate), utils.GetTimestamp()
		}
	})
	run(func() {
		captureTs(device.GetPowerUsage, &gpu.PowerUsageMw, &gpu.PowerUsageMwT, func(v uint32) int64 { return int64(v) })
	})
	run(func() {
		captureTs(device.GetPowerManagementLimit, &gpu.PowerLimitMw, &gpu.PowerLimitMwT, func(v uint32) int64 { return int64(v) })
	})
	run(func() {
		captureTs(device.GetEnforcedPowerLimit, &gpu.PowerEnforcedLimitMw, &gpu.PowerEnforcedLimitMwT, func(v uint32) int64 { return int64(v) })
	})
	run(func() {
		captureTs(device.GetTotalEnergyConsumption, &gpu.EnergyConsumptionMj, &gpu.EnergyConsumptionMjT, func(v uint64) int64 { return int64(v) })
	})

	for _, pm := range []struct {
		counter nvml.PcieUtilCounter
		val     *int64
		ts      *int64
		scale   int64
	}{
		{nvml.PCIE_UTIL_TX_BYTES, &gpu.PcieTxBytesPerSec, &gpu.PcieTxBytesPerSecT, 1000},
		{nvml.PCIE_UTIL_RX_BYTES, &gpu.PcieRxBytesPerSec, &gpu.PcieRxBytesPerSecT, 1000},
	} {
		pm := pm
		run(func() {
			if throughput, ret := device.GetPcieThroughput(pm.counter); errors.Is(ret, nvml.SUCCESS) {
				*pm.val, *pm.ts = int64(throughput)*pm.scale, utils.GetTimestamp()
			}
		})
	}

	run(func() {
		if gen, ret := device.GetCurrPcieLinkGeneration(); errors.Is(ret, nvml.SUCCESS) {
			gpu.PcieCurrentLinkGen, gpu.PcieCurrentLinkGenT = int(gen), utils.GetTimestamp()
		}
	})
	run(func() {
		if width, ret := device.GetCurrPcieLinkWidth(); errors.Is(ret, nvml.SUCCESS) {
			gpu.PcieCurrentLinkWidth, gpu.PcieCurrentLinkWidthT = int(width), utils.GetTimestamp()
		}
	})
	run(func() {
		captureTs(device.GetPcieReplayCounter, &gpu.PcieReplayCounter, &gpu.PcieReplayCounterT, func(v int) int64 { return int64(v) })
	})

	run(func() {
		if reasons, ret := device.GetCurrentClocksEventReasons(); errors.Is(ret, nvml.SUCCESS) {
			gpu.ClocksEventReasons, gpu.ClocksEventReasonsT = reasons, utils.GetTimestamp()
		}
	})

	run(func() { n.collectViolationStatus(device, &gpu) })
	run(func() { n.collectEccErrors(device, &gpu) })
	run(func() { n.collectNvLinkMetrics(device, &gpu) })
	if n.collectProcs {
		run(func() { n.collectGPUProcesses(device, &gpu) })
	}

	wg.Wait()
	return gpu
}

func (n *NvidiaCollector) collectViolationStatus(device nvml.Device, gpu *GPUDynamic) {
	for _, v := range []struct {
		policy nvml.PerfPolicyType
		val    *int64
		ts     *int64
	}{
		{nvml.PERF_POLICY_POWER, &gpu.ViolationPowerNs, &gpu.ViolationPowerNsT},
		{nvml.PERF_POLICY_THERMAL, &gpu.ViolationThermalNs, &gpu.ViolationThermalNsT},
		{nvml.PERF_POLICY_RELIABILITY, &gpu.ViolationReliabilityNs, &gpu.ViolationReliabilityNsT},
		{nvml.PERF_POLICY_BOARD_LIMIT, &gpu.ViolationBoardLimitNs, &gpu.ViolationBoardLimitNsT},
		{nvml.PERF_POLICY_LOW_UTILIZATION, &gpu.ViolationLowUtilNs, &gpu.ViolationLowUtilNsT},
		{nvml.PERF_POLICY_SYNC_BOOST, &gpu.ViolationSyncBoostNs, &gpu.ViolationSyncBoostNsT},
	} {
		if viol, ret := device.GetViolationStatus(v.policy); errors.Is(ret, nvml.SUCCESS) {
			*v.val, *v.ts = int64(viol.ViolationTime), utils.GetTimestamp()
		}
	}
}

func (n *NvidiaCollector) collectEccErrors(device nvml.Device, gpu *GPUDynamic) {
	for _, et := range []struct {
		errType  nvml.MemoryErrorType
		location nvml.EccCounterType
		val      *int64
		ts       *int64
	}{
		{nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.AGGREGATE_ECC, &gpu.EccAggregateSbe, &gpu.EccAggregateSbeT},
		{nvml.MEMORY_ERROR_TYPE_UNCORRECTED, nvml.AGGREGATE_ECC, &gpu.EccAggregateDbe, &gpu.EccAggregateDbeT},
	} {
		if count, ret := device.GetTotalEccErrors(et.errType, et.location); errors.Is(ret, nvml.SUCCESS) {
			*et.val, *et.ts = int64(count), utils.GetTimestamp()
		}
	}

	for _, rt := range []struct {
		cause nvml.PageRetirementCause
		val   *int64
		ts    *int64
	}{
		{nvml.PAGE_RETIREMENT_CAUSE_MULTIPLE_SINGLE_BIT_ECC_ERRORS, &gpu.RetiredPagesSbe, &gpu.RetiredPagesT},
		{nvml.PAGE_RETIREMENT_CAUSE_DOUBLE_BIT_ECC_ERROR, &gpu.RetiredPagesDbe, &gpu.RetiredPagesT},
	} {
		if pages, ret := device.GetRetiredPages(rt.cause); errors.Is(ret, nvml.SUCCESS) {
			*rt.val, *rt.ts = int64(len(pages)), utils.GetTimestamp()
		}
	}

	if pending, ret := device.GetRetiredPagesPendingStatus(); errors.Is(ret, nvml.SUCCESS) {
		gpu.RetiredPending, gpu.RetiredPendingT = pending == nvml.FEATURE_ENABLED, utils.GetTimestamp()
	}
	if correctable, uncorrectable, pending, failure, ret := device.GetRemappedRows(); errors.Is(ret, nvml.SUCCESS) {
		ts := utils.GetTimestamp()
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

	for link := 0; link < maxNvlinks; link++ {
		state, ret := device.GetNvLinkState(link)
		if !errors.Is(ret, nvml.SUCCESS) || state != nvml.FEATURE_ENABLED {
			break
		}

		bw := NvLinkBandwidth{Link: link}
		if rx, tx, ret := device.GetNvLinkUtilizationCounter(link, 0); errors.Is(ret, nvml.SUCCESS) {
			bw.TxBytes, bw.RxBytes = int64(tx), int64(rx)
		}
		bandwidths = append(bandwidths, bw)

		errs := NvLinkErrors{Link: link}
		for _, et := range []struct {
			errType nvml.NvLinkErrorCounter
			dst     *int64
		}{
			{nvml.NVLINK_ERROR_DL_CRC_FLIT, &errs.CrcErrors},
			{nvml.NVLINK_ERROR_DL_ECC_DATA, &errs.EccErrors},
			{nvml.NVLINK_ERROR_DL_REPLAY, &errs.ReplayErrors},
			{nvml.NVLINK_ERROR_DL_RECOVERY, &errs.RecoveryCount},
		} {
			captureInt64(func() (uint64, nvml.Return) { return device.GetNvLinkErrorCounter(link, et.errType) }, et.dst, func(v uint64) int64 { return int64(v) })
		}
		linkErrors = append(linkErrors, errs)
	}

	marshalToJSON(bandwidths, &gpu.NvLinkBandwidthJSON)
	marshalToJSON(linkErrors, &gpu.NvLinkErrorsJSON)
}

func (n *NvidiaCollector) collectGPUProcesses(device nvml.Device, gpu *GPUDynamic) {
	ts := utils.GetTimestamp()
	seen := make(map[uint32]bool)
	var procs []GPUProcess

	for _, pl := range []struct {
		call  func() ([]nvml.ProcessInfo, nvml.Return)
		pType string
	}{
		{device.GetComputeRunningProcesses, "compute"},
		{device.GetGraphicsRunningProcesses, "graphics"},
		{device.GetMPSComputeRunningProcesses, "mps"},
	} {
		if list, ret := pl.call(); errors.Is(ret, nvml.SUCCESS) {
			for _, p := range list {
				if !seen[p.Pid] {
					seen[p.Pid] = true
					procs = append(procs, GPUProcess{
						PID:             p.Pid,
						Name:            getProcessName(int(p.Pid)),
						UsedMemoryBytes: int64(p.UsedGpuMemory),
						Type:            pl.pType,
					})
				}
			}
		}
	}

	gpu.ProcessCount, gpu.ProcessCountT = int64(len(procs)), ts
	marshalToJSON(procs, &gpu.ProcessesJSON)

	if samples, ret := device.GetProcessUtilization(0); errors.Is(ret, nvml.SUCCESS) {
		result := make([]GPUProcessUtilization, 0, len(samples))
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
		marshalToJSON(result, &gpu.ProcessUtilizationJSON)
	}
}

func getProcessName(pid int) string {
	val, _, _ := utils.File(fmt.Sprintf("/proc/%d/comm", pid))
	return val
}

func archToString(arch nvml.DeviceArchitecture) string {
	return enumToString(arch, map[nvml.DeviceArchitecture]string{
		nvml.DEVICE_ARCH_KEPLER:  "Kepler",
		nvml.DEVICE_ARCH_MAXWELL: "Maxwell",
		nvml.DEVICE_ARCH_PASCAL:  "Pascal",
		nvml.DEVICE_ARCH_VOLTA:   "Volta",
		nvml.DEVICE_ARCH_TURING:  "Turing",
		nvml.DEVICE_ARCH_AMPERE:  "Ampere",
		nvml.DEVICE_ARCH_ADA:     "Ada",
		nvml.DEVICE_ARCH_HOPPER:  "Hopper",
	}, "Architecture")
}

func brandToString(brand nvml.BrandType) string {
	return enumToString(brand, map[nvml.BrandType]string{
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
	}, "Brand")
}

func computeModeToString(mode nvml.ComputeMode) string {
	return enumToString(mode, map[nvml.ComputeMode]string{
		nvml.COMPUTEMODE_DEFAULT:           "Default",
		nvml.COMPUTEMODE_EXCLUSIVE_THREAD:  "ExclusiveThread",
		nvml.COMPUTEMODE_PROHIBITED:        "Prohibited",
		nvml.COMPUTEMODE_EXCLUSIVE_PROCESS: "ExclusiveProcess",
	}, "ComputeMode")
}
