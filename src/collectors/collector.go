package collectors

// Collector defines the interface for all metric collectors
type Collector interface {
	Name() string
	CollectStatic(m *StaticMetrics)
	CollectDynamic(m *DynamicMetrics)
	Close() error
}

// BaseCollector provides a default Close() implementation
type BaseCollector struct{}

func (BaseCollector) Close() error { return nil }
