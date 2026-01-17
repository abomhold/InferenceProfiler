package memory

import "InferenceProfiler/pkg/collectors/types"

// Static contains static memory information.
type Static struct {
	MemoryTotalBytes int64
	SwapTotalBytes   int64
}

// ToRecord converts Static to a Record.
func (s *Static) ToRecord() types.Record {
	return types.Record{
		"vMemoryTotalBytes": s.MemoryTotalBytes,
		"vSwapTotalBytes":   s.SwapTotalBytes,
	}
}

// Dynamic contains dynamic memory metrics.
type Dynamic struct {
	MemoryTotal           int64
	MemoryTotalT          int64
	MemoryFree            int64
	MemoryFreeT           int64
	MemoryUsed            int64
	MemoryUsedT           int64
	MemoryBuffers         int64
	MemoryBuffersT        int64
	MemoryCached          int64
	MemoryCachedT         int64
	MemoryPercent         float64
	MemoryPercentT        int64
	MemorySwapTotal       int64
	MemorySwapTotalT      int64
	MemorySwapFree        int64
	MemorySwapFreeT       int64
	MemorySwapUsed        int64
	MemorySwapUsedT       int64
	MemoryPgFault         int64
	MemoryPgFaultT        int64
	MemoryMajorPageFault  int64
	MemoryMajorPageFaultT int64
}

// ToRecord converts Dynamic to a Record with pre-allocated capacity.
func (d *Dynamic) ToRecord() types.Record {
	return types.Record{
		"vMemoryTotal":           d.MemoryTotal,
		"vMemoryTotalT":          d.MemoryTotalT,
		"vMemoryFree":            d.MemoryFree,
		"vMemoryFreeT":           d.MemoryFreeT,
		"vMemoryUsed":            d.MemoryUsed,
		"vMemoryUsedT":           d.MemoryUsedT,
		"vMemoryBuffers":         d.MemoryBuffers,
		"vMemoryBuffersT":        d.MemoryBuffersT,
		"vMemoryCached":          d.MemoryCached,
		"vMemoryCachedT":         d.MemoryCachedT,
		"vMemoryPercent":         d.MemoryPercent,
		"vMemoryPercentT":        d.MemoryPercentT,
		"vMemorySwapTotal":       d.MemorySwapTotal,
		"vMemorySwapTotalT":      d.MemorySwapTotalT,
		"vMemorySwapFree":        d.MemorySwapFree,
		"vMemorySwapFreeT":       d.MemorySwapFreeT,
		"vMemorySwapUsed":        d.MemorySwapUsed,
		"vMemorySwapUsedT":       d.MemorySwapUsedT,
		"vMemoryPgFault":         d.MemoryPgFault,
		"vMemoryPgFaultT":        d.MemoryPgFaultT,
		"vMemoryMajorPageFault":  d.MemoryMajorPageFault,
		"vMemoryMajorPageFaultT": d.MemoryMajorPageFaultT,
	}
}
