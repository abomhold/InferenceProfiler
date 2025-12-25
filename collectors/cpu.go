package collectors

import (
	"path/filepath"
	"strconv"
	"strings"
)

type CPUCollector struct {
	BaseCollector
}

func (c *CPUCollector) Collect() CPUMetrics {
	stat, ts := c.parseProcStat()
	load, loadTS := c.parseLoadAvg()
	freq, freqTS := c.getCPUFreq()

	user := stat["user"]
	system := stat["system"]

	return CPUMetrics{
		Time:            TimedAt((user+system)*jiffiesPerSec, ts),
		UserMode:        TimedAt(user*jiffiesPerSec, ts),
		KernelMode:      TimedAt(system*jiffiesPerSec, ts),
		Idle:            TimedAt(stat["idle"]*jiffiesPerSec, ts),
		IOWait:          TimedAt(stat["iowait"]*jiffiesPerSec, ts),
		IRQ:             TimedAt(stat["irq"]*jiffiesPerSec, ts),
		SoftIRQ:         TimedAt(stat["softirq"]*jiffiesPerSec, ts),
		Nice:            TimedAt(stat["nice"]*jiffiesPerSec, ts),
		Steal:           TimedAt(stat["steal"]*jiffiesPerSec, ts),
		ContextSwitches: TimedAt(stat["ctxt"], ts),
		LoadAvg:         TimedAt(load, loadTS),
		FreqMHz:         TimedAt(freq, freqTS),
	}
}

// GetStatic gathers static CPU info.
func (c *CPUCollector) GetStatic() CPUStatic {
	return CPUStatic{
		NumProcessors: c.getNumCPU(),
		Model:         c.getCPUModel(),
		Cache:         c.getCPUCache(),
		KernelInfo:    c.getKernelInfo(),
	}
}

// parseProcStat parses /proc/stat for CPU and context switch counters.
func (c *CPUCollector) parseProcStat() (map[string]int64, int64) {
	lines, ts := c.ReadLines("/proc/stat")
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
			if len(fields) > 1 {
				m["user"] = c.ParseInt(fields[1])
			}
			if len(fields) > 2 {
				m["nice"] = c.ParseInt(fields[2])
			}
			if len(fields) > 3 {
				m["system"] = c.ParseInt(fields[3])
			}
			if len(fields) > 4 {
				m["idle"] = c.ParseInt(fields[4])
			}
			if len(fields) > 5 {
				m["iowait"] = c.ParseInt(fields[5])
			}
			if len(fields) > 6 {
				m["irq"] = c.ParseInt(fields[6])
			}
			if len(fields) > 7 {
				m["softirq"] = c.ParseInt(fields[7])
			}
			if len(fields) > 8 {
				m["steal"] = c.ParseInt(fields[8])
			}
		case "ctxt":
			if len(fields) > 1 {
				m["ctxt"] = c.ParseInt(fields[1])
			}
		}
	}
	return m, ts
}

// parseLoadAvg reads the 1-minute load average from /proc/loadavg.
func (c *CPUCollector) parseLoadAvg() (float64, int64) {
	content, ts := c.ReadFile("/proc/loadavg")
	fields := strings.Fields(content)
	if len(fields) == 0 {
		return 0, ts
	}
	v, _ := strconv.ParseFloat(fields[0], 64)
	return v, ts
}

// getCPUFreq returns average CPU frequency in MHz.
func (c *CPUCollector) getCPUFreq() (float64, int64) {
	ts := NowMilli()

	// Try sysfs scaling_cur_freq (in kHz)
	pattern := "/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq"
	matches, _ := filepath.Glob(pattern)

	if len(matches) > 0 {
		var sum int64
		var count int
		for _, p := range matches {
			v, _ := c.ReadInt(p)
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
	lines, ts := c.ReadLines("/proc/cpuinfo")
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

func (c *CPUCollector) getNumCPU() int {
	matches, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if len(matches) > 0 {
		return len(matches)
	}
	return 1
}

func (c *CPUCollector) getCPUModel() string {
	lines, _ := c.ReadLines("/proc/cpuinfo")
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

func (c *CPUCollector) getKernelInfo() string {
	content, _ := c.ReadFile("/proc/version")
	return content
}

func (c *CPUCollector) getCPUCache() map[string]int64 {
	result := make(map[string]int64)
	seen := make(map[string]map[string]bool)

	pattern := "/sys/devices/system/cpu/cpu*/cache/index*"
	matches, _ := filepath.Glob(pattern)

	for _, dir := range matches {
		level, _ := c.ReadFile(filepath.Join(dir, "level"))
		ctype, _ := c.ReadFile(filepath.Join(dir, "type"))
		sizeStr, _ := c.ReadFile(filepath.Join(dir, "size"))
		cpuMap, _ := c.ReadFile(filepath.Join(dir, "shared_cpu_map"))

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

		size := c.ParseSize(sizeStr)

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
