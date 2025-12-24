package src

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CPUMetrics contains VM-level CPU measurements.
// All time values are in centiseconds (1 cs = 10ms).
type CPUMetrics struct {
	// Aggregate CPU time
	Time       exporter.Timed[int64] `json:"vCpuTime"`            // User + Kernel time
	UserMode   exporter.Timed[int64] `json:"vCpuTimeUserMode"`    // User space execution
	KernelMode exporter.Timed[int64] `json:"vCpuTimeKernelMode"`  // Kernel execution
	Idle       exporter.Timed[int64] `json:"vCpuIdleTime"`        // Idle state
	IOWait     exporter.Timed[int64] `json:"vCpuTimeIOWait"`      // Waiting for I/O
	IRQ        exporter.Timed[int64] `json:"vCpuTimeIntSrvc"`     // Hardware interrupts
	SoftIRQ    exporter.Timed[int64] `json:"vCpuTimeSoftIntSrvc"` // Software interrupts
	Nice       exporter.Timed[int64] `json:"vCpuNice"`            // Low priority user procs
	Steal      exporter.Timed[int64] `json:"vCpuSteal"`           // Stolen by hypervisor

	// Performance counters
	ContextSwitches exporter.Timed[int64]   `json:"vCpuContextSwitches"` // Total context switches
	LoadAvg         exporter.Timed[float64] `json:"vLoadAvg"`            // 1-min load average
	FreqMHz         exporter.Timed[float64] `json:"vCpuMhz"`             // Current avg frequency
}

// CPUStatic contains static CPU information.
type CPUStatic struct {
	NumProcessors int              `json:"vNumProcessors"` // Logical CPU count
	Model         string           `json:"vCpuType"`       // Processor model
	Cache         map[string]int64 `json:"vCpuCache"`      // Cache sizes (L1d, L1i, L2, L3)
	KernelInfo    string           `json:"vKernelInfo"`    // Kernel version
}

const jiffiesPerSec = 100 // Standard Linux jiffy rate

// CollectCPU gathers CPU metrics from /proc/stat, /proc/loadavg, and sysfs.
func CollectCPU() CPUMetrics {
	stat, ts := parseProcStat()
	load, loadTS := parseLoadAvg()
	freq, freqTS := getCPUFreq()

	user := stat["user"]
	system := stat["system"]

	return CPUMetrics{
		Time:            exporter.TimedAt((user+system)*jiffiesPerSec, ts),
		UserMode:        exporter.TimedAt(user*jiffiesPerSec, ts),
		KernelMode:      exporter.TimedAt(system*jiffiesPerSec, ts),
		Idle:            exporter.TimedAt(stat["idle"]*jiffiesPerSec, ts),
		IOWait:          exporter.TimedAt(stat["iowait"]*jiffiesPerSec, ts),
		IRQ:             exporter.TimedAt(stat["irq"]*jiffiesPerSec, ts),
		SoftIRQ:         exporter.TimedAt(stat["softirq"]*jiffiesPerSec, ts),
		Nice:            exporter.TimedAt(stat["nice"]*jiffiesPerSec, ts),
		Steal:           exporter.TimedAt(stat["steal"]*jiffiesPerSec, ts),
		ContextSwitches: exporter.TimedAt(stat["ctxt"], ts),
		LoadAvg:         exporter.TimedAt(load, loadTS),
		FreqMHz:         exporter.TimedAt(freq, freqTS),
	}
}

// CollectCPUStatic gathers static CPU info.
func CollectCPUStatic() CPUStatic {
	return CPUStatic{
		NumProcessors: getNumCPU(),
		Model:         getCPUModel(),
		Cache:         getCPUCache(),
		KernelInfo:    getKernelInfo(),
	}
}

// parseProcStat parses /proc/stat for CPU and context switch counters.
func parseProcStat() (map[string]int64, int64) {
	lines, ts := exporter.readLines("/proc/stat")
	m := map[string]int64{
		"user": 0, "nice": 0, "system": 0, "idle": 0,
		"iowait": 0, "irq": 0, "softirq": 0, "steal": 0, "ctxt": 0,
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "cpu":
			// cpu user nice system idle iowait irq softirq steal guest guest_nice
			if len(fields) > 1 {
				m["user"], _ = strconv.ParseInt(fields[1], 10, 64)
			}
			if len(fields) > 2 {
				m["nice"], _ = strconv.ParseInt(fields[2], 10, 64)
			}
			if len(fields) > 3 {
				m["system"], _ = strconv.ParseInt(fields[3], 10, 64)
			}
			if len(fields) > 4 {
				m["idle"], _ = strconv.ParseInt(fields[4], 10, 64)
			}
			if len(fields) > 5 {
				m["iowait"], _ = strconv.ParseInt(fields[5], 10, 64)
			}
			if len(fields) > 6 {
				m["irq"], _ = strconv.ParseInt(fields[6], 10, 64)
			}
			if len(fields) > 7 {
				m["softirq"], _ = strconv.ParseInt(fields[7], 10, 64)
			}
			if len(fields) > 8 {
				m["steal"], _ = strconv.ParseInt(fields[8], 10, 64)
			}
		case "ctxt":
			if len(fields) > 1 {
				m["ctxt"], _ = strconv.ParseInt(fields[1], 10, 64)
			}
		}
	}
	return m, ts
}

// parseLoadAvg reads the 1-minute load average from /proc/loadavg.
func parseLoadAvg() (float64, int64) {
	content, ts := exporter.readFile("/proc/loadavg")
	fields := strings.Fields(content)
	if len(fields) == 0 {
		return 0, ts
	}
	v, _ := strconv.ParseFloat(fields[0], 64)
	return v, ts
}

// getCPUFreq returns average CPU frequency in MHz.
// Tries sysfs first, falls back to /proc/cpuinfo.
func getCPUFreq() (float64, int64) {
	ts := nowMilli()

	// Try sysfs scaling_cur_freq (in kHz)
	pattern := "/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq"
	matches, _ := filepath.Glob(pattern)

	if len(matches) > 0 {
		var sum int64
		var count int
		for _, p := range matches {
			v, _ := exporter.readInt(p)
			if v > 0 {
				sum += v
				count++
			}
		}
		if count > 0 {
			return float64(sum) / float64(count) / 1000.0, ts
		}
	}

	// Fallback to /proc/cpuinfo
	lines, ts := exporter.readLines("/proc/cpuinfo")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu MHz") {
			idx := strings.Index(line, ":")
			if idx >= 0 {
				v, _ := strconv.ParseFloat(strings.TrimSpace(line[idx+1:]), 64)
				return v, ts
			}
		}
	}
	return 0, ts
}

func getNumCPU() int {
	// Count cpu directories in sysfs
	matches, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if len(matches) > 0 {
		return len(matches)
	}
	return 1
}

func getCPUModel() string {
	lines, _ := exporter.readLines("/proc/cpuinfo")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			idx := strings.Index(line, ":")
			if idx >= 0 {
				return strings.TrimSpace(line[idx+1:])
			}
		}
	}
	return "unknown"
}

func getKernelInfo() string {
	content, _ := exporter.readFile("/proc/version")
	return content
}

// getCPUCache reads cache info from sysfs.
// Returns map with keys like "L1d", "L1i", "L2", "L3" and values in bytes.
func getCPUCache() map[string]int64 {
	result := make(map[string]int64)
	seen := make(map[string]map[string]bool) // key -> shared_cpu_map -> seen

	pattern := "/sys/devices/system/cpu/cpu*/cache/index*"
	matches, _ := filepath.Glob(pattern)

	for _, dir := range matches {
		level, _ := exporter.readFile(filepath.Join(dir, "level"))
		ctype, _ := exporter.readFile(filepath.Join(dir, "type"))
		sizeStr, _ := exporter.readFile(filepath.Join(dir, "size"))
		cpuMap, _ := exporter.readFile(filepath.Join(dir, "shared_cpu_map"))

		if level == "" || sizeStr == "" {
			continue
		}

		// Build cache key (L1d, L1i, L2, L3)
		suffix := ""
		switch ctype {
		case "Data":
			suffix = "d"
		case "Instruction":
			suffix = "i"
		}
		key := "L" + level + suffix

		// Parse size (e.g., "32K", "256K", "8192K")
		size := parseSize(sizeStr)

		// Deduplicate by shared_cpu_map
		if seen[key] == nil {
			seen[key] = make(map[string]bool)
		}
		if !seen[key][cpuMap] {
			seen[key][cpuMap] = true
			result[key] += size
		}
	}
	return result
}

func parseSize(s string) int64 {
	s = strings.TrimSpace(s)
	mult := int64(1)
	if strings.HasSuffix(s, "K") {
		mult = 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "M") {
		mult = 1024 * 1024
		s = s[:len(s)-1]
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v * mult
}

func nowMilli() int64 {
	return time.Now().UnixMilli()
}
