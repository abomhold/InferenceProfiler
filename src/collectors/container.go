package collectors

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const CgroupDir = "/sys/fs/cgroup"

// CollectContainerStatic populates static container information
func CollectContainerStatic(m *StaticMetrics) {
	if !isCgroupDir() {
		return
	}
	m.ContainerID = getContainerID()
	m.ContainerNumCPUs = int64(runtime.NumCPU())
	m.CgroupVersion = getCgroupVersion()
}

// CollectContainerDynamic populates dynamic container metrics
func CollectContainerDynamic(m *DynamicMetrics) {
	if !isCgroupDir() {
		return
	}

	netRecv, netSent, tNet := getContainerNetStats()
	m.ContainerNetworkBytesRecvd = netRecv
	m.ContainerNetworkBytesRecvdT = tNet
	m.ContainerNetworkBytesSent = netSent
	m.ContainerNetworkBytesSentT = tNet

	switch getCgroupVersion() {
	case 1:
		collectContainerV1(m)
	case 2:
		collectContainerV2(m)
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

// collectContainerV1 collects cgroup v1 specific metrics
func collectContainerV1(m *DynamicMetrics) {
	cpuPath := filepath.Join(CgroupDir, "cpuacct")
	memPath := filepath.Join(CgroupDir, "memory")

	cpuUsage, tCpu := ProbeFileInt(filepath.Join(cpuPath, "cpuacct.usage"))
	cpuStat, tCpuStat := ParseProcKV(filepath.Join(cpuPath, "cpuacct.stat"), " ")
	usage, tU := ProbeFileInt(filepath.Join(memPath, "memory.usage_in_bytes"))
	maxMem, tM := ProbeFileInt(filepath.Join(memPath, "memory.max_usage_in_bytes"))
	dr, dw, tBlk := getBlkioV1()

	m.ContainerCPUTime = cpuUsage
	m.ContainerCPUTimeT = tCpu
	m.ContainerCPUTimeUserMode = parseInt64(cpuStat["user"]) * JiffiesPerSecond
	m.ContainerCPUTimeUserModeT = tCpuStat
	m.ContainerCPUTimeKernelMode = parseInt64(cpuStat["system"]) * JiffiesPerSecond
	m.ContainerCPUTimeKernelModeT = tCpuStat
	m.ContainerMemoryUsed = usage
	m.ContainerMemoryUsedT = tU
	m.ContainerMemoryMaxUsed = maxMem
	m.ContainerMemoryMaxUsedT = tM
	m.ContainerDiskReadBytes = dr
	m.ContainerDiskReadBytesT = tBlk
	m.ContainerDiskWriteBytes = dw
	m.ContainerDiskWriteBytesT = tBlk
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

// collectContainerV2 collects cgroup v2 specific metrics
func collectContainerV2(m *DynamicMetrics) {
	cpuStats, tCpu := ParseProcKV(filepath.Join(CgroupDir, "cpu.stat"), " ")
	memUsage, tMem := ProbeFileInt(filepath.Join(CgroupDir, "memory.current"))
	memPeak, tPeak := ProbeFileInt(filepath.Join(CgroupDir, "memory.peak"))
	dr, dw, tIO := getIOStatV2()

	m.ContainerCPUTime = parseInt64(cpuStats["usage_usec"]) * 1000
	m.ContainerCPUTimeT = tCpu
	m.ContainerCPUTimeUserMode = parseInt64(cpuStats["user_usec"]) / 10000
	m.ContainerCPUTimeUserModeT = tCpu
	m.ContainerCPUTimeKernelMode = parseInt64(cpuStats["system_usec"]) / 10000
	m.ContainerCPUTimeKernelModeT = tCpu
	m.ContainerMemoryUsed = memUsage
	m.ContainerMemoryUsedT = tMem
	m.ContainerMemoryMaxUsed = memPeak
	m.ContainerMemoryMaxUsedT = tPeak
	m.ContainerDiskReadBytes = dr
	m.ContainerDiskReadBytesT = tIO
	m.ContainerDiskWriteBytes = dw
	m.ContainerDiskWriteBytesT = tIO
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
