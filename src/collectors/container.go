package collectors

import (
	"encoding/json"
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
	blkioPath := filepath.Join(CgroupDir, "blkio")
	pidsPath := filepath.Join(CgroupDir, "pids")

	// CPU metrics
	cpuUsage, tCpu := ProbeFileInt(filepath.Join(cpuPath, "cpuacct.usage"))
	cpuStat, tCpuStat := ProbeFileKV(filepath.Join(cpuPath, "cpuacct.stat"), " ")

	// Per-CPU times (serialized as JSON)
	perCpuJSON, tPerCpu := getPerCPUTimesV1(cpuPath)

	// Memory metrics
	usage, tU := ProbeFileInt(filepath.Join(memPath, "memory.usage_in_bytes"))
	maxMem, tM := ProbeFileInt(filepath.Join(memPath, "memory.max_usage_in_bytes"))
	memStat, tMemStat := ProbeFileKV(filepath.Join(memPath, "memory.stat"), " ")

	// Disk metrics
	dr, dw, tBlk := getBlkioV1()
	sectorIO, tSector := getBlkioSectorsV1(blkioPath)

	// Process count
	numProcs, tProcs := getProcessCountV1(pidsPath)

	m.ContainerCPUTime = cpuUsage
	m.ContainerCPUTimeT = tCpu
	m.ContainerCPUTimeUserMode = parseInt64(cpuStat["user"]) * JiffiesPerSecond
	m.ContainerCPUTimeUserModeT = tCpuStat
	m.ContainerCPUTimeKernelMode = parseInt64(cpuStat["system"]) * JiffiesPerSecond
	m.ContainerCPUTimeKernelModeT = tCpuStat
	m.ContainerPerCPUTimesJSON = perCpuJSON
	m.ContainerPerCPUTimesT = tPerCpu
	m.ContainerMemoryUsed = usage
	m.ContainerMemoryUsedT = tU
	m.ContainerMemoryMaxUsed = maxMem
	m.ContainerMemoryMaxUsedT = tM
	m.ContainerPgFault = parseInt64(memStat["pgfault"])
	m.ContainerPgFaultT = tMemStat
	m.ContainerMajorPgFault = parseInt64(memStat["pgmajfault"])
	m.ContainerMajorPgFaultT = tMemStat
	m.ContainerDiskReadBytes = dr
	m.ContainerDiskReadBytesT = tBlk
	m.ContainerDiskWriteBytes = dw
	m.ContainerDiskWriteBytesT = tBlk
	m.ContainerDiskSectorIO = sectorIO
	m.ContainerDiskSectorIOT = tSector
	m.ContainerNumProcesses = numProcs
	m.ContainerNumProcessesT = tProcs
}

// getPerCPUTimesV1 reads per-CPU times from cpuacct.usage_percpu and returns as JSON
func getPerCPUTimesV1(cpuPath string) (string, int64) {
	content, ts := ProbeFile(filepath.Join(cpuPath, "cpuacct.usage_percpu"))
	if content == "" {
		return "", ts
	}
	fields := strings.Fields(content)
	times := make([]int64, 0, len(fields))
	for _, f := range fields {
		times = append(times, parseInt64(f))
	}
	// Serialize to JSON
	if len(times) > 0 {
		data, _ := json.Marshal(times)
		return string(data), ts
	}
	return "", ts
}

// getBlkioSectorsV1 reads total sector IO from blkio.sectors
func getBlkioSectorsV1(blkioPath string) (int64, int64) {
	var total int64
	lines, ts := ProbeFileLines(filepath.Join(blkioPath, "blkio.sectors"))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) >= 2 {
			total += parseInt64(f[1])
		}
	}
	return total, ts
}

// getProcessCountV1 counts processes in the cgroup
func getProcessCountV1(pidsPath string) (int64, int64) {
	// Try pids.current first (if pids controller available)
	if count, ts := ProbeFileInt(filepath.Join(pidsPath, "pids.current")); count > 0 {
		return count, ts
	}
	// Fallback: count lines in tasks file
	lines, ts := ProbeFileLines(filepath.Join(CgroupDir, "cpuacct", "tasks"))
	return int64(len(lines)), ts
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
	cpuStats, tCpu := ProbeFileKV(filepath.Join(CgroupDir, "cpu.stat"), " ")
	memUsage, tMem := ProbeFileInt(filepath.Join(CgroupDir, "memory.current"))
	memPeak, tPeak := ProbeFileInt(filepath.Join(CgroupDir, "memory.peak"))
	memStat, tMemStat := ProbeFileKV(filepath.Join(CgroupDir, "memory.stat"), " ")
	dr, dw, tIO := getIOStatV2()
	numProcs, tProcs := ProbeFileInt(filepath.Join(CgroupDir, "pids.current"))

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
	m.ContainerPgFault = parseInt64(memStat["pgfault"])
	m.ContainerPgFaultT = tMemStat
	m.ContainerMajorPgFault = parseInt64(memStat["pgmajfault"])
	m.ContainerMajorPgFaultT = tMemStat
	m.ContainerDiskReadBytes = dr
	m.ContainerDiskReadBytesT = tIO
	m.ContainerDiskWriteBytes = dw
	m.ContainerDiskWriteBytesT = tIO
	m.ContainerNumProcesses = numProcs
	m.ContainerNumProcessesT = tProcs
	// Note: cgroup v2 doesn't have per-CPU breakdown or sector IO equivalent
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
