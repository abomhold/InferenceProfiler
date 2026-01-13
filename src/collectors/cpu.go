package collectors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// CPUCollector collects CPU metrics from /proc/stat and related sources
type CPUCollector struct {
	BaseCollector
}

// NewCPUCollector creates a new CPU collector
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Name() string {
	return "CPU"
}

func (c *CPUCollector) CollectStatic(m *StaticMetrics) {
	m.VMID = getVMID()
	m.BootTime = getBootTime()
	m.CPUType = getCPUType()
	m.NumProcessors = runtime.NumCPU()
	m.CPUCache = getCPUCache()

	// Time synchronization from adjtimex syscall
	m.TimeSynced, m.TimeOffsetSeconds, m.TimeMaxErrorSeconds = getNTPInfo()

	// Kernel info
	getKernelInfo(m)
}

func (c *CPUCollector) CollectDynamic(m *DynamicMetrics) {
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

// =============================================================================
// Helper Functions
// =============================================================================

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

func getCPUCache() string {
	result := make(map[string]int64)
	seen := make(map[string]bool)

	dirs, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/cache/index*")
	for _, dir := range dirs {
		level, _ := ProbeFile(filepath.Join(dir, "level"))
		cType, _ := ProbeFile(filepath.Join(dir, "type"))
		sizeStr, _ := ProbeFile(filepath.Join(dir, "size"))
		shared, _ := ProbeFile(filepath.Join(dir, "shared_cpu_map"))

		// Generate unique ID to prevent double counting shared caches
		cacheID := fmt.Sprintf("L%s-%s-%s", level, cType, shared)
		if seen[cacheID] || level == "" || sizeStr == "" {
			continue
		}
		seen[cacheID] = true

		// Parse size (e.g., "32K", "12M")
		var size int64
		var unit rune
		fmt.Sscanf(sizeStr, "%d%c", &size, &unit)
		switch unit {
		case 'K':
			size *= 1024
		case 'M':
			size *= 1024 * 1024
		}

		// Determine suffix for L1 (Data vs Instruction)
		suffix := ""
		if level == "1" {
			switch cType {
			case "Data":
				suffix = "d"
			case "Instruction":
				suffix = "i"
			}
		}

		result["L"+level+suffix] += size
	}

	// Format into sorted, readable string
	var parts []string
	order := []string{"L1d", "L1i", "L2", "L3", "L4"}

	for _, label := range order {
		if size, ok := result[label]; ok && size > 0 {
			var formattedSize string
			if size >= 1048576 { // >= 1MB
				formattedSize = fmt.Sprintf("%dM", size/1048576)
			} else {
				formattedSize = fmt.Sprintf("%dK", size/1024)
			}
			parts = append(parts, fmt.Sprintf("%s:%s", label, formattedSize))
		}
	}

	// Check for any keys not in standard order
	for k, size := range result {
		found := false
		for _, o := range order {
			if k == o {
				found = true
				break
			}
		}
		if !found {
			parts = append(parts, fmt.Sprintf("%s:%d", k, size))
		}
	}

	return strings.Join(parts, " ")
}

func getKernelInfo(m *StaticMetrics) {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return
	}

	toString := func(data any) string {
		var b []byte
		switch v := data.(type) {
		case [65]int8:
			for _, c := range v {
				b = append(b, byte(c))
			}
		case [65]uint8:
			b = v[:]
		}
		return unix.ByteSliceToString(b)
	}

	m.KernelInfo = fmt.Sprintf("%s %s %s %s %s",
		toString(uname.Sysname),
		toString(uname.Nodename),
		toString(uname.Release),
		toString(uname.Version),
		toString(uname.Machine))
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

// getNTPInfo performs the adjtimex syscall to get kernel clock status
func getNTPInfo() (synced bool, offset float64, maxErr float64) {
	tx := &unix.Timex{}

	state, err := unix.Adjtimex(tx)
	if err != nil {
		return false, 0, 0
	}

	// TIME_ERROR (5) indicates the clock is not synchronized
	isSynced := state != unix.TIME_ERROR

	// tx.Offset is in microseconds, convert to seconds
	offsetSeconds := float64(tx.Offset) / 1_000_000.0
	maxErrorSeconds := float64(tx.Maxerror) / 1_000_000.0

	return isSynced, offsetSeconds, maxErrorSeconds
}
