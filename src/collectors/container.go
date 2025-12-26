package collectors

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const CgroupDir = "/sys/fs/cgroup"

// CollectContainer collects container/cgroup metrics
func CollectContainer() map[string]MetricValue {
	if _, err := os.Stat(CgroupDir); os.IsNotExist(err) {
		return map[string]MetricValue{}
	}

	if isV2() {
		return collectContainerV2()
	}
	return collectContainerV1()
}

func isV2() bool {
	_, err := os.Stat(filepath.Join(CgroupDir, "cgroup.controllers"))
	return err == nil
}

func getContainerID() string {
	lines, _ := ProbeFileLines("/proc/self/cgroup")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			path := parts[2]
			if strings.Contains(path, "/docker/") {
				segments := strings.Split(path, "/docker/")
				if len(segments) > 1 {
					id := segments[len(segments)-1]
					if len(id) >= 12 {
						return id[:12]
					}
				}
			}
		}
	}

	hostname := GetHostname()
	if len(hostname) == 12 {
		isHex := true
		for _, c := range strings.ToLower(hostname) {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				isHex = false
				break
			}
		}
		if isHex {
			return hostname
		}
	}

	return "unavailable"
}

func collectContainerV1() map[string]MetricValue {
	metrics := make(map[string]MetricValue)

	cpuPath := filepath.Join(CgroupDir, "cpuacct", "cpuacct.usage")
	cpuStatPath := filepath.Join(CgroupDir, "cpuacct", "cpuacct.stat")
	memPath := filepath.Join(CgroupDir, "memory", "memory.usage_in_bytes")
	memMaxPath := filepath.Join(CgroupDir, "memory", "memory.max_usage_in_bytes")
	blkioPath := filepath.Join(CgroupDir, "blkio", "blkio.throttle.io_service_bytes")

	cpuUsage, tCpu := ProbeFileInt(cpuPath)
	cpuStat, tCpuStat := ParseProcKV(cpuStatPath, " ")
	cpuUserJiffies := parseInt64(cpuStat["user"])
	cpuKernelJiffies := parseInt64(cpuStat["system"])
	memUsage, tMem := ProbeFileInt(memPath)
	memMax, tMemMax := ProbeFileInt(memMaxPath)
	diskRead, diskWrite, tBlkio := parseBlkioV1(blkioPath)
	perCpu := getPerCpuV1()
	netRecv, netSent, tNet := getContainerNetStats()

	metrics["cId"] = MetricValue{Value: getContainerID(), Time: GetTimestamp()}
	metrics["cCgroupVersion"] = MetricValue{Value: int64(1), Time: GetTimestamp()}
	metrics["cCpuTime"] = NewMetricWithTime(cpuUsage, tCpu)
	metrics["cCpuTimeUserMode"] = NewMetricWithTime(cpuUserJiffies*JiffiesPerSecond, tCpuStat)
	metrics["cCpuTimeKernelMode"] = NewMetricWithTime(cpuKernelJiffies*JiffiesPerSecond, tCpuStat)
	metrics["cNumProcessors"] = MetricValue{Value: int64(runtime.NumCPU()), Time: GetTimestamp()}
	metrics["cMemoryUsed"] = NewMetricWithTime(memUsage, tMem)
	metrics["cMemoryMaxUsed"] = NewMetricWithTime(memMax, tMemMax)
	metrics["cDiskReadBytes"] = NewMetricWithTime(diskRead, tBlkio)
	metrics["cDiskWriteBytes"] = NewMetricWithTime(diskWrite, tBlkio)
	metrics["cNetworkBytesRecvd"] = NewMetricWithTime(netRecv, tNet)
	metrics["cNetworkBytesSent"] = NewMetricWithTime(netSent, tNet)

	for k, v := range perCpu {
		metrics[k] = v
	}

	return metrics
}

func collectContainerV2() map[string]MetricValue {
	metrics := make(map[string]MetricValue)

	cpuStatPath := filepath.Join(CgroupDir, "cpu.stat")
	memPath := filepath.Join(CgroupDir, "memory.current")
	memPeakPath := filepath.Join(CgroupDir, "memory.peak")
	ioStatPath := filepath.Join(CgroupDir, "io.stat")

	memUsage, tMem := ProbeFileInt(memPath)
	memPeak, tMemPeak := ProbeFileInt(memPeakPath)
	cpuStats, tCpu := ParseProcKV(cpuStatPath, " ")
	cpuUsageMicros := parseInt64(cpuStats["usage_usec"])
	cpuUserMicros := parseInt64(cpuStats["user_usec"])
	cpuSystemMicros := parseInt64(cpuStats["system_usec"])
	diskRead, diskWrite, tIO := parseIOStatV2(ioStatPath)
	netRecv, netSent, tNet := getContainerNetStats()

	metrics["cId"] = MetricValue{Value: getContainerID()}
	metrics["cCgroupVersion"] = MetricValue{Value: int64(2)}
	metrics["cCpuTime"] = NewMetricWithTime(cpuUsageMicros*1000, tCpu)
	metrics["cCpuTimeUserMode"] = NewMetricWithTime(cpuUserMicros/10000, tCpu)
	metrics["cCpuTimeKernelMode"] = NewMetricWithTime(cpuSystemMicros/10000, tCpu)
	metrics["cNumProcessors"] = MetricValue{Value: int64(runtime.NumCPU()), Time: GetTimestamp()}
	metrics["cMemoryUsed"] = NewMetricWithTime(memUsage, tMem)
	metrics["cMemoryMaxUsed"] = NewMetricWithTime(memPeak, tMemPeak)
	metrics["cDiskReadBytes"] = NewMetricWithTime(diskRead, tIO)
	metrics["cDiskWriteBytes"] = NewMetricWithTime(diskWrite, tIO)
	metrics["cNetworkBytesRecvd"] = NewMetricWithTime(netRecv, tNet)
	metrics["cNetworkBytesSent"] = NewMetricWithTime(netSent, tNet)

	return metrics
}

func getPerCpuV1() map[string]MetricValue {
	result := make(map[string]MetricValue)
	percpuPath := filepath.Join(CgroupDir, "cpuacct", "cpuacct.usage_percpu")
	content, ts := ProbeFile(percpuPath)

	if content != "" {
		values := strings.Fields(content)
		for i, val := range values {
			v := parseInt64(val)
			result[strings.ReplaceAll("cCpu"+string(rune('0'+i))+"Time", "", "")] = NewMetricWithTime(v, ts)
		}
	}
	return result
}

func parseBlkioV1(path string) (int64, int64, int64) {
	var readBytes, writeBytes int64

	lines, ts := ProbeFileLines(path)
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			op := strings.ToLower(parts[1])
			value := parseInt64(parts[2])
			if op == "read" {
				readBytes += value
			} else if op == "write" {
				writeBytes += value
			}
		}
	}
	return readBytes, writeBytes, ts
}

func parseIOStatV2(path string) (int64, int64, int64) {
	var readBytes, writeBytes int64

	lines, ts := ProbeFileLines(path)
	for _, line := range lines {
		parts := strings.Fields(line)
		for _, part := range parts {
			if strings.HasPrefix(part, "rbytes=") {
				val := strings.TrimPrefix(part, "rbytes=")
				readBytes += parseInt64(val)
			} else if strings.HasPrefix(part, "wbytes=") {
				val := strings.TrimPrefix(part, "wbytes=")
				writeBytes += parseInt64(val)
			}
		}
	}
	return readBytes, writeBytes, ts
}

func getContainerNetStats() (int64, int64, int64) {
	var recv, sent int64

	lines, ts := ProbeFileLines("/proc/net/dev")
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) >= 9 {
			recv += parseInt64(fields[0])
			sent += parseInt64(fields[8])
		}
	}
	return recv, sent, ts
}
