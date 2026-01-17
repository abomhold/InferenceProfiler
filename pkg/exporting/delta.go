package exporting

import (
	"InferenceProfiler/pkg/utils"
)

// DeltaRecord calculates the difference between two records.
func DeltaRecord(initial, final Record, durationMs int64) Record {
	if initial == nil || final == nil {
		return final
	}

	result := make(Record, len(final)+3)

	// Add metadata
	if startTs, ok := initial["timestamp"]; ok {
		result["_delta_start_ts"] = startTs
	}
	if endTs, ok := final["timestamp"]; ok {
		result["_delta_end_ts"] = endTs
		result["timestamp"] = endTs
	}
	result["_delta_duration_ms"] = durationMs

	for key, finalVal := range final {
		if key == "timestamp" {
			continue
		}

		initialVal, hasInitial := initial[key]

		if !hasInitial {
			result[key] = finalVal
			continue
		}

		if delta, ok := computeDelta(initialVal, finalVal); ok {
			result[key] = delta
		} else {
			result[key] = finalVal
		}
	}

	// Capture keys present in initial but missing in final
	for key, initialVal := range initial {
		if _, exists := result[key]; !exists {
			result[key] = initialVal
		}
	}

	return result
}

func computeDelta(initial, final interface{}) (interface{}, bool) {
	if initInt, ok := utils.ToInt64Ok(initial); ok {
		finalInt := utils.ToInt64(final)
		return finalInt - initInt, true
	}

	if initFloat, ok := utils.ToFloat64Ok(initial); ok {
		finalFloat := utils.ToFloat64(final)
		return finalFloat - initFloat, true

	}

	return nil, false
}
