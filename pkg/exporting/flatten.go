package exporting

import (
	"encoding/json"
)

// Special keys for deferred serialization (must match types package)
const (
	keyProcesses = "_processes"
	keyGPUs      = "_gpus"
)

// FlattenRecord processes a Record, serializing any deferred slice data.
func FlattenRecord(r Record) Record {
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

	if procs, ok := r[keyProcesses]; ok && procs != nil {
		if data, err := json.Marshal(procs); err == nil {
			result["processesJson"] = string(data)
		}
	}

	if gpus, ok := r[keyGPUs]; ok && gpus != nil {
		if data, err := json.Marshal(gpus); err == nil {
			result["nvidiaGpusDynamic"] = string(data)
		}
	}

	return result
}
