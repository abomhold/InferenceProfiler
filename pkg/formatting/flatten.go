// Package formatting provides flatten utilities for export.
package formatting

import (
	"encoding/json"
)

// Special keys for deferred serialization (must match types package)
const (
	keyProcesses = "_processes"
	keyGPUs      = "_gpus"
)

// FlattenRecord processes a Record, serializing any deferred slice data.
// This should be called before writing to handle special keys like _processes and _gpus.
func FlattenRecord(r Record) Record {
	if r == nil {
		return nil
	}

	// Check if flattening is needed
	_, hasProcs := r[keyProcesses]
	_, hasGPUs := r[keyGPUs]
	if !hasProcs && !hasGPUs {
		return r
	}

	// Create new record without special keys
	result := make(Record, len(r))
	for k, v := range r {
		if k == keyProcesses || k == keyGPUs {
			continue
		}
		result[k] = v
	}

	// Serialize processes
	if procs, ok := r[keyProcesses]; ok && procs != nil {
		if data, err := json.Marshal(procs); err == nil {
			result["processesJson"] = string(data)
		}
	}

	// Serialize GPUs
	if gpus, ok := r[keyGPUs]; ok && gpus != nil {
		if data, err := json.Marshal(gpus); err == nil {
			result["nvidiaGpusDynamic"] = string(data)
		}
	}

	return result
}

// MustFlatten flattens a record, panicking on nil.
// Useful for benchmarks and tests.
func MustFlatten(r Record) Record {
	if r == nil {
		panic("nil record")
	}
	return FlattenRecord(r)
}
