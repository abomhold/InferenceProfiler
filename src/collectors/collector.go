package collectors

import "InferenceProfiler/src/collectors/types"

// Collector defines the interface for all metric collectors
type Collector interface {
	// Name returns a human-readable name for this collector
	Name() string

	// CollectStatic gathers one-time system information at startup
	// Implementations should gracefully handle missing data
	CollectStatic(m *types.StaticMetrics)

	// CollectDynamic gathers periodic metrics during profiling
	// Implementations should gracefully handle missing data
	CollectDynamic(m *types.DynamicMetrics)

	// Close releases any resources held by the collector
	// Called once when profiling ends
	Close() error
}

// BaseCollector provides a default Close() implementation
// Embed this in collectors that don't need cleanup
type BaseCollector struct{}

func (BaseCollector) Close() error { return nil }
