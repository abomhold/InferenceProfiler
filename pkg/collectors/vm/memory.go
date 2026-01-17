package vm

import (
	"strings"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

// Collector collects memory metrics.
type NetworkCollector struct{}

// New creates a new Memory collector.
func NewMemoryCollector() *NetworkCollector {
	return &NetworkCollector{}
}

// Name returns the collector name.
func (c *NetworkCollector) Name() string {
	return "VM-Memory"
}

// Close releases any resources.
func (c *NetworkCollector) Close() error {
	return nil
}

// CollectStatic collects static memory information.
func (c *NetworkCollector) CollectStatic() types.Record {
	info, _ := getMemInfo()
	s := &Static{
		MemoryTotalBytes: info["MemTotal"] * 1024,
		SwapTotalBytes:   info["SwapTotal"] * 1024,
	}
	return s.ToRecord()
}

// CollectDynamic collects dynamic memory metrics.
func (c *NetworkCollector) CollectDynamic() types.Record {
	d := &Dynamic{}

	info, tMem := getMemInfo()
	vmstat, tVmstat := getVMStat()

	memTotal := info["MemTotal"] * 1024
	memFree := info["MemFree"] * 1024
	memAvailable := info["MemAvailable"] * 1024
	buffers := info["Buffers"] * 1024
	cached := (info["Cached"] + info["SReclaimable"]) * 1024

	d.MemoryTotal, d.MemoryTotalT = memTotal, tMem
	d.MemoryFree, d.MemoryFreeT = memFree, tMem
	d.MemoryBuffers, d.MemoryBuffersT = buffers, tMem
	d.MemoryCached, d.MemoryCachedT = cached, tMem

	memUsed := memTotal - memAvailable
	if memAvailable == 0 {
		memUsed = memTotal - memFree - buffers - cached
	}
	d.MemoryUsed, d.MemoryUsedT = memUsed, tMem

	if memTotal > 0 {
		d.MemoryPercent = float64(memUsed) / float64(memTotal) * 100
		d.MemoryPercentT = tMem
	}

	swapTotal := info["SwapTotal"] * 1024
	swapFree := info["SwapFree"] * 1024
	d.MemorySwapTotal, d.MemorySwapTotalT = swapTotal, tMem
	d.MemorySwapFree, d.MemorySwapFreeT = swapFree, tMem
	d.MemorySwapUsed, d.MemorySwapUsedT = swapTotal-swapFree, tMem

	d.MemoryPgFault, d.MemoryPgFaultT = vmstat["pgfault"], tVmstat
	d.MemoryMajorPageFault, d.MemoryMajorPageFaultT = vmstat["pgmajfault"], tVmstat

	return d.ToRecord()
}

func getMemInfo() (map[string]int64, int64) {
	kv, ts := probing.FileKV(config.ProcMeminfo, ":")
	result := make(map[string]int64)
	for k, v := range kv {
		v = strings.TrimSuffix(v, " kB")
		v = strings.TrimSpace(v)
		result[k] = probing.ParseInt64(v)
	}
	return result, ts
}

func getVMStat() (map[string]int64, int64) {
	lines, ts := probing.FileLines(config.ProcVMStat)
	result := make(map[string]int64)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			result[fields[0]] = probing.ParseInt64(fields[1])
		}
	}
	return result, ts
}
