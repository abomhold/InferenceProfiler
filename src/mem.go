package src

import (
	"regexp"
	"strconv"

	"github.com/inference-profiler/utils"
)

// MemoryMetrics contains VM-level memory measurements.
// All sizes are in bytes.
type MemoryMetrics struct {
	// Physical memory
	Total   Timed[int64]            `json:"vMemoryTotal"`   // Total RAM installed
	Free    utils.Timed[int64]      `json:"vMemoryFree"`    // Available memory
	Used    utils.Timed[int64]      `json:"vMemoryUsed"`    // Actively used
	Buffers exporter.Timed[int64]   `json:"vMemoryBuffers"` // Kernel buffers
	Cached  exporter.Timed[int64]   `json:"vMemoryCached"`  // Page cache
	Percent exporter.Timed[float64] `json:"vMemoryPercent"` // Usage percentage

	// Swap space
	SwapTotal exporter.Timed[int64] `json:"vSwapTotal"` // Total swap
	SwapFree  exporter.Timed[int64] `json:"vSwapFree"`  // Unused swap
	SwapUsed  exporter.Timed[int64] `json:"vSwapUsed"`  // Used swap

	// Page faults
	PageFault      exporter.Timed[int64] `json:"vPgFault"`        // Minor page faults
	MajorPageFault exporter.Timed[int64] `json:"vMajorPageFault"` // Major (disk) faults
}

// MemoryStatic contains static memory information.
type MemoryStatic struct {
	TotalBytes     int64 `json:"vMemoryTotalBytes"` // Total RAM
	SwapTotalBytes int64 `json:"vSwapTotalBytes"`   // Total swap
}

var (
	pgFaultRE    = regexp.MustCompile(`pgfault\s+(\d+)`)
	pgMajFaultRE = regexp.MustCompile(`pgmajfault\s+(\d+)`)
)

// CollectMemory gathers memory metrics from /proc/meminfo and /proc/vmstat.
func CollectMemory() MemoryMetrics {
	info, ts := parseMemInfo()
	pgFault, pgMajFault, vmstatTS := parseVMStat()

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
		Total:          exporter.TimedAt(total, ts),
		Free:           exporter.TimedAt(available, ts),
		Used:           exporter.TimedAt(used, ts),
		Buffers:        exporter.TimedAt(buffers, ts),
		Cached:         exporter.TimedAt(cached, ts),
		Percent:        exporter.TimedAt(percent, ts),
		SwapTotal:      exporter.TimedAt(swapTotal, ts),
		SwapFree:       exporter.TimedAt(swapFree, ts),
		SwapUsed:       exporter.TimedAt(swapTotal-swapFree, ts),
		PageFault:      exporter.TimedAt(pgFault, vmstatTS),
		MajorPageFault: exporter.TimedAt(pgMajFault, vmstatTS),
	}
}

// CollectMemoryStatic gathers static memory info.
func CollectMemoryStatic() MemoryStatic {
	info, _ := parseMemInfo()
	return MemoryStatic{
		TotalBytes:     info["MemTotal"],
		SwapTotalBytes: info["SwapTotal"],
	}
}

// parseMemInfo parses /proc/meminfo into bytes.
func parseMemInfo() (map[string]int64, int64) {
	kv, ts := utils.parseKV("/proc/meminfo", ':')
	m := make(map[string]int64)
	for k, v := range kv {
		m[k] = utils.parseMemValue(v)
	}
	return m, ts
}

// parseVMStat extracts page fault counters from /proc/vmstat.
func parseVMStat() (pgFault, pgMajFault int64, ts int64) {
	content, ts := utils.readFile("/proc/vmstat")
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
