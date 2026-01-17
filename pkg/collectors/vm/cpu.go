package vm

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"

	"golang.org/x/sys/unix"
)

// Collector collects CPU metrics.
type Collector struct{}

// New creates a new CPU collector.
func NewCPUCollector() *Collector {
	return &Collector{}
}

// Name returns the collector name.
func (c *Collector) Name() string {
	return "VM-CPU"
}

// Close releases any resources.
func (c *Collector) Close() error {
	return nil
}

// CollectStatic collects static CPU information.
func (c *Collector) CollectStatic() types.Record {
	s := &Static{
		NumProcessors: runtime.NumCPU(),
		CPUType:       getCPUType(),
		CPUCache:      getCPUCache(),
		KernelInfo:    getKernelInfo(),
	}

	synced, offset, maxErr := getNTPInfo()
	s.TimeSynced = synced
	s.TimeOffsetSeconds = offset
	s.TimeMaxErrorSeconds = maxErr

	return s.ToRecord()
}

// CollectDynamic collects dynamic CPU metrics.
func (c *Collector) CollectDynamic() types.Record {
	d := &Dynamic{}

	lines, tStat := probing.FileLines(config.ProcStat)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if fields[0] == "cpu" && len(fields) >= 11 {
			mult := int64(config.NanosecondsPerSec / config.JiffiesPerSecond)
			d.CPUTimeUserMode = probing.ParseInt64(fields[1]) * mult
			d.CPUTimeUserModeT = tStat
			d.CPUNice = probing.ParseInt64(fields[2]) * mult
			d.CPUNiceT = tStat
			d.CPUTimeKernelMode = probing.ParseInt64(fields[3]) * mult
			d.CPUTimeKernelModeT = tStat
			d.CPUIdleTime = probing.ParseInt64(fields[4]) * mult
			d.CPUIdleTimeT = tStat
			d.CPUTimeIOWait = probing.ParseInt64(fields[5]) * mult
			d.CPUTimeIOWaitT = tStat
			d.CPUTimeIntSrvc = probing.ParseInt64(fields[6]) * mult
			d.CPUTimeIntSrvcT = tStat
			d.CPUTimeSoftIntSrvc = probing.ParseInt64(fields[7]) * mult
			d.CPUTimeSoftIntSrvcT = tStat
			d.CPUSteal = probing.ParseInt64(fields[8]) * mult
			d.CPUStealT = tStat

			d.CPUTime = (probing.ParseInt64(fields[1]) + probing.ParseInt64(fields[2]) +
				probing.ParseInt64(fields[3]) + probing.ParseInt64(fields[4]) +
				probing.ParseInt64(fields[5]) + probing.ParseInt64(fields[6]) +
				probing.ParseInt64(fields[7]) + probing.ParseInt64(fields[8])) * mult
			d.CPUTimeT = tStat
		} else if fields[0] == "ctxt" && len(fields) >= 2 {
			d.CPUContextSwitches = probing.ParseInt64(fields[1])
			d.CPUContextSwitchesT = tStat
		}
	}

	d.LoadAvg, d.LoadAvgT = getLoadAvg()
	d.CPUMhz, d.CPUMhzT = getCPUFreq()

	return d.ToRecord()
}

func getCPUType() string {
	lines, _ := probing.FileLines(config.ProcCPUInfo)
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

	pattern := filepath.Join(config.SysCPUPath, "cpu*/cache/index*")
	dirs, _ := filepath.Glob(pattern)
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
	val, ts := probing.File(config.ProcLoadavg)
	parts := strings.Fields(val)
	if len(parts) > 0 {
		return probing.ParseFloat64(parts[0]), ts
	}
	return 0.0, ts
}

func getCPUFreq() (float64, int64) {
	pattern := filepath.Join(config.SysCPUPath, "cpu*/cpufreq/scaling_cur_freq")
	files, err := filepath.Glob(pattern)
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

	lines, ts := probing.FileLines(config.ProcCPUInfo)
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
