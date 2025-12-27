package collectors

import (
	"regexp"
	"strconv"
	"strings"
)

// --- Static Metrics ---

func GetMemoryStaticInfo() StaticMetrics {
	memInfo, _ := getMeminfo()
	return StaticMetrics{
		"vMemoryTotalBytes": memInfo["MemTotal"],
		"vSwapTotalBytes":   memInfo["SwapTotal"],
	}
}

// --- Dynamic Metrics ---

func CollectMemoryDynamic() DynamicMetrics {
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

	return DynamicMetrics{
		"vMemoryTotal":          NewMetricWithTime(total, tMem),
		"vMemoryFree":           NewMetricWithTime(available, tMem),
		"vMemoryUsed":           NewMetricWithTime(used, tMem),
		"vMemoryBuffers":        NewMetricWithTime(buffers, tMem),
		"vMemoryCached":         NewMetricWithTime(cached, tMem),
		"vMemoryPercent":        NewMetricWithTime(percent, tMem),
		"vMemoryPgFault":        NewMetricWithTime(pgFault, tVmstat),
		"vMemoryMajorPageFault": NewMetricWithTime(pgMajFault, tVmstat),
		"vMemorySwapTotal":      NewMetricWithTime(memInfo["SwapTotal"], tMem),
		"vMemorySwapFree":       NewMetricWithTime(memInfo["SwapFree"], tMem),
		"vMemorySwapUsed":       NewMetricWithTime(memInfo["SwapTotal"]-memInfo["SwapFree"], tMem),
	}
}

func getMeminfo() (map[string]int64, int64) {
	rawInfo, ts := ParseProcKV("/proc/meminfo", ":")
	processed := make(map[string]int64)

	for k, v := range rawInfo {
		v = strings.TrimSuffix(v, " kB")
		v = strings.TrimSpace(v)
		val, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			processed[k] = val * 1024
		}
	}
	return processed, ts
}

func getPageFaults() (int64, int64, int64) {
	content, ts := ProbeFile("/proc/vmstat")
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
