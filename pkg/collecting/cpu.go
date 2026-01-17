package collecting

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"InferenceProfiler/pkg/metrics/vm"
	"InferenceProfiler/pkg/probing"

	"golang.org/x/sys/unix"
)

type CPU struct{}

func NewCPU() *CPU          { return &CPU{} }
func (c *CPU) Name() string { return "VM-CPU" }
func (c *CPU) Close() error { return nil }

func (c *CPU) CollectStatic() any {
	m := &vm.CPUStatic{
		NumProcessors: runtime.NumCPU(),
		CPUType:       getCPUType(),
		CPUCache:      getCPUCache(),
		KernelInfo:    getKernelInfo(),
	}

	synced, offset, maxErr := getNTPInfo()
	m.TimeSynced = synced
	m.TimeOffsetSeconds = offset
	m.TimeMaxErrorSeconds = maxErr

	return m
}

func (c *CPU) CollectDynamic() any {
	m := &vm.CPUDynamic{}

	// Parse /proc/stat
	lines, tStat := probing.FileLines("/proc/stat")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if fields[0] == "cpu" && len(fields) >= 11 {
			mult := int64(1000000000 / JiffiesPerSecond)
			m.CPUTimeUserMode = probing.ParseInt64(fields[1]) * mult
			m.CPUTimeUserModeT = tStat
			m.CPUNice = probing.ParseInt64(fields[2]) * mult
			m.CPUNiceT = tStat
			m.CPUTimeKernelMode = probing.ParseInt64(fields[3]) * mult
			m.CPUTimeKernelModeT = tStat
			m.CPUIdleTime = probing.ParseInt64(fields[4]) * mult
			m.CPUIdleTimeT = tStat
			m.CPUTimeIOWait = probing.ParseInt64(fields[5]) * mult
			m.CPUTimeIOWaitT = tStat
			m.CPUTimeIntSrvc = probing.ParseInt64(fields[6]) * mult
			m.CPUTimeIntSrvcT = tStat
			m.CPUTimeSoftIntSrvc = probing.ParseInt64(fields[7]) * mult
			m.CPUTimeSoftIntSrvcT = tStat
			m.CPUSteal = probing.ParseInt64(fields[8]) * mult
			m.CPUStealT = tStat

			m.CPUTime = (probing.ParseInt64(fields[1]) + probing.ParseInt64(fields[2]) +
				probing.ParseInt64(fields[3]) + probing.ParseInt64(fields[4]) +
				probing.ParseInt64(fields[5]) + probing.ParseInt64(fields[6]) +
				probing.ParseInt64(fields[7]) + probing.ParseInt64(fields[8])) * mult
			m.CPUTimeT = tStat
		} else if fields[0] == "ctxt" && len(fields) >= 2 {
			m.CPUContextSwitches = probing.ParseInt64(fields[1])
			m.CPUContextSwitchesT = tStat
		}
	}

	// Load average
	m.LoadAvg, m.LoadAvgT = getLoadAvg()

	// CPU frequency
	m.CPUMhz, m.CPUMhzT = getCPUFreq()

	return m
}

func getCPUType() string {
	lines, _ := probing.FileLines("/proc/cpuinfo")
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
		levelVal, _ := probing.File(filepath.Join(dir, "level"))
		level := levelVal

		typeVal, _ := probing.File(filepath.Join(dir, "type"))
		cType := typeVal

		sizeVal, _ := probing.File(filepath.Join(dir, "size"))
		sizeStr := sizeVal

		sharedVal, _ := probing.File(filepath.Join(dir, "shared_cpu_map"))
		shared := sharedVal

		cacheID := fmt.Sprintf("L%s-%s-%s", level, cType, shared)
		if seen[cacheID] || level == "" || sizeStr == "" {
			continue
		}
		seen[cacheID] = true

		var size int64
		var unit rune
		_, _ = fmt.Sscanf(sizeStr, "%d%c", &size, &unit)
		switch unit {
		case 'K':
			size *= 1024
		case 'M':
			size *= 1024 * 1024
		}

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

	var parts []string
	order := []string{"L1d", "L1i", "L2", "L3", "L4"}
	for _, label := range order {
		if size, ok := result[label]; ok && size > 0 {
			var formattedSize string
			if size >= 1048576 {
				formattedSize = fmt.Sprintf("%dM", size/1048576)
			} else {
				formattedSize = fmt.Sprintf("%dK", size/1024)
			}
			parts = append(parts, fmt.Sprintf("%s:%s", label, formattedSize))
		}
	}

	return strings.Join(parts, " ")
}

func getKernelInfo() string {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return ""
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

	return fmt.Sprintf("%s %s %s %s %s",
		toString(uname.Sysname),
		toString(uname.Nodename),
		toString(uname.Release),
		toString(uname.Version),
		toString(uname.Machine))
}

func getNTPInfo() (bool, float64, float64) {
	tx := &unix.Timex{}
	state, err := unix.Adjtimex(tx)
	if err != nil {
		return false, 0, 0
	}

	isSynced := state != unix.TIME_ERROR
	offsetSeconds := float64(tx.Offset) / 1_000_000.0
	maxErrorSeconds := float64(tx.Maxerror) / 1_000_000.0

	return isSynced, offsetSeconds, maxErrorSeconds
}

func getLoadAvg() (float64, int64) {
	val, ts := probing.File("/proc/loadavg")
	parts := strings.Fields(val)
	if len(parts) > 0 {
		return probing.ParseFloat64(parts[0]), ts
	}
	return 0.0, ts
}

func getCPUFreq() (float64, int64) {
	// Try sysfs first
	files, err := filepath.Glob("/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq")
	ts := probing.GetTimestamp()
	if err == nil && len(files) > 0 {
		var total int64
		var count int64
		for _, f := range files {
			val, _ := probing.FileInt(f)
			if val > 0 {
				total += val
				count++
			}
		}
		if count > 0 {
			return float64(total) / float64(count) / 1000.0, ts
		}
	}

	// Fallback to /proc/cpuinfo
	lines, ts := probing.FileLines("/proc/cpuinfo")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return probing.ParseFloat64(parts[1]), ts
			}
		}
	}

	return 0.0, ts
}
