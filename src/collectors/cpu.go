package collectors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// CollectCPUStatic populates static CPU information
func CollectCPUStatic(m *StaticMetrics) {
	m.VMID = getVMID()
	m.BootTime = getBootTime()
	m.CPUType = getCPUType()
	m.NumProcessors = runtime.NumCPU()

	// CPU cache info
	cache := getCPUCache()
	m.CPUCacheL1d = cache["L1d"]
	m.CPUCacheL1i = cache["L1i"]
	m.CPUCacheL2 = cache["L2"]
	m.CPUCacheL3 = cache["L3"]

	// Kernel info
	getKernelInfo(m)
}

// CollectCPUDynamic populates dynamic CPU metrics
func CollectCPUDynamic(m *DynamicMetrics) {
	// /proc/stat metrics
	statMetrics, tStat := getProcStat()
	m.CPUTimeUserMode = statMetrics["user"]
	m.CPUTimeUserModeT = tStat
	m.CPUTimeKernelMode = statMetrics["system"]
	m.CPUTimeKernelModeT = tStat
	m.CPUIdleTime = statMetrics["idle"]
	m.CPUIdleTimeT = tStat
	m.CPUTimeIOWait = statMetrics["iowait"]
	m.CPUTimeIOWaitT = tStat
	m.CPUTimeIntSrvc = statMetrics["irq"]
	m.CPUTimeIntSrvcT = tStat
	m.CPUTimeSoftIntSrvc = statMetrics["softirq"]
	m.CPUTimeSoftIntSrvcT = tStat
	m.CPUNice = statMetrics["nice"]
	m.CPUNiceT = tStat
	m.CPUSteal = statMetrics["steal"]
	m.CPUStealT = tStat
	m.CPUTime = statMetrics["user"] + statMetrics["system"]
	m.CPUTimeT = tStat
	m.CPUContextSwitches = statMetrics["ctxt"]
	m.CPUContextSwitchesT = tStat

	// Load average
	m.LoadAvg, m.LoadAvgT = getLoadAvg()

	// CPU frequency
	m.CPUMhz, m.CPUMhzT = getCPUFreq()
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
	result := make(map[string]int64)
	seen := make(map[string]bool)

	dirs, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/cache/index*")
	for _, dir := range dirs {
		level, _ := ProbeFile(filepath.Join(dir, "level"))
		cType, _ := ProbeFile(filepath.Join(dir, "type"))
		sizeStr, _ := ProbeFile(filepath.Join(dir, "size"))
		shared, _ := ProbeFile(filepath.Join(dir, "shared_cpu_map"))

		cacheID := fmt.Sprintf("L%s-%s-%s", level, cType, shared)
		if seen[cacheID] || level == "" || sizeStr == "" {
			continue
		}
		seen[cacheID] = true

		var size int64
		var unit rune
		fmt.Sscanf(sizeStr, "%d%c", &size, &unit)
		if unit == 'K' {
			size *= 1024
		} else if unit == 'M' {
			size *= 1024 * 1024
		}

		suffix := ""
		if cType == "Data" {
			suffix = "d"
		} else if cType == "Instruction" {
			suffix = "i"
		}

		result["L"+level+suffix] += size
	}
	return result
}

func getKernelInfo(m *StaticMetrics) {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return
	}
	m.SystemName = ByteSliceToString(uname.Sysname[:])
	m.NodeName = ByteSliceToString(uname.Nodename[:])
	m.Hostname = ByteSliceToString(uname.Nodename[:])
	m.KernelRelease = ByteSliceToString(uname.Release[:])
	m.KernelVersion = ByteSliceToString(uname.Version[:])
	m.Machine = ByteSliceToString(uname.Machine[:])
}

func getBootTime() int64 {
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

func getVMID() string {
	content, _ := ProbeFile("/sys/class/dmi/id/product_uuid")
	if content != "" && content != "None" {
		return content
	}

	content, _ = ProbeFile("/etc/machine-id")
	if content != "" {
		return content
	}

	content, _ = ProbeFile("/etc/hostname")
	if content != "" {
		return content
	}

	return "unavailable"
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

		if parts[0] == "cpu" && len(parts) >= 9 {
			metrics["user"], _ = strconv.ParseInt(parts[1], 10, 64)
			metrics["nice"], _ = strconv.ParseInt(parts[2], 10, 64)
			metrics["system"], _ = strconv.ParseInt(parts[3], 10, 64)
			metrics["idle"], _ = strconv.ParseInt(parts[4], 10, 64)
			metrics["iowait"], _ = strconv.ParseInt(parts[5], 10, 64)
			metrics["irq"], _ = strconv.ParseInt(parts[6], 10, 64)
			metrics["softirq"], _ = strconv.ParseInt(parts[7], 10, 64)
			metrics["steal"], _ = strconv.ParseInt(parts[8], 10, 64)
		} else if parts[0] == "ctxt" && len(parts) >= 2 {
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
