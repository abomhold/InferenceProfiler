package collectors

import (
	"regexp"
	"strconv"
	"strings"
)

// CollectMemory collects memory metrics
func CollectMemory() map[string]MetricValue {
	metrics := make(map[string]MetricValue)

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

	metrics["vMemoryTotal"] = NewMetricWithTime(total, tMem)
	metrics["vMemoryFree"] = NewMetricWithTime(available, tMem)
	metrics["vMemoryUsed"] = NewMetricWithTime(used, tMem)
	metrics["vMemoryBuffers"] = NewMetricWithTime(buffers, tMem)
	metrics["vMemoryCached"] = NewMetricWithTime(cached, tMem)
	metrics["vMemoryPercent"] = NewMetricWithTime(percent, tMem)
	metrics["vPgFault"] = NewMetricWithTime(pgFault, tVmstat)
	metrics["vMajorPageFault"] = NewMetricWithTime(pgMajFault, tVmstat)
	metrics["vSwapTotal"] = NewMetricWithTime(memInfo["SwapTotal"], tMem)
	metrics["vSwapFree"] = NewMetricWithTime(memInfo["SwapFree"], tMem)
	metrics["vSwapUsed"] = NewMetricWithTime(memInfo["SwapTotal"]-memInfo["SwapFree"], tMem)

	return metrics
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

// GetMemoryStaticInfo returns static memory information
func GetMemoryStaticInfo() (int64, int64) {
	memInfo, _ := getMeminfo()
	return memInfo["MemTotal"], memInfo["SwapTotal"]
}
