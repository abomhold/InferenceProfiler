package aggregate

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

	flat[prefix+"UtilizationGpu"] = gpu.UtilizationGPU
	flat[prefix+"UtilizationGpuT"] = gpu.UtilizationGPUT
	flat[prefix+"UtilizationMem"] = gpu.UtilizationMem
	flat[prefix+"UtilizationMemT"] = gpu.UtilizationMemT
	flat[prefix+"MemoryUsedMb"] = gpu.MemoryUsedMb
	flat[prefix+"MemoryUsedMbT"] = gpu.MemoryUsedMbT
	flat[prefix+"MemoryFreeMb"] = gpu.MemoryFreeMb
	flat[prefix+"MemoryFreeMbT"] = gpu.MemoryFreeMbT
	flat[prefix+"Bar1UsedMb"] = gpu.Bar1UsedMb
	flat[prefix+"Bar1UsedMbT"] = gpu.Bar1UsedMbT
	flat[prefix+"TemperatureC"] = gpu.TemperatureC
	flat[prefix+"TemperatureCT"] = gpu.TemperatureCT
	flat[prefix+"FanSpeed"] = gpu.FanSpeed
	flat[prefix+"FanSpeedT"] = gpu.FanSpeedT
	flat[prefix+"ClockGraphicsMhz"] = gpu.ClockGraphicsMhz
	flat[prefix+"ClockGraphicsMhzT"] = gpu.ClockGraphicsMhzT
	flat[prefix+"ClockSmMhz"] = gpu.ClockSmMhz
	flat[prefix+"ClockSmMhzT"] = gpu.ClockSmMhzT
	flat[prefix+"ClockMemMhz"] = gpu.ClockMemMhz
	flat[prefix+"ClockMemMhzT"] = gpu.ClockMemMhzT
	flat[prefix+"PcieTxKbps"] = gpu.PcieTxKbps
	flat[prefix+"PcieTxKbpsT"] = gpu.PcieTxKbpsT
	flat[prefix+"PcieRxKbps"] = gpu.PcieRxKbps
	flat[prefix+"PcieRxKbpsT"] = gpu.PcieRxKbpsT
	flat[prefix+"PowerDrawW"] = gpu.PowerDrawW
	flat[prefix+"PowerDrawWT"] = gpu.PowerDrawWT
	flat[prefix+"PerfState"] = gpu.PerfState
	flat[prefix+"PerfStateT"] = gpu.PerfStateT
	flat[prefix+"ProcessCount"] = gpu.ProcessCount
	flat[prefix+"ProcessCountT"] = gpu.ProcessCountT

	// GPU processes as JSON string
	if gpu.ProcessesJSON != "" {
		flat[prefix+"ProcessesJson"] = gpu.ProcessesJSON
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
