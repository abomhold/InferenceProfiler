package collecting

// Collector defines the interface for all metric collectors
type Collector interface {
	Name() string
	CollectStatic() any
	CollectDynamic() any
	Close() error
}
