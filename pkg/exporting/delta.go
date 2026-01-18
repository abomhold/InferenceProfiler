package exporting

import (
	"InferenceProfiler/pkg/utils"
	"strings"
)

// DeltaError represents an error computing delta for a field
type DeltaError struct {
	Field   string
	Reason  string
	Initial interface{}
	Final   interface{}
}

// DeltaResult holds the delta record and any errors encountered
type DeltaResult struct {
	Record Record
	Errors []DeltaError
}

// DeltaRecord calculates the difference between two records.
func DeltaRecord(initial, final Record, durationMs int64) Record {
	result := DeltaRecordWithErrors(initial, final, durationMs)
	return result.Record
}

// DeltaRecordWithErrors calculates delta and tracks errors for non-subtractable fields.
func DeltaRecordWithErrors(initial, final Record, durationMs int64) DeltaResult {
	result := DeltaResult{
		Record: make(Record),
		Errors: []DeltaError{},
	}

	if initial == nil || final == nil {
		if final != nil {
			result.Record = final
		}
		return result
	}

	// Add metadata
	if startTs, ok := initial["timestamp"]; ok {
		result.Record["_delta_start_ts"] = startTs
	}
	if endTs, ok := final["timestamp"]; ok {
		result.Record["_delta_end_ts"] = endTs
		result.Record["timestamp"] = endTs
	}
	result.Record["_delta_duration_ms"] = durationMs

	// Process all keys from final record
	for key, finalVal := range final {
		if key == "timestamp" {
			continue
		}

		initialVal, hasInitial := initial[key]

		// Skip non-subtractable fields (JSON strings, certain metadata)
		if shouldSkipDelta(key, finalVal) {
			result.Record[key] = finalVal
			continue
		}

		if !hasInitial {
			result.Record[key] = finalVal
			result.Errors = append(result.Errors, DeltaError{
				Field:  key,
				Reason: "missing_in_initial",
				Final:  finalVal,
			})
			continue
		}

		delta, err := computeDeltaWithError(key, initialVal, finalVal)
		if err != nil {
			result.Errors = append(result.Errors, *err)
			result.Record[key] = finalVal // Keep final value on error
		} else {
			result.Record[key] = delta
		}
	}

	// Capture keys present in initial but missing in final
	for key, initialVal := range initial {
		if _, exists := result.Record[key]; !exists {
			result.Record[key] = initialVal
			result.Errors = append(result.Errors, DeltaError{
				Field:   key,
				Reason:  "missing_in_final",
				Initial: initialVal,
			})
		}
	}

	return result
}

// DeltaRecordFlattened flattens both records before computing delta.
func DeltaRecordFlattened(initial, final Record, durationMs int64) DeltaResult {
	flatInitial := FlattenRecord(initial)
	flatFinal := FlattenRecord(final)
	return DeltaRecordWithErrors(flatInitial, flatFinal, durationMs)
}

// shouldSkipDelta returns true for fields that shouldn't have delta computed
func shouldSkipDelta(key string, val interface{}) bool {
	// Skip JSON string fields
	if strings.HasSuffix(key, "Json") || strings.HasSuffix(key, "JSON") {
		return true
	}
	// Skip timestamp marker fields (end with T)
	if strings.HasSuffix(key, "T") {
		return true
	}
	// Skip string fields
	if _, ok := val.(string); ok {
		return true
	}
	// Skip boolean fields
	if _, ok := val.(bool); ok {
		return true
	}
	return false
}

func computeDelta(initial, final interface{}) (interface{}, bool) {
	delta, err := computeDeltaWithError("", initial, final)
	return delta, err == nil
}

func computeDeltaWithError(key string, initial, final interface{}) (interface{}, *DeltaError) {
	// Try int64 first
	if initInt, ok := utils.ToInt64Ok(initial); ok {
		if finalInt, ok := utils.ToInt64Ok(final); ok {
			return finalInt - initInt, nil
		}
		return nil, &DeltaError{
			Field:   key,
			Reason:  "type_mismatch_int",
			Initial: initial,
			Final:   final,
		}
	}

	// Try float64
	if initFloat, ok := utils.ToFloat64Ok(initial); ok {
		if finalFloat, ok := utils.ToFloat64Ok(final); ok {
			return finalFloat - initFloat, nil
		}
		return nil, &DeltaError{
			Field:   key,
			Reason:  "type_mismatch_float",
			Initial: initial,
			Final:   final,
		}
	}

	return nil, &DeltaError{
		Field:   key,
		Reason:  "not_numeric",
		Initial: initial,
		Final:   final,
	}
}
