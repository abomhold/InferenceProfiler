package collectors

import (
	"regexp"
	"strconv"
)

// MemoryCollector gathers memory metrics.
type MemoryCollector struct {
	BaseCollector
}

var (
	pgFaultRE    = regexp.MustCompile(`pgfault\s+(\d+)`)
	pgMajFaultRE = regexp.MustCompile(`pgmajfault\s+(\d+)`)
)

// Collect gathers memory metrics from /proc/meminfo and /proc/vmstat.
func (c *MemoryCollector) Collect() MemoryMetrics {
	info, ts := c.parseMemInfo()
	pgFault, pgMajFault, vmstatTS := c.parseVMStat()

	total := info["MemTotal"]
	freeRaw := info["MemFree"]
	buffers := info["Buffers"]
	cached := info["Cached"] + info["SReclaimable"]

	// Calculate available memory
	available := info["MemAvailable"]
	if available == 0 {
		available = freeRaw + buffers + cached
	}

	used := total - freeRaw - buffers - cached

	var percent float64
	if total > 0 {
		percent = float64(total-available) / float64(total) * 100
	}

	swapTotal := info["SwapTotal"]
	swapFree := info["SwapFree"]

	return MemoryMetrics{
		Total:          TimedAt(total, ts),
		Free:           TimedAt(available, ts),
		Used:           TimedAt(used, ts),
		Buffers:        TimedAt(buffers, ts),
		Cached:         TimedAt(cached, ts),
		Percent:        TimedAt(percent, ts),
		SwapTotal:      TimedAt(swapTotal, ts),
		SwapFree:       TimedAt(swapFree, ts),
		SwapUsed:       TimedAt(swapTotal-swapFree, ts),
		PageFault:      TimedAt(pgFault, vmstatTS),
		MajorPageFault: TimedAt(pgMajFault, vmstatTS),
	}
}

// GetStatic gathers static memory info.
func (c *MemoryCollector) GetStatic() MemoryStatic {
	info, _ := c.parseMemInfo()
	return MemoryStatic{
		TotalBytes:     info["MemTotal"],
		SwapTotalBytes: info["SwapTotal"],
	}
}

// parseMemInfo parses /proc/meminfo into bytes.
func (c *MemoryCollector) parseMemInfo() (map[string]int64, int64) {
	kv, ts := c.ParseKV("/proc/meminfo", ':')
	m := make(map[string]int64)
	for k, v := range kv {
		m[k] = c.ParseMemValue(v)
	}
	return m, ts
}

// parseVMStat extracts page fault counters from /proc/vmstat.
func (c *MemoryCollector) parseVMStat() (pgFault, pgMajFault int64, ts int64) {
	content, ts := c.ReadFile("/proc/vmstat")
	if content == "" {
		return 0, 0, ts
	}

	if m := pgFaultRE.FindStringSubmatch(content); len(m) > 1 {
		pgFault, _ = strconv.ParseInt(m[1], 10, 64)
	}
	if m := pgMajFaultRE.FindStringSubmatch(content); len(m) > 1 {
		pgMajFault, _ = strconv.ParseInt(m[1], 10, 64)
	}
	return pgFault, pgMajFault, ts
}
