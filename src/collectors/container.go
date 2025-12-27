package collectors

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const CgroupDir = "/sys/fs/cgroup"

// --- Static Metrics ---

func CollectContainerStatic() StaticMetrics {
	if !isCgroupDir() {
		return StaticMetrics{}
	}
	return StaticMetrics{
		"cId":            NewMetric(getContainerID()),
		"cNumProcessors": NewMetric(int64(runtime.NumCPU())),
		"cCgroupVersion": NewMetric(getCgroupVersion()),
	}
}

func isCgroupDir() bool {
	if _, err := os.Stat(CgroupDir); os.IsNotExist(err) {
		log.Println("Cgroup directory not found: " + CgroupDir)
		return false
	}
	return true
}

func getCgroupVersion() int64 {
	_, err := os.Stat(filepath.Join(CgroupDir, "cgroup.controllers"))
	if err == nil {
		return 2
	}
	return 1
}

func getContainerID() string {
	lines, _ := ProbeFileLines("/proc/self/cgroup")
	for _, line := range lines {
		if parts := strings.SplitN(line, ":", 3); len(parts) >= 3 {
			path := parts[2]
			if segments := strings.Split(path, "/docker/"); len(segments) > 1 {
				return segments[len(segments)-1]
			}
		}
	}

	hostname, err := os.Hostname()
	if err == nil {
		return hostname
	}

	return "unavailable"
}

// --- Dynamic Metrics ---

func CollectContainerDynamic() DynamicMetrics {
	if !isCgroupDir() {
		return DynamicMetrics{}
	}
	netRecv, netSent, tNet := getContainerNetStats()

	var vData map[string]MetricValue
	switch getCgroupVersion() {
	case 1:
		vData = collectContainerV1()
		break
	case 2:
		vData = collectContainerV2()
		break
	}

	return DynamicMetrics{
		"cNetworkBytesRecvd": NewMetricWithTime(netRecv, tNet),
		"cNetworkBytesSent":  NewMetricWithTime(netSent, tNet),
	}.Merge(vData)
}

func getContainerNetStats() (int64, int64, int64) {
	var recv, sent int64
	lines, ts := ProbeFileLines("/proc/net/dev")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		fields := strings.Fields(strings.SplitN(line, ":", 2)[1])
		if len(fields) >= 9 && !strings.Contains(line, "lo:") {
			recv += parseInt64(fields[0])
			sent += parseInt64(fields[8])
		}
	}
	return recv, sent, ts
}

// --- V1 Specific Metrics ---

func collectContainerV1() map[string]MetricValue {
	cpuPath := filepath.Join(CgroupDir, "cpuacct")
	memPath := filepath.Join(CgroupDir, "memory")

	cpuUsage, tCpu := ProbeFileInt(filepath.Join(cpuPath, "cpuacct.usage"))
	cpuStat, tCpuStat := ParseProcKV(filepath.Join(cpuPath, "cpuacct.stat"), " ")
	usage, tU := ProbeFileInt(filepath.Join(memPath, "memory.usage_in_bytes"))
	maxMem, tM := ProbeFileInt(filepath.Join(memPath, "memory.max_usage_in_bytes"))
	dr, dw, tBlk := getBlkioV1()

	return DynamicMetrics{
		"cCpuTime":           NewMetricWithTime(cpuUsage, tCpu),
		"cCpuTimeUserMode":   NewMetricWithTime(parseInt64(cpuStat["user"])*JiffiesPerSecond, tCpuStat),
		"cCpuTimeKernelMode": NewMetricWithTime(parseInt64(cpuStat["system"])*JiffiesPerSecond, tCpuStat),
		"cMemoryUsed":        NewMetricWithTime(usage, tU),
		"cMemoryMaxUsed":     NewMetricWithTime(maxMem, tM),
		"cDiskReadBytes":     NewMetricWithTime(dr, tBlk),
		"cDiskWriteBytes":    NewMetricWithTime(dw, tBlk),
	}.Merge(getPerCpuV1())
}

func getPerCpuV1() map[string]MetricValue {
	res := make(map[string]MetricValue)
	content, ts := ProbeFile(filepath.Join(CgroupDir, "cpuacct", "cpuacct.usage_percpu"))
	if content != "" {
		for i, val := range strings.Fields(content) {
			key := "cCpu" + string(rune('0'+i)) + "Time"
			res[key] = NewMetricWithTime(parseInt64(val), ts)
		}
	}
	return res
}

func getBlkioV1() (int64, int64, int64) {
	var r, w int64
	path := filepath.Join(CgroupDir, "blkio", "blkio.throttle.io_service_bytes")
	lines, ts := ProbeFileLines(path)
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		val := parseInt64(f[2])
		if strings.EqualFold(f[1], "read") {
			r += val
		} else if strings.EqualFold(f[1], "write") {
			w += val
		}
	}
	return r, w, ts
}

// --- V2 Specific Metrics ---

func collectContainerV2() map[string]MetricValue {
	cpuStats, tCpu := ParseProcKV(filepath.Join(CgroupDir, "cpu.stat"), " ")
	memUsage, tMem := ProbeFileInt(filepath.Join(CgroupDir, "memory.current"))
	memPeak, tPeak := ProbeFileInt(filepath.Join(CgroupDir, "memory.peak"))
	dr, dw, tIO := getIOStatV2()

	return DynamicMetrics{
		"cCpuTime":           NewMetricWithTime(parseInt64(cpuStats["usage_usec"])*1000, tCpu),
		"cCpuTimeUserMode":   NewMetricWithTime(parseInt64(cpuStats["user_usec"])/10000, tCpu),
		"cCpuTimeKernelMode": NewMetricWithTime(parseInt64(cpuStats["system_usec"])/10000, tCpu),
		"cMemoryUsed":        NewMetricWithTime(memUsage, tMem),
		"cMemoryMaxUsed":     NewMetricWithTime(memPeak, tPeak),
		"cDiskReadBytes":     NewMetricWithTime(dr, tIO),
		"cDiskWriteBytes":    NewMetricWithTime(dw, tIO),
	}
}

func getIOStatV2() (int64, int64, int64) {
	var r, w int64
	lines, ts := ProbeFileLines(filepath.Join(CgroupDir, "io.stat"))
	for _, line := range lines {
		for _, p := range strings.Fields(line) {
			if v, ok := strings.CutPrefix(p, "rbytes="); ok {
				r += parseInt64(v)
			}
			if v, ok := strings.CutPrefix(p, "wbytes="); ok {
				w += parseInt64(v)
			}
		}
	}
	return r, w, ts
}
