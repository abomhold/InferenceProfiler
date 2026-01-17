package vm

import (
	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/collectors/types"
	"regexp"
	"strconv"
	"strings"
)

// MemoryCollector collects memory metrics from /proc/meminfo and /proc/vmstat
type MemoryCollector struct {
	collectors.BaseCollector
}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

func (c *MemoryCollector) Name() string {
	return "Memory"
}

func (c *MemoryCollector) CollectStatic(m *types.StaticMetrics) {
	memInfo, _ := getMeminfo()
	m.MemoryTotalBytes = memInfo["MemTotal"]
	m.SwapTotalBytes = memInfo["SwapTotal"]
}

func (c *MemoryCollector) CollectDynamic(m *types.DynamicMetrics) {
	memInfo, tMem := getMeminfo()
	pgFault, pgMajFault, tVmstat := getPageFaults()

	total := memInfo["MemTotal"]
	freeRaw := memInfo["MemFree"]
	buffers := memInfo["Buffers"]
	cached := memInfo["Cached"] + memInfo["SReclaimable"]

	var available int64
	if val, ok := memInfo["MemAvailable"]; ok {
		available = val
	} else {
		available = freeRaw + buffers + cached
	}

	used := total - freeRaw - buffers - cached

	var percent float64
	if total > 0 {
		percent = float64(total-available) / float64(total) * 100
	}

	m.MemoryTotal = total
	m.MemoryTotalT = tMem
	m.MemoryFree = available
	m.MemoryFreeT = tMem
	m.MemoryUsed = used
	m.MemoryUsedT = tMem
	m.MemoryBuffers = buffers
	m.MemoryBuffersT = tMem
	m.MemoryCached = cached
	m.MemoryCachedT = tMem
	m.MemoryPercent = percent
	m.MemoryPercentT = tMem
	m.MemoryPgFault = pgFault
	m.MemoryPgFaultT = tVmstat
	m.MemoryMajorPageFault = pgMajFault
	m.MemoryMajorPageFaultT = tVmstat
	m.MemorySwapTotal = memInfo["SwapTotal"]
	m.MemorySwapTotalT = tMem
	m.MemorySwapFree = memInfo["SwapFree"]
	m.MemorySwapFreeT = tMem
	m.MemorySwapUsed = memInfo["SwapTotal"] - memInfo["SwapFree"]
	m.MemorySwapUsedT = tMem
}

func getMeminfo() (map[string]int64, int64) {
	rawInfo, ts := collectors.ProbeFileKV("/proc/meminfo", ":")
	processed := make(map[string]int64)

	for k, v := range rawInfo {
		v = strings.TrimSuffix(v, " kB")
		v = strings.TrimSpace(v)
		val, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			processed[k] = val * 1024 // Convert kB to bytes
		}
	}
	return processed, ts
}

func getPageFaults() (int64, int64, int64) {
	content, ts := collectors.ProbeFile("/proc/vmstat")
	if content == "" {
		return 0, 0, ts
	}

	var pgFault, pgMajFault int64

	pgFaultRe := regexp.MustCompile(`pgfault\s+(\d+)`)
	pgMajFaultRe := regexp.MustCompile(`pgmajfault\s+(\d+)`)

	if match := pgFaultRe.FindStringSubmatch(content); len(match) > 1 {
		pgFault, _ = strconv.ParseInt(match[1], 10, 64)
	}
	if match := pgMajFaultRe.FindStringSubmatch(content); len(match) > 1 {
		pgMajFault, _ = strconv.ParseInt(match[1], 10, 64)
	}

	return pgFault, pgMajFault, ts
}
