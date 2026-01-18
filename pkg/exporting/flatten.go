package exporting

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Special keys for deferred serialization (must match types package)
const (
	keyProcesses = "_processes"
	keyGPUs      = "_gpus"
)

// FlattenMode controls how arrays and JSON strings are handled
type FlattenMode int

const (
	// FlattenDefault expands nvidia GPU dynamic metrics only.
	// Keeps as JSON strings: processes, disks, networkInterfaces, nvidiaGpus (static), GPU processes
	FlattenDefault FlattenMode = iota

	// FlattenAll expands everything into top-level prefixed fields (--no-json-string).
	// Expands: nvidia dynamic, processes, disks, networkInterfaces, nvidiaGpus (static), GPU processes
	FlattenAll
)

// FlattenRecord processes a Record with default flattening.
// - Expands GPU dynamic metrics: nvidia0*, nvidia1*, etc.
// - Keeps as JSON: processesJson, disks, networkInterfaces, nvidiaGpus, nvidia0ProcessesJson
func FlattenRecord(r Record) Record {
	return FlattenRecordWithMode(r, FlattenDefault)
}

// FlattenRecordExpandAll expands everything into top-level fields (--no-json-string mode).
// - Expands GPU dynamic: nvidia0*, nvidia1*, etc.
// - Expands GPU processes: nvidia0Proc0*, nvidia0Proc1*, etc.
// - Expands OS processes: proc0*, proc1*, etc.
// - Expands static: disk0*, net0*, gpuStatic0*, etc.
func FlattenRecordExpandAll(r Record) Record {
	return FlattenRecordWithMode(r, FlattenAll)
}

// FlattenRecordWithMode processes a Record based on the specified mode.
func FlattenRecordWithMode(r Record, mode FlattenMode) Record {
	if r == nil {
		return nil
	}

	_, hasProcs := r[keyProcesses]
	_, hasGPUs := r[keyGPUs]

	// Check if we have any JSON string fields that might need expansion
	hasJsonFields := false
	for k := range r {
		if strings.HasSuffix(k, "Json") || strings.HasSuffix(k, "JSON") ||
			k == "disks" || k == "networkInterfaces" || k == "nvidiaGpus" {
			hasJsonFields = true
			break
		}
	}

	if !hasProcs && !hasGPUs && !hasJsonFields {
		return r
	}

	result := make(Record, len(r))
	for k, v := range r {
		if k == keyProcesses || k == keyGPUs {
			continue
		}
		result[k] = v
	}

	// Always flatten nvidia GPU dynamic metrics into top-level fields
	if gpus, ok := r[keyGPUs]; ok && gpus != nil {
		flattenGPUsDynamic(gpus, result, mode)
	}

	// Handle OS processes based on mode
	if procs, ok := r[keyProcesses]; ok && procs != nil {
		if mode == FlattenAll {
			flattenProcesses(procs, result, "proc")
		} else {
			// Default: serialize to JSON string
			if data, err := json.Marshal(procs); err == nil {
				result["processesJson"] = string(data)
			}
		}
	}

	// Handle static JSON strings based on mode
	if mode == FlattenAll {
		expandStaticJsonFields(result)
	}

	return result
}

// flattenGPUsDynamic expands GPU dynamic metrics into top-level fields with nvidia{index} prefix.
func flattenGPUsDynamic(gpus interface{}, result Record, mode FlattenMode) {
	slice := toSlice(gpus)
	if slice == nil {
		return
	}

	for _, gpu := range slice {
		gpuMap := toMap(gpu)
		if gpuMap == nil {
			continue
		}

		index := 0
		if idx, ok := gpuMap["index"]; ok {
			index = toInt(idx)
		}
		prefix := fmt.Sprintf("nvidia%d", index)

		for k, v := range gpuMap {
			if k == "index" {
				continue
			}

			// Handle nested GPU processes
			if k == "processesJson" && v != nil {
				handleGPUProcesses(v, result, prefix, mode)
				continue
			}
			if k == "processUtilizationJson" && v != nil {
				handleGPUProcessUtilization(v, result, prefix, mode)
				continue
			}
			// Handle nvlink JSON fields
			if (k == "nvlinkBandwidthJson" || k == "nvlinkErrorsJson") && v != nil {
				handleNvlinkJson(k, v, result, prefix, mode)
				continue
			}

			// Capitalize first letter for camelCase: nvidia0UtilizationGpu
			fieldName := prefix + capitalizeFirst(k)
			result[fieldName] = v
		}
	}
}

// handleGPUProcesses handles the processesJson field within a GPU.
func handleGPUProcesses(v interface{}, result Record, gpuPrefix string, mode FlattenMode) {
	if mode == FlattenAll {
		// Parse JSON and flatten: nvidia0Proc0Pid, nvidia0Proc1UsedMemoryBytes
		procs := parseJSONStringToSlice(v)
		if procs != nil {
			flattenProcesses(procs, result, gpuPrefix+"Proc")
		}
	} else {
		// Default: keep as JSON string: nvidia0ProcessesJson
		if str, ok := v.(string); ok && str != "" {
			result[gpuPrefix+"ProcessesJson"] = str
		}
	}
}

// handleGPUProcessUtilization handles the processUtilizationJson field within a GPU.
func handleGPUProcessUtilization(v interface{}, result Record, gpuPrefix string, mode FlattenMode) {
	if mode == FlattenAll {
		// Parse JSON and flatten: nvidia0ProcUtil0SmUtil, etc.
		utils := parseJSONStringToSlice(v)
		if utils != nil {
			for i, util := range utils {
				utilMap := toMap(util)
				if utilMap == nil {
					continue
				}
				prefix := fmt.Sprintf("%sProcUtil%d", gpuPrefix, i)
				for k, val := range utilMap {
					result[prefix+capitalizeFirst(k)] = val
				}
			}
		}
	} else {
		// Default: keep as JSON string
		if str, ok := v.(string); ok && str != "" {
			result[gpuPrefix+"ProcessUtilizationJson"] = str
		}
	}
}

// handleNvlinkJson handles nvlink bandwidth and error JSON fields.
func handleNvlinkJson(key string, v interface{}, result Record, gpuPrefix string, mode FlattenMode) {
	if mode == FlattenAll {
		items := parseJSONStringToSlice(v)
		if items != nil {
			var prefix string
			if key == "nvlinkBandwidthJson" {
				prefix = gpuPrefix + "Nvlink"
			} else {
				prefix = gpuPrefix + "NvlinkErr"
			}
			for i, item := range items {
				itemMap := toMap(item)
				if itemMap == nil {
					continue
				}
				itemPrefix := fmt.Sprintf("%s%d", prefix, i)
				for k, val := range itemMap {
					if k != "link" { // skip redundant link index
						result[itemPrefix+capitalizeFirst(k)] = val
					}
				}
			}
		}
	} else {
		// Default: keep as JSON string
		if str, ok := v.(string); ok && str != "" {
			fieldName := gpuPrefix + capitalizeFirst(key)
			result[fieldName] = str
		}
	}
}

// expandStaticJsonFields parses and expands static JSON string fields when in FlattenAll mode.
func expandStaticJsonFields(result Record) {
	// Expand disks: disk0Name, disk0Size, etc.
	if disksJson, ok := result["disks"].(string); ok && disksJson != "" {
		if disks := parseJSONStringToSlice(interface{}(disksJson)); disks != nil {
			for i, disk := range disks {
				diskMap := toMap(disk)
				if diskMap == nil {
					continue
				}
				prefix := fmt.Sprintf("disk%d", i)
				for k, v := range diskMap {
					result[prefix+capitalizeFirst(k)] = v
				}
			}
			delete(result, "disks")
		}
	}

	// Expand networkInterfaces: net0Name, net0Mac, etc.
	if netJson, ok := result["networkInterfaces"].(string); ok && netJson != "" {
		if nets := parseJSONStringToSlice(interface{}(netJson)); nets != nil {
			for i, net := range nets {
				netMap := toMap(net)
				if netMap == nil {
					continue
				}
				prefix := fmt.Sprintf("net%d", i)
				for k, v := range netMap {
					result[prefix+capitalizeFirst(k)] = v
				}
			}
			delete(result, "networkInterfaces")
		}
	}

	// Expand nvidiaGpus (static GPU info): gpuStatic0Name, gpuStatic0MemoryTotalBytes, etc.
	if gpusJson, ok := result["nvidiaGpus"].(string); ok && gpusJson != "" {
		if gpus := parseJSONStringToSlice(interface{}(gpusJson)); gpus != nil {
			for i, gpu := range gpus {
				gpuMap := toMap(gpu)
				if gpuMap == nil {
					continue
				}
				// Use index from the data if available
				idx := i
				if index, ok := gpuMap["index"]; ok {
					idx = toInt(index)
				}
				prefix := fmt.Sprintf("gpuStatic%d", idx)
				for k, v := range gpuMap {
					if k != "index" { // skip index, it's in the prefix
						result[prefix+capitalizeFirst(k)] = v
					}
				}
			}
			delete(result, "nvidiaGpus")
		}
	}
}

// flattenProcesses expands a process slice into top-level fields with prefix.
// prefix is "proc" for OS processes or "nvidia0Proc" for GPU processes.
func flattenProcesses(procs interface{}, result Record, prefix string) {
	slice := toSlice(procs)
	if slice == nil {
		return
	}

	for i, proc := range slice {
		procMap := toMap(proc)
		if procMap == nil {
			continue
		}

		procPrefix := fmt.Sprintf("%s%d", prefix, i)
		for k, v := range procMap {
			// proc0PId, proc0PName, nvidia0Proc0Pid, etc.
			fieldName := procPrefix + capitalizeFirst(k)
			result[fieldName] = v
		}
	}
}

// parseJSONStringToSlice parses a JSON string into a slice of interfaces.
func parseJSONStringToSlice(v interface{}) []interface{} {
	str, ok := v.(string)
	if !ok || str == "" {
		return nil
	}

	var result []interface{}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		return nil
	}
	return result
}

// toSlice converts an interface to a slice of interfaces.
func toSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result
	}
	return nil
}

// toMap converts an interface to a map[string]interface{}.
func toMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}

	if m, ok := v.(map[string]interface{}); ok {
		return m
	}

	// Handle typed maps through reflection
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Map {
		result := make(map[string]interface{})
		for _, key := range rv.MapKeys() {
			result[key.String()] = rv.MapIndex(key).Interface()
		}
		return result
	}

	// Handle struct
	if rv.Kind() == reflect.Struct {
		result := make(map[string]interface{})
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.PkgPath != "" { // skip unexported
				continue
			}
			jsonTag := field.Tag.Get("json")
			name := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					name = parts[0]
				}
			}
			result[name] = rv.Field(i).Interface()
		}
		return result
	}

	return nil
}

// toInt converts an interface to int.
func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
