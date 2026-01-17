// Package exporting provides delta calculation utilities.
package exporting

import (
	"strings"
)

// DeltaRecord calculates the difference between two records.
// For numeric values: final - initial
// For strings/bools: keeps the final value
// Adds metadata fields: _delta_duration_ms, _delta_start_ts, _delta_end_ts
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

	// Process all keys from final record
	for key, finalVal := range final {
		initialVal, hasInitial := initial[key]

		// Skip if it's a timestamp field we already handled
		if key == "timestamp" {
			continue
		}

		// If no initial value, just use final
		if !hasInitial {
			result[key] = finalVal
			continue
		}

		// Try to compute delta for numeric values
		if delta, ok := computeDelta(initialVal, finalVal); ok {
			result[key] = delta
		} else {
			// Non-numeric: keep final value
			result[key] = finalVal
		}
	}

	// Add any keys from initial that aren't in final (shouldn't happen normally)
	for key, initialVal := range initial {
		if _, exists := result[key]; !exists {
			result[key] = initialVal
		}
	}

	return result
}

// computeDelta attempts to compute final - initial for numeric types.
// Returns (delta, true) on success, (nil, false) if not numeric.
func computeDelta(initial, final interface{}) (interface{}, bool) {
	// Handle int64
	if initInt, ok := toInt64Safe(initial); ok {
		if finalInt, ok := toInt64Safe(final); ok {
			return finalInt - initInt, true
		}
	}

	// Handle float64
	if initFloat, ok := toFloat64Safe(initial); ok {
		if finalFloat, ok := toFloat64Safe(final); ok {
			return finalFloat - initFloat, true
		}
	}

	return nil, false
}

// toInt64Safe attempts to convert a value to int64.
func toInt64Safe(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return int64(n), true
	case float64:
		// Only convert if it's a whole number
		if n == float64(int64(n)) {
			return int64(n), true
		}
		return 0, false
	}
	return 0, false
}

// toFloat64Safe attempts to convert a value to float64.
func toFloat64Safe(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// IsCounterField returns true if the field name suggests it's a cumulative counter.
// These fields are good candidates for delta calculation.
func IsCounterField(name string) bool {
	counterSuffixes := []string{
		"Time", "Bytes", "Count", "Switches", "Fault", "Reads", "Writes",
		"Sent", "Recvd", "Errors", "Drops", "Packets",
	}

	for _, suffix := range counterSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	return false
}

// DeltaRecordSelective calculates delta only for counter fields,
// keeping gauge values as-is from the final record.
func DeltaRecordSelective(initial, final Record, durationMs int64) Record {
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

	// Process all keys from final record
	for key, finalVal := range final {
		if key == "timestamp" {
			continue
		}

		initialVal, hasInitial := initial[key]

		// Only compute delta for counter fields
		if hasInitial && IsCounterField(key) {
			if delta, ok := computeDelta(initialVal, finalVal); ok {
				result[key] = delta
				continue
			}
		}

		// For non-counter fields or failed delta, use final value
		result[key] = finalVal
	}

	return result
}
