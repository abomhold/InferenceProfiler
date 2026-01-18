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

// FlattenMode controls how arrays are handled
type FlattenMode int

const (
	// FlattenAll expands all arrays into top-level prefixed fields
	// nvidia0UtilizationGpu, nvidia0proc0Pid, proc0PId, etc.
	FlattenAll FlattenMode = iota

	// FlattenNvidiaOnly expands nvidia GPUs but keeps processes as JSON strings
	// nvidia0UtilizationGpu, nvidia0proc0Pid (expanded), processesJson (string)
	FlattenNvidiaOnly
)

// FlattenRecord processes a Record with full flattening (default behavior).
// - Expands GPU dynamic metrics: nvidia0*, nvidia1*, etc.
// - Expands GPU processes: nvidia0proc0*, nvidia0proc1*, etc.
// - Expands OS processes: proc0*, proc1*, etc.
func FlattenRecord(r Record) Record {
	return FlattenRecordWithMode(r, FlattenAll)
}

// FlattenRecordNoProcesses flattens nvidia but keeps processes as JSON string.
// Use this when --no-flatten is specified (processes become JSON, nvidia always expanded).
func FlattenRecordNoProcesses(r Record) Record {
	return FlattenRecordWithMode(r, FlattenNvidiaOnly)
}

// FlattenRecordWithMode processes a Record based on the specified mode.
func FlattenRecordWithMode(r Record, mode FlattenMode) Record {
	if r == nil {
		return nil
	}

	_, hasProcs := r[keyProcesses]
	_, hasGPUs := r[keyGPUs]
	if !hasProcs && !hasGPUs {
		return r
	}

	result := make(Record, len(r))
	for k, v := range r {
		if k == keyProcesses || k == keyGPUs {
			continue
		}
		result[k] = v
	}

	// Always flatten nvidia GPUs into top-level fields
	if gpus, ok := r[keyGPUs]; ok && gpus != nil {
		flattenGPUs(gpus, result, mode)
	}

	// Processes: either flatten or serialize to JSON based on mode
	if procs, ok := r[keyProcesses]; ok && procs != nil {
		if mode == FlattenAll {
			flattenProcesses(procs, result, "proc")
		} else {
			// FlattenNvidiaOnly: serialize processes to JSON string
			if data, err := json.Marshal(procs); err == nil {
				result["processesJson"] = string(data)
			}
		}
	}

	return result
}

// flattenGPUs expands GPU dynamic metrics into top-level fields with nvidia{index} prefix.
func flattenGPUs(gpus interface{}, result Record, mode FlattenMode) {
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
			index = idx.(int)
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

			// Capitalize first letter for camelCase: nvidia0UtilizationGpu
			fieldName := prefix + capitalizeFirst(k)
			result[fieldName] = v
		}
	}
}

// handleGPUProcesses handles the processesJson field within a GPU.
func handleGPUProcesses(v interface{}, result Record, gpuPrefix string, mode FlattenMode) {
	switch mode {
	case FlattenAll:
		// Parse JSON and flatten: nvidia0proc0Pid, nvidia0proc1UsedMemoryBytes
		procs := parseJSONStringToSlice(v)
		if procs != nil {
			flattenProcesses(procs, result, gpuPrefix+"proc")
		}
	case FlattenNvidiaOnly:
		// Keep as JSON string: nvidia0ProcessesJson
		if str, ok := v.(string); ok {
			result[gpuPrefix+"ProcessesJson"] = str
		}
	}
}

// handleGPUProcessUtilization handles the processUtilizationJson field within a GPU.
func handleGPUProcessUtilization(v interface{}, result Record, gpuPrefix string, mode FlattenMode) {
	switch mode {
	case FlattenAll:
		// Parse JSON and flatten: nvidia0procUtil0SmUtil, etc.
		utils := parseJSONStringToSlice(v)
		if utils != nil {
			for i, util := range utils {
				utilMap := toMap(util)
				if utilMap == nil {
					continue
				}
				prefix := fmt.Sprintf("%sprocUtil%d", gpuPrefix, i)
				for k, val := range utilMap {
					result[prefix+capitalizeFirst(k)] = val
				}
			}
		}
	case FlattenNvidiaOnly:
		// Keep as JSON string
		if str, ok := v.(string); ok {
			result[gpuPrefix+"ProcessUtilizationJson"] = str
		}
	}
}

// flattenProcesses expands a process slice into top-level fields with prefix.
// prefix is "proc" for OS processes or "nvidia0proc" for GPU processes.
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
			// proc0PId, proc0PName, nvidia0proc0Pid, etc.
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

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
