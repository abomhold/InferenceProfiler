package collectors

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// CollectCPU collects CPU metrics
func CollectCPU() map[string]MetricValue {
	metrics := make(map[string]MetricValue)

	statMetrics, tStat := getProcStat()
	loadAvg, tLoad := getLoadAvg()
	cpuMhz, tFreq := getCPUFreq()

	metrics["vCpuTimeUserMode"] = NewMetricWithTime(statMetrics["user"]*JiffiesPerSecond, tStat)
	metrics["vCpuTimeKernelMode"] = NewMetricWithTime(statMetrics["system"]*JiffiesPerSecond, tStat)
	metrics["vCpuIdleTime"] = NewMetricWithTime(statMetrics["idle"]*JiffiesPerSecond, tStat)
	metrics["vCpuTimeIOWait"] = NewMetricWithTime(statMetrics["iowait"]*JiffiesPerSecond, tStat)
	metrics["vCpuTimeIntSrvc"] = NewMetricWithTime(statMetrics["irq"]*JiffiesPerSecond, tStat)
	metrics["vCpuTimeSoftIntSrvc"] = NewMetricWithTime(statMetrics["softirq"]*JiffiesPerSecond, tStat)
	metrics["vCpuNice"] = NewMetricWithTime(statMetrics["nice"]*JiffiesPerSecond, tStat)
	metrics["vCpuSteal"] = NewMetricWithTime(statMetrics["steal"]*JiffiesPerSecond, tStat)
	metrics["vCpuTime"] = NewMetricWithTime((statMetrics["user"]+statMetrics["system"])*JiffiesPerSecond, tStat)
	metrics["vCpuContextSwitches"] = NewMetricWithTime(statMetrics["ctxt"], tStat)
	metrics["vLoadAvg"] = NewMetricWithTime(loadAvg, tLoad)
	metrics["vCpuMhz"] = NewMetricWithTime(cpuMhz, tFreq)

	return metrics
}

func getProcStat() (map[string]int64, int64) {
	metrics := map[string]int64{
		"user": 0, "nice": 0, "system": 0, "idle": 0,
		"iowait": 0, "irq": 0, "softirq": 0, "steal": 0, "ctxt": 0,
	}

	lines, ts := ProbeFileLines("/proc/stat")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "cpu" {
			metrics["user"], _ = strconv.ParseInt(parts[1], 10, 64)
			metrics["nice"], _ = strconv.ParseInt(parts[2], 10, 64)
			metrics["system"], _ = strconv.ParseInt(parts[3], 10, 64)
			metrics["idle"], _ = strconv.ParseInt(parts[4], 10, 64)
			metrics["iowait"], _ = strconv.ParseInt(parts[5], 10, 64)
			metrics["irq"], _ = strconv.ParseInt(parts[6], 10, 64)
			metrics["softirq"], _ = strconv.ParseInt(parts[7], 10, 64)
			metrics["steal"], _ = strconv.ParseInt(parts[8], 10, 64)
		} else if parts[0] == "ctxt" {
			metrics["ctxt"], _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}
	return metrics, ts
}

func getLoadAvg() (float64, int64) {
	content, ts := ProbeFile("/proc/loadavg")
	if content == "" {
		return 0.0, ts
	}
	parts := strings.Fields(content)
	if len(parts) > 0 {
		val, _ := strconv.ParseFloat(parts[0], 64)
		return val, ts
	}
	return 0.0, ts
}

func getCPUFreq() (float64, int64) {
	ts := GetTimestamp()

	// Method 1: SysFS
	files, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq")
	if len(files) > 0 {
		var total int64
		var count int64
		for _, f := range files {
			val, _ := ProbeFileInt(f)
			if val > 0 {
				total += val
				count++
			}
		}
		if count > 0 {
			return float64(total) / float64(count) / 1000.0, ts
		}
	}

	// Method 2: /proc/cpuinfo fallback
	lines, _ := ProbeFileLines("/proc/cpuinfo")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				return val, ts
			}
		}
	}
	return 0.0, ts
}

// GetCPUStaticInfo returns static CPU information
func GetCPUStaticInfo() (int, string, map[string]int64, string) {
	numCPUs := runtime.NumCPU()
	cpuType := getCPUType()
	cpuCache := getCPUCache()
	kernelInfo := getKernelInfo()
	return numCPUs, cpuType, cpuCache, kernelInfo
}

// getKernelInfo reads kernel details from /proc/sys/kernel instead of syscalls
func getKernelInfo() string {
	sysname, _ := ProbeFile("/proc/sys/kernel/ostype")
	nodename, _ := ProbeFile("/proc/sys/kernel/hostname")
	release, _ := ProbeFile("/proc/sys/kernel/osrelease")
	version, _ := ProbeFile("/proc/sys/kernel/version")
	machine := runtime.GOARCH

	sysname = strings.TrimSpace(sysname)
	nodename = strings.TrimSpace(nodename)
	release = strings.TrimSpace(release)
	version = strings.TrimSpace(version)

	if sysname == "" {
		sysname = "Linux"
	}

	return fmt.Sprintf("%s %s %s %s %s", sysname, nodename, release, version, machine)
}

func getCPUType() string {
	lines, _ := ProbeFileLines("/proc/cpuinfo")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}

func getCPUCache() map[string]int64 {
	cacheMap := make(map[string]map[string]int64)
	result := make(map[string]int64)

	dirs, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/cache/index*")
	for _, dir := range dirs {
		level, _ := ProbeFile(filepath.Join(dir, "level"))
		cacheType, _ := ProbeFile(filepath.Join(dir, "type"))
		sizeStr, _ := ProbeFile(filepath.Join(dir, "size"))
		sharedCPUMap, _ := ProbeFile(filepath.Join(dir, "shared_cpu_map"))

		if level == "" || sizeStr == "" {
			continue
		}

		suffix := ""
		if cacheType == "Data" {
			suffix = "d"
		} else if cacheType == "Instruction" {
			suffix = "i"
		}

		key := "L" + level + suffix

		multiplier := int64(1)
		if strings.HasSuffix(sizeStr, "K") {
			multiplier = 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		} else if strings.HasSuffix(sizeStr, "M") {
			multiplier = 1024 * 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		}

		sizeBytes, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			continue
		}
		sizeBytes *= multiplier

		if cacheMap[key] == nil {
			cacheMap[key] = make(map[string]int64)
		}
		cacheMap[key][sharedCPUMap] = sizeBytes
	}

	for key, cpuMaps := range cacheMap {
		var total int64
		for _, v := range cpuMaps {
			total += v
		}
		result[key] = total
	}
	return result
}

// GetHostname returns the system hostname
func GetHostname() string {
	if name, err := os.Hostname(); err == nil {
		return name
	}

	if content, _ := ProbeFile("/proc/sys/kernel/hostname"); content != "" {
		return strings.TrimSpace(content)
	}

	return "unknown"
}

// GetBootTime returns the system boot time
func GetBootTime() int64 {
	lines, _ := ProbeFileLines("/proc/stat")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				val, _ := strconv.ParseInt(parts[1], 10, 64)
				return val
			}
		}
	}
	return 0
}

// GetVMID attempts to get VM/instance ID
func GetVMID() string {
	content, _ := ProbeFile("/sys/class/dmi/id/product_uuid")
	if content != "" && content != "None" {
		return content
	}

	content, _ = ProbeFile("/etc/machine-id")
	if content != "" {
		return content
	}

	return "unavailable"
}
