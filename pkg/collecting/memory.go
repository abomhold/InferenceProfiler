package collecting

import (
	"InferenceProfiler/pkg/metrics/vm"
	"InferenceProfiler/pkg/probing"
	"strings"
)

type Memory struct{}

func NewMemory() *Memory       { return &Memory{} }
func (c *Memory) Name() string { return "VM-Memory" }
func (c *Memory) Close() error { return nil }

func (c *Memory) CollectStatic() any {
	info, _ := getMemInfo()
	return &vm.MemoryStatic{
		MemoryTotalBytes: info["MemTotal"] * 1024,
		SwapTotalBytes:   info["SwapTotal"] * 1024,
	}
}

func (c *Memory) CollectDynamic() any {
	m := &vm.MemoryDynamic{}

	info, tMem := getMemInfo()
	vmstat, tVmstat := getVMStat()

	memTotal := info["MemTotal"] * 1024
	memFree := info["MemFree"] * 1024
	memAvailable := info["MemAvailable"] * 1024
	buffers := info["Buffers"] * 1024
	cached := (info["Cached"] + info["SReclaimable"]) * 1024

	m.MemoryTotal, m.MemoryTotalT = memTotal, tMem
	m.MemoryFree, m.MemoryFreeT = memFree, tMem
	m.MemoryBuffers, m.MemoryBuffersT = buffers, tMem
	m.MemoryCached, m.MemoryCachedT = cached, tMem

	memUsed := memTotal - memAvailable
	if memAvailable == 0 {
		memUsed = memTotal - memFree - buffers - cached
	}
	m.MemoryUsed, m.MemoryUsedT = memUsed, tMem

	if memTotal > 0 {
		m.MemoryPercent = float64(memUsed) / float64(memTotal) * 100
		m.MemoryPercentT = tMem
	}

	swapTotal := info["SwapTotal"] * 1024
	swapFree := info["SwapFree"] * 1024
	m.MemorySwapTotal, m.MemorySwapTotalT = swapTotal, tMem
	m.MemorySwapFree, m.MemorySwapFreeT = swapFree, tMem
	m.MemorySwapUsed, m.MemorySwapUsedT = swapTotal-swapFree, tMem

	m.MemoryPgFault, m.MemoryPgFaultT = vmstat["pgfault"], tVmstat
	m.MemoryMajorPageFault, m.MemoryMajorPageFaultT = vmstat["pgmajfault"], tVmstat

	return m
}

func getMemInfo() (map[string]int64, int64) {
	kv, ts := probing.FileKV("/proc/meminfo", ":")
	result := make(map[string]int64)
	for k, v := range kv {
		v = strings.TrimSuffix(v, " kB")
		v = strings.TrimSpace(v)
		result[k] = probing.ParseInt64(v)
	}
	return result, ts
}

func getVMStat() (map[string]int64, int64) {
	lines, ts := probing.FileLines("/proc/vmstat")
	result := make(map[string]int64)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			result[fields[0]] = probing.ParseInt64(fields[1])
		}
	}
	return result, ts
}
