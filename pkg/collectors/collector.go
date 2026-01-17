package collectors

import "InferenceProfiler/pkg/collectors/types"

// Re-export types for convenience
type Record = types.Record
type Collector = types.Collector

// MergeRecords re-exports the merge function
var MergeRecords = types.MergeRecords
