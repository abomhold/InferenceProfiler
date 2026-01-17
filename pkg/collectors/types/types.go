// Package types provides shared types for the collectors package.
package types

// Record is a flat map of metric names to values.
// Using type alias for compatibility with formatting.Record.
type Record = map[string]interface{}

// Collector defines the interface for all metric collectors.
type Collector interface {
	Name() string
	CollectStatic() Record
	CollectDynamic() Record
	Close() error
}

// MergeRecords combines multiple Record maps into one.
// Later records override earlier ones for duplicate keys.
func MergeRecords(records ...Record) Record {
	// Count total keys for capacity hint
	total := 0
	for _, r := range records {
		total += len(r)
	}

	result := make(Record, total)
	for _, r := range records {
		for k, v := range r {
			result[k] = v
		}
	}
	return result
}
