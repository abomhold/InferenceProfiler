package output

import (
	"InferenceProfiler/src/collectors"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// FlattenMetrics converts DynamicMetrics to a flat map[string]interface{}
// GPU metrics are flattened as nvidia{i}FieldName (e.g., nvidia0UtilizationGpu)
// Process metrics are flattened as process{i}FieldName (e.g., process0Pid)
// Other fields are copied using their json tag names
func FlattenMetrics(m *collectors.DynamicMetrics) map[string]interface{} {
	flat := make(map[string]interface{})

	// Flatten all scalar fields using reflection
	v := reflect.ValueOf(m).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip slices - we handle them specially below
		if field.Name == "NvidiaGPUs" || field.Name == "Processes" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]

		// For scalar types, just copy the value
		flat[jsonName] = value.Interface()
	}

	// Add GPU count and flatten each GPU
	flat["nvidiaGpuCount"] = len(m.NvidiaGPUs)
	for _, gpu := range m.NvidiaGPUs {
		flattenGPU(flat, gpu)
	}

	// Add process count and flatten each process
	flat["processCount"] = len(m.Processes)
	for i, proc := range m.Processes {
		flattenProcess(flat, proc, i)
	}

	return flat
}

// ToJSONMode converts DynamicMetrics to a map with nested data as JSON strings
// GPUs are serialized as nvidiaGpusJson, processes as processesJson
func ToJSONMode(m *collectors.DynamicMetrics) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy all scalar fields using reflection
	v := reflect.ValueOf(m).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip slices - we handle them specially below
		if field.Name == "NvidiaGPUs" || field.Name == "Processes" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]

		// For scalar types, just copy the value
		result[jsonName] = value.Interface()
	}

	// Serialize GPUs as JSON string
	result["nvidiaGpuCount"] = len(m.NvidiaGPUs)
	if len(m.NvidiaGPUs) > 0 {
		if data, err := json.Marshal(m.NvidiaGPUs); err == nil {
			result["nvidiaGpusJson"] = string(data)
		}
	}

	// Serialize processes as JSON string
	result["processCount"] = len(m.Processes)
	if len(m.Processes) > 0 {
		if data, err := json.Marshal(m.Processes); err == nil {
			result["processesJson"] = string(data)
		}
	}

	return result
}

// flattenGPU adds GPU metrics to the flat map with nvidia{index}FieldName format
func flattenGPU(flat map[string]interface{}, gpu collectors.NvidiaGPUDynamic) {
	prefix := fmt.Sprintf("nvidia%d", gpu.Index)

	// =========================================================================
	// UTILIZATION
	// =========================================================================
	flat[prefix+"UtilizationGpu"] = gpu.UtilizationGPU
	flat[prefix+"UtilizationGpuT"] = gpu.UtilizationGPUT
	flat[prefix+"UtilizationMemory"] = gpu.UtilizationMemory
	flat[prefix+"UtilizationMemoryT"] = gpu.UtilizationMemoryT
	flat[prefix+"UtilizationEncoder"] = gpu.UtilizationEncoder
	flat[prefix+"UtilizationEncoderT"] = gpu.UtilizationEncoderT
	flat[prefix+"UtilizationDecoder"] = gpu.UtilizationDecoder
	flat[prefix+"UtilizationDecoderT"] = gpu.UtilizationDecoderT

	// Optional utilization fields (Turing+)
	if gpu.UtilizationJpeg != 0 || gpu.UtilizationJpegT != 0 {
		flat[prefix+"UtilizationJpeg"] = gpu.UtilizationJpeg
		flat[prefix+"UtilizationJpegT"] = gpu.UtilizationJpegT
	}
	if gpu.UtilizationOfa != 0 || gpu.UtilizationOfaT != 0 {
		flat[prefix+"UtilizationOfa"] = gpu.UtilizationOfa
		flat[prefix+"UtilizationOfaT"] = gpu.UtilizationOfaT
	}

	// Sampling periods
	if gpu.EncoderSamplingPeriodUs != 0 {
		flat[prefix+"EncoderSamplingPeriodUs"] = gpu.EncoderSamplingPeriodUs
	}
	if gpu.DecoderSamplingPeriodUs != 0 {
		flat[prefix+"DecoderSamplingPeriodUs"] = gpu.DecoderSamplingPeriodUs
	}

	// =========================================================================
	// MEMORY
	// =========================================================================
	flat[prefix+"MemoryUsedBytes"] = gpu.MemoryUsedBytes
	flat[prefix+"MemoryUsedBytesT"] = gpu.MemoryUsedBytesT
	flat[prefix+"MemoryFreeBytes"] = gpu.MemoryFreeBytes
	flat[prefix+"MemoryFreeBytesT"] = gpu.MemoryFreeBytesT
	flat[prefix+"MemoryTotalBytes"] = gpu.MemoryTotalBytes

	if gpu.MemoryReservedBytes != 0 {
		flat[prefix+"MemoryReservedBytes"] = gpu.MemoryReservedBytes
		flat[prefix+"MemoryReservedBytesT"] = gpu.MemoryReservedBytesT
	}

	flat[prefix+"Bar1UsedBytes"] = gpu.Bar1UsedBytes
	flat[prefix+"Bar1UsedBytesT"] = gpu.Bar1UsedBytesT
	flat[prefix+"Bar1FreeBytes"] = gpu.Bar1FreeBytes
	flat[prefix+"Bar1FreeBytesT"] = gpu.Bar1FreeBytesT
	flat[prefix+"Bar1TotalBytes"] = gpu.Bar1TotalBytes

	// =========================================================================
	// TEMPERATURE
	// =========================================================================
	flat[prefix+"TemperatureGpuC"] = gpu.TemperatureGpuC
	flat[prefix+"TemperatureGpuCT"] = gpu.TemperatureGpuCT

	if gpu.TemperatureMemoryC != 0 || gpu.TemperatureMemoryCT != 0 {
		flat[prefix+"TemperatureMemoryC"] = gpu.TemperatureMemoryC
		flat[prefix+"TemperatureMemoryCT"] = gpu.TemperatureMemoryCT
	}

	// =========================================================================
	// FAN
	// =========================================================================
	flat[prefix+"FanSpeedPercent"] = gpu.FanSpeedPercent
	flat[prefix+"FanSpeedPercentT"] = gpu.FanSpeedPercentT

	if gpu.FanSpeedsJSON != "" {
		flat[prefix+"FanSpeedsJson"] = gpu.FanSpeedsJSON
	}

	// =========================================================================
	// CLOCKS
	// =========================================================================
	flat[prefix+"ClockGraphicsMhz"] = gpu.ClockGraphicsMhz
	flat[prefix+"ClockGraphicsMhzT"] = gpu.ClockGraphicsMhzT
	flat[prefix+"ClockSmMhz"] = gpu.ClockSmMhz
	flat[prefix+"ClockSmMhzT"] = gpu.ClockSmMhzT
	flat[prefix+"ClockMemoryMhz"] = gpu.ClockMemoryMhz
	flat[prefix+"ClockMemoryMhzT"] = gpu.ClockMemoryMhzT
	flat[prefix+"ClockVideoMhz"] = gpu.ClockVideoMhz
	flat[prefix+"ClockVideoMhzT"] = gpu.ClockVideoMhzT

	// Application clocks (optional)
	if gpu.AppClockGraphicsMhz != 0 || gpu.AppClockMemoryMhz != 0 {
		flat[prefix+"AppClockGraphicsMhz"] = gpu.AppClockGraphicsMhz
		flat[prefix+"AppClockMemoryMhz"] = gpu.AppClockMemoryMhz
		flat[prefix+"AppClocksT"] = gpu.AppClocksT
	}

	// =========================================================================
	// PERFORMANCE STATE
	// =========================================================================
	flat[prefix+"PerformanceState"] = gpu.PerformanceState
	flat[prefix+"PerformanceStateT"] = gpu.PerformanceStateT

	// =========================================================================
	// POWER
	// =========================================================================
	flat[prefix+"PowerUsageMw"] = gpu.PowerUsageMw
	flat[prefix+"PowerUsageMwT"] = gpu.PowerUsageMwT
	flat[prefix+"PowerLimitMw"] = gpu.PowerLimitMw
	flat[prefix+"PowerLimitMwT"] = gpu.PowerLimitMwT
	flat[prefix+"PowerEnforcedLimitMw"] = gpu.PowerEnforcedLimitMw
	flat[prefix+"PowerEnforcedLimitMwT"] = gpu.PowerEnforcedLimitMwT
	flat[prefix+"EnergyConsumptionMj"] = gpu.EnergyConsumptionMj
	flat[prefix+"EnergyConsumptionMjT"] = gpu.EnergyConsumptionMjT

	// =========================================================================
	// PCIe
	// =========================================================================
	flat[prefix+"PcieTxBytesPerSec"] = gpu.PcieTxBytesPerSec
	flat[prefix+"PcieTxBytesPerSecT"] = gpu.PcieTxBytesPerSecT
	flat[prefix+"PcieRxBytesPerSec"] = gpu.PcieRxBytesPerSec
	flat[prefix+"PcieRxBytesPerSecT"] = gpu.PcieRxBytesPerSecT
	flat[prefix+"PcieCurrentLinkGen"] = gpu.PcieCurrentLinkGen
	flat[prefix+"PcieCurrentLinkGenT"] = gpu.PcieCurrentLinkGenT
	flat[prefix+"PcieCurrentLinkWidth"] = gpu.PcieCurrentLinkWidth
	flat[prefix+"PcieCurrentLinkWidthT"] = gpu.PcieCurrentLinkWidthT
	flat[prefix+"PcieReplayCounter"] = gpu.PcieReplayCounter
	flat[prefix+"PcieReplayCounterT"] = gpu.PcieReplayCounterT

	// =========================================================================
	// THROTTLING
	// =========================================================================
	flat[prefix+"ClocksEventReasons"] = gpu.ClocksEventReasons
	flat[prefix+"ClocksEventReasonsT"] = gpu.ClocksEventReasonsT

	if len(gpu.ThrottleReasonsActive) > 0 {
		if data, err := json.Marshal(gpu.ThrottleReasonsActive); err == nil {
			flat[prefix+"ThrottleReasonsActiveJson"] = string(data)
		}
	}

	// Violation times
	flat[prefix+"ViolationPowerNs"] = gpu.ViolationPowerNs
	flat[prefix+"ViolationPowerNsT"] = gpu.ViolationPowerNsT
	flat[prefix+"ViolationThermalNs"] = gpu.ViolationThermalNs
	flat[prefix+"ViolationThermalNsT"] = gpu.ViolationThermalNsT

	if gpu.ViolationReliabilityNs != 0 {
		flat[prefix+"ViolationReliabilityNs"] = gpu.ViolationReliabilityNs
		flat[prefix+"ViolationReliabilityNsT"] = gpu.ViolationReliabilityNsT
	}
	if gpu.ViolationBoardLimitNs != 0 {
		flat[prefix+"ViolationBoardLimitNs"] = gpu.ViolationBoardLimitNs
		flat[prefix+"ViolationBoardLimitNsT"] = gpu.ViolationBoardLimitNsT
	}
	if gpu.ViolationLowUtilNs != 0 {
		flat[prefix+"ViolationLowUtilNs"] = gpu.ViolationLowUtilNs
		flat[prefix+"ViolationLowUtilNsT"] = gpu.ViolationLowUtilNsT
	}
	if gpu.ViolationSyncBoostNs != 0 {
		flat[prefix+"ViolationSyncBoostNs"] = gpu.ViolationSyncBoostNs
		flat[prefix+"ViolationSyncBoostNsT"] = gpu.ViolationSyncBoostNsT
	}

	// =========================================================================
	// ECC ERRORS
	// =========================================================================
	flat[prefix+"EccVolatileSbe"] = gpu.EccVolatileSbe
	flat[prefix+"EccVolatileSbeT"] = gpu.EccVolatileSbeT
	flat[prefix+"EccVolatileDbe"] = gpu.EccVolatileDbe
	flat[prefix+"EccVolatileDbeT"] = gpu.EccVolatileDbeT
	flat[prefix+"EccAggregateSbe"] = gpu.EccAggregateSbe
	flat[prefix+"EccAggregateSbeT"] = gpu.EccAggregateSbeT
	flat[prefix+"EccAggregateDbe"] = gpu.EccAggregateDbe
	flat[prefix+"EccAggregateDbeT"] = gpu.EccAggregateDbeT

	// Retired pages (optional)
	if gpu.RetiredPagesSbe != 0 || gpu.RetiredPagesDbe != 0 {
		flat[prefix+"RetiredPagesSbe"] = gpu.RetiredPagesSbe
		flat[prefix+"RetiredPagesDbe"] = gpu.RetiredPagesDbe
		flat[prefix+"RetiredPagesT"] = gpu.RetiredPagesT
		flat[prefix+"RetiredPending"] = gpu.RetiredPending
		flat[prefix+"RetiredPendingT"] = gpu.RetiredPendingT
	}

	// Remapped rows (Ampere+)
	if gpu.RemappedRowsCorrectable != 0 || gpu.RemappedRowsUncorrectable != 0 || gpu.RemappedRowsPending || gpu.RemappedRowsFailure {
		flat[prefix+"RemappedRowsCorrectable"] = gpu.RemappedRowsCorrectable
		flat[prefix+"RemappedRowsUncorrectable"] = gpu.RemappedRowsUncorrectable
		flat[prefix+"RemappedRowsPending"] = gpu.RemappedRowsPending
		flat[prefix+"RemappedRowsFailure"] = gpu.RemappedRowsFailure
		flat[prefix+"RemappedRowsT"] = gpu.RemappedRowsT
	}

	// =========================================================================
	// ENCODER/DECODER STATS
	// =========================================================================
	flat[prefix+"EncoderSessionCount"] = gpu.EncoderSessionCount
	flat[prefix+"EncoderAvgFps"] = gpu.EncoderAvgFps
	flat[prefix+"EncoderAvgLatencyUs"] = gpu.EncoderAvgLatencyUs
	flat[prefix+"EncoderStatsT"] = gpu.EncoderStatsT

	flat[prefix+"FbcSessionCount"] = gpu.FbcSessionCount
	flat[prefix+"FbcAvgFps"] = gpu.FbcAvgFps
	flat[prefix+"FbcAvgLatencyUs"] = gpu.FbcAvgLatencyUs
	flat[prefix+"FbcStatsT"] = gpu.FbcStatsT

	// =========================================================================
	// NVLINK
	// =========================================================================
	if gpu.NvLinkBandwidthJSON != "" {
		flat[prefix+"NvlinkBandwidthJson"] = gpu.NvLinkBandwidthJSON
	}
	if gpu.NvLinkErrorsJSON != "" {
		flat[prefix+"NvlinkErrorsJson"] = gpu.NvLinkErrorsJSON
	}

	// =========================================================================
	// PROCESSES
	// =========================================================================
	flat[prefix+"ProcessCount"] = gpu.ProcessCount
	flat[prefix+"ProcessCountT"] = gpu.ProcessCountT

	if gpu.ProcessesJSON != "" {
		flat[prefix+"ProcessesJson"] = gpu.ProcessesJSON
	}
	if gpu.ProcessUtilizationJSON != "" {
		flat[prefix+"ProcessUtilizationJson"] = gpu.ProcessUtilizationJSON
	}
}

// flattenProcess adds process metrics to the flat map with process{index}FieldName format
func flattenProcess(flat map[string]interface{}, proc collectors.ProcessMetrics, index int) {
	prefix := fmt.Sprintf("process%d", index)

	flat[prefix+"Pid"] = proc.PID
	flat[prefix+"PidT"] = proc.PIDT
	flat[prefix+"Name"] = proc.Name
	flat[prefix+"NameT"] = proc.NameT
	flat[prefix+"Cmdline"] = proc.Cmdline
	flat[prefix+"CmdlineT"] = proc.CmdlineT
	flat[prefix+"NumThreads"] = proc.NumThreads
	flat[prefix+"NumThreadsT"] = proc.NumThreadsT
	flat[prefix+"CpuTimeUserMode"] = proc.CPUTimeUserMode
	flat[prefix+"CpuTimeUserModeT"] = proc.CPUTimeUserModeT
	flat[prefix+"CpuTimeKernelMode"] = proc.CPUTimeKernelMode
	flat[prefix+"CpuTimeKernelModeT"] = proc.CPUTimeKernelModeT
	flat[prefix+"ChildrenUserMode"] = proc.ChildrenUserMode
	flat[prefix+"ChildrenUserModeT"] = proc.ChildrenUserModeT
	flat[prefix+"ChildrenKernelMode"] = proc.ChildrenKernelMode
	flat[prefix+"ChildrenKernelModeT"] = proc.ChildrenKernelModeT
	flat[prefix+"VoluntaryCtxSwitches"] = proc.VoluntaryContextSwitches
	flat[prefix+"VoluntaryCtxSwitchesT"] = proc.VoluntaryContextSwitchesT
	flat[prefix+"NonvoluntaryCtxSwitches"] = proc.NonvoluntaryContextSwitches
	flat[prefix+"NonvoluntaryCtxSwitchesT"] = proc.NonvoluntaryContextSwitchesT
	flat[prefix+"BlockIODelays"] = proc.BlockIODelays
	flat[prefix+"BlockIODelaysT"] = proc.BlockIODelaysT
	flat[prefix+"VirtualMemoryBytes"] = proc.VirtualMemoryBytes
	flat[prefix+"VirtualMemoryBytesT"] = proc.VirtualMemoryBytesT
	flat[prefix+"ResidentSetSize"] = proc.ResidentSetSize
	flat[prefix+"ResidentSetSizeT"] = proc.ResidentSetSizeT
}
