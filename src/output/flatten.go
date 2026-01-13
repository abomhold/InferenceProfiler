package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"InferenceProfiler/src/collectors"
)

// FlattenMetrics converts DynamicMetrics to a slice of flat records.
// Each record contains all base metrics plus one GPU and optionally processes.
// If there are no GPUs, returns a single record with base metrics only.
func FlattenMetrics(m *collectors.DynamicMetrics) []map[string]interface{} {
	// Start with base metrics (convert struct to map)
	base := structToMap(m)

	// Remove nested arrays that we'll flatten separately
	delete(base, "NvidiaGPUs")
	delete(base, "Processes")

	// If no GPUs, return single record
	if len(m.NvidiaGPUs) == 0 {
		// Still include process metrics if available
		if len(m.Processes) > 0 {
			flattenProcesses(base, m.Processes)
		}
		return []map[string]interface{}{base}
	}

	// Create one record per GPU
	records := make([]map[string]interface{}, 0, len(m.NvidiaGPUs))
	for i, gpu := range m.NvidiaGPUs {
		record := copyMap(base)

		// Add GPU index
		record["gpuIndex"] = i

		// Flatten GPU metrics with "gpu" prefix
		flattenGPU(record, &gpu, "gpu")

		// Add process metrics (same for all GPUs)
		if len(m.Processes) > 0 {
			flattenProcesses(record, m.Processes)
		}

		records = append(records, record)
	}

	return records
}

// flattenGPU adds GPU metrics to the record with the given prefix
func flattenGPU(record map[string]interface{}, gpu *collectors.NvidiaGPUDynamic, prefix string) {
	// Use reflection to iterate over all fields
	v := reflect.ValueOf(gpu).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get JSON tag name or use field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}

		// Handle special nested types
		switch field.Name {
		case "Processes":
			// Skip - handled separately
			continue
		case "ProcessUtilizations":
			// Skip - handled separately
			continue
		case "NvLinkBandwidth":
			// Flatten NVLink bandwidth array
			if !value.IsNil() {
				for j := 0; j < value.Len(); j++ {
					nvlink := value.Index(j).Interface().(collectors.NvLinkBandwidth)
					linkPrefix := fmt.Sprintf("%sNvLink%d", prefix, j)
					record[linkPrefix+"RxKB"] = nvlink.RxKB
					record[linkPrefix+"TxKB"] = nvlink.TxKB
				}
			}
			continue
		case "NvLinkErrors":
			// Flatten NVLink errors array
			if !value.IsNil() {
				for j := 0; j < value.Len(); j++ {
					nvlink := value.Index(j).Interface().(collectors.NvLinkErrors)
					linkPrefix := fmt.Sprintf("%sNvLink%d", prefix, j)
					record[linkPrefix+"CRCErrors"] = nvlink.CRCErrors
					record[linkPrefix+"ECCErrors"] = nvlink.ECCErrors
					record[linkPrefix+"ReplayErrors"] = nvlink.ReplayErrors
					record[linkPrefix+"RecoveryErrors"] = nvlink.RecoveryErrors
				}
			}
			continue
		}

		// Add field with prefix
		key := prefix + capitalize(name)
		record[key] = value.Interface()
	}
}

// flattenProcesses adds process metrics to the record
func flattenProcesses(record map[string]interface{}, procs []collectors.ProcessMetrics) {
	// Flatten up to MaxProcesses processes
	const MaxProcesses = 10

	for i, proc := range procs {
		if i >= MaxProcesses {
			break
		}

		prefix := fmt.Sprintf("proc%d", i)
		record[prefix+"Pid"] = proc.PID
		record[prefix+"CpuTimeUser"] = proc.CPUTimeUser
		record[prefix+"CpuTimeUserT"] = proc.CPUTimeUserT
		record[prefix+"CpuTimeSystem"] = proc.CPUTimeSystem
		record[prefix+"CpuTimeSystemT"] = proc.CPUTimeSystemT
		record[prefix+"NumThreads"] = proc.NumThreads
		record[prefix+"NumThreadsT"] = proc.NumThreadsT
		record[prefix+"VoluntaryCtxSwitches"] = proc.VoluntaryCtxSwitches
		record[prefix+"VoluntaryCtxSwitchesT"] = proc.VoluntaryCtxSwitchesT
		record[prefix+"NonVoluntaryCtxSwitches"] = proc.NonVoluntaryCtxSwitches
		record[prefix+"NonVoluntaryCtxSwitchesT"] = proc.NonVoluntaryCtxSwitchesT
		record[prefix+"DelayBlkIO"] = proc.DelayBlkIO
		record[prefix+"DelayBlkIOT"] = proc.DelayBlkIOT
		record[prefix+"DelaySwapin"] = proc.DelaySwapin
		record[prefix+"DelaySwapinT"] = proc.DelaySwapinT
		record[prefix+"DelayFreepages"] = proc.DelayFreepages
		record[prefix+"DelayFreepagesT"] = proc.DelayFreepagesT
		record[prefix+"MemoryRSS"] = proc.MemoryRSS
		record[prefix+"MemoryRSST"] = proc.MemoryRSST
		record[prefix+"MemoryVMS"] = proc.MemoryVMS
		record[prefix+"MemoryVMST"] = proc.MemoryVMST
	}
}

// structToMap converts a struct to a map using JSON marshaling/unmarshaling
func structToMap(v interface{}) map[string]interface{} {
	data, _ := json.Marshal(v)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// copyMap creates a shallow copy of a map
func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// capitalize returns a string with the first letter uppercased
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FlattenStaticMetrics converts StaticMetrics to a flat map for display/export
func FlattenStaticMetrics(m *collectors.StaticMetrics) map[string]interface{} {
	return structToMap(m)
}
