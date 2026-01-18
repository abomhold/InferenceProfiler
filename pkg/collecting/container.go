package collecting

import (
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ContainerCollector struct {
	cgroupVersion int
	cgroupPath    string
}

func NewContainerCollector() *ContainerCollector {
	if !utils.IsDir(cgroupDir) {
		return nil
	}

	c := &ContainerCollector{}
	if utils.Exists(filepath.Join(cgroupDir, "cgroup.controllers")) {
		c.cgroupVersion = 2
		c.cgroupPath = findCgroupV2Path()
		if c.cgroupPath == "" {
			return nil
		}
	} else {
		c.cgroupVersion = 1
	}

	return c
}

func (c *ContainerCollector) Name() string { return "Container" }
func (c *ContainerCollector) Close() error { return nil }

func (c *ContainerCollector) CollectStatic(m *StaticMetrics) {
	m.ContainerID = getContainerID()
	m.ContainerNumCPUs = int64(runtime.NumCPU())
	m.CgroupVersion = int64(c.cgroupVersion)
}

func (c *ContainerCollector) CollectDynamic(m *DynamicMetrics) {
	m.ContainerNetworkBytesRecvd, m.ContainerNetworkBytesSent, m.ContainerNetworkBytesRecvdT = getContainerNetStats()
	m.ContainerNetworkBytesSentT = m.ContainerNetworkBytesRecvdT

	switch c.cgroupVersion {
	case 1:
		c.collectV1(m)
	case 2:
		c.collectV2(m)
	}
}

func (c *ContainerCollector) collectV1(m *DynamicMetrics) {
	cpuPath := filepath.Join(cgroupDir, "cpuacct")
	memPath := filepath.Join(cgroupDir, "memory")
	blkioPath := filepath.Join(cgroupDir, "blkio")
	pidsPath := filepath.Join(cgroupDir, "pids")

	cpuUsage, tCpu, _ := utils.FileInt(filepath.Join(cpuPath, "cpuacct.usage"))
	cpuStat, tCpuStat, _ := utils.FileKV(filepath.Join(cpuPath, "cpuacct.stat"), fieldSeparatorSpace)
	perCpuJSON, tPerCpu := getPerCPUTimesV1(cpuPath)
	usage, tU, _ := utils.FileInt(filepath.Join(memPath, "memory.usage_in_bytes"))
	maxMem, tM, _ := utils.FileInt(filepath.Join(memPath, "memory.max_usage_in_bytes"))
	memStat, tMemStat, _ := utils.FileKV(filepath.Join(memPath, "memory.stat"), fieldSeparatorSpace)
	dr, dw, tBlk := getBlkioV1()
	sectorIO, tSector := getBlkioSectorsV1(blkioPath)
	numProcs, tProcs := getProcessCountV1(pidsPath)

	m.ContainerCPUTime, m.ContainerCPUTimeT = cpuUsage, tCpu
	m.ContainerCPUTimeUserMode = utils.ParseInt64(cpuStat[("user")]) * jiffiesPerSecond
	m.ContainerCPUTimeUserModeT = tCpuStat
	m.ContainerCPUTimeKernelMode = utils.ParseInt64(cpuStat[("system")]) * jiffiesPerSecond
	m.ContainerCPUTimeKernelModeT = tCpuStat
	m.ContainerPerCPUTimesJSON, m.ContainerPerCPUTimesT = perCpuJSON, tPerCpu
	m.ContainerMemoryUsed, m.ContainerMemoryUsedT = usage, tU
	m.ContainerMemoryMaxUsed, m.ContainerMemoryMaxUsedT = maxMem, tM
	m.ContainerPgFault = utils.ParseInt64(memStat[("pgfault")])
	m.ContainerPgFaultT = tMemStat
	m.ContainerMajorPgFault = utils.ParseInt64(memStat[("pgmajfault")])
	m.ContainerMajorPgFaultT = tMemStat
	m.ContainerDiskReadBytes, m.ContainerDiskReadBytesT = dr, tBlk
	m.ContainerDiskWriteBytes, m.ContainerDiskWriteBytesT = dw, tBlk
	m.ContainerDiskSectorIO, m.ContainerDiskSectorIOT = sectorIO, tSector
	m.ContainerNumProcesses, m.ContainerNumProcessesT = numProcs, tProcs
}

func (c *ContainerCollector) collectV2(m *DynamicMetrics) {
	cpuStats, tCpu, _ := utils.FileKV(filepath.Join(c.cgroupPath, "cpu.stat"), fieldSeparatorSpace)
	memUsage, tMem, _ := utils.FileInt(filepath.Join(c.cgroupPath, "memory.current"))
	memPeak, tPeak, _ := utils.FileInt(filepath.Join(c.cgroupPath, "memory.peak"))
	memStat, tMemStat, _ := utils.FileKV(filepath.Join(c.cgroupPath, "memory.stat"), fieldSeparatorSpace)
	dr, dw, tIO := getIOStatV2(c.cgroupPath)
	numProcs, tProcs, _ := utils.FileInt(filepath.Join(c.cgroupPath, "pids.current"))

	m.ContainerCPUTime = utils.ParseInt64(cpuStats[("usage_usec")]) * microsecondsToNanoseconds
	m.ContainerCPUTimeT = tCpu
	m.ContainerCPUTimeUserMode = utils.ParseInt64(cpuStats[("user_usec")]) / microsecondsToJiffiesDiv
	m.ContainerCPUTimeUserModeT = tCpu
	m.ContainerCPUTimeKernelMode = utils.ParseInt64(cpuStats[("system_usec")]) / microsecondsToJiffiesDiv
	m.ContainerCPUTimeKernelModeT = tCpu
	m.ContainerMemoryUsed, m.ContainerMemoryUsedT = memUsage, tMem
	m.ContainerMemoryMaxUsed, m.ContainerMemoryMaxUsedT = memPeak, tPeak
	m.ContainerPgFault = utils.ParseInt64(memStat[("pgfault")])
	m.ContainerPgFaultT = tMemStat
	m.ContainerMajorPgFault = utils.ParseInt64(memStat[("pgmajfault")])
	m.ContainerMajorPgFaultT = tMemStat
	m.ContainerDiskReadBytes, m.ContainerDiskReadBytesT = dr, tIO
	m.ContainerDiskWriteBytes, m.ContainerDiskWriteBytesT = dw, tIO
	m.ContainerNumProcesses, m.ContainerNumProcessesT = numProcs, tProcs
}

func getContainerID() string {
	lines, _, _ := utils.FileLines("/proc/self/cgroup")
	for _, line := range lines {
		if parts := strings.SplitN(line, fieldSeparatorColon, 3); len(parts) >= 3 {
			path := parts[2]
			if segments := strings.Split(path, "/docker/"); len(segments) > 1 {
				return segments[len(segments)-1]
			}
		}
	}

	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return unavailableValue
}

func getContainerNetStats() (int64, int64, int64) {
	var recv, sent int64
	lines, ts, _ := utils.FileLines("/proc/net/dev")
	for _, line := range lines {
		if !strings.Contains(line, fieldSeparatorColon) || strings.Contains(line, ("lo")+fieldSeparatorColon) {
			continue
		}
		fields := strings.Fields(strings.SplitN(line, fieldSeparatorColon, 2)[1])
		if len(fields) >= 16-7 {
			recv += utils.ParseInt64(fields[0])
			sent += utils.ParseInt64(fields[8])
		}
	}
	return recv, sent, ts
}

func getPerCPUTimesV1(cpuPath string) (string, int64) {
	val, ts, _ := utils.File(filepath.Join(cpuPath, "cpuacct.usage_percpu"))
	fields := strings.Fields(val)
	times := make([]int64, 0, len(fields))
	for _, f := range fields {
		times = append(times, utils.ParseInt64(f))
	}
	if len(times) > 0 {
		data, _ := json.Marshal(times)
		return string(data), ts
	}
	return "", ts
}

func getBlkioSectorsV1(blkioPath string) (int64, int64) {
	var total int64
	lines, ts, _ := utils.FileLines(filepath.Join(blkioPath, "blkio.sectors"))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) >= 2 {
			total += utils.ParseInt64(f[1])
		}
	}
	return total, ts
}

func getProcessCountV1(pidsPath string) (int64, int64) {
	if count, ts, err := utils.FileInt(filepath.Join(pidsPath, "pids.current")); err == nil && count > 0 {
		return count, ts
	}
	lines, ts, _ := utils.FileLines(filepath.Join(cgroupDir, "cpuacct", "tasks"))
	return int64(len(lines)), ts
}

func getBlkioV1() (int64, int64, int64) {
	var r, w int64
	path := filepath.Join(cgroupDir, "blkio", "blkio.throttle.io_service_bytes")
	lines, ts, _ := utils.FileLines(path)
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		val := utils.ParseInt64(f[2])
		if strings.EqualFold(f[1], "read") {
			r += val
		} else if strings.EqualFold(f[1], "write") {
			w += val
		}
	}
	return r, w, ts
}

func getIOStatV2(cgroupPath string) (int64, int64, int64) {
	var r, w int64
	ioStatPath := filepath.Join(cgroupPath, "io.stat")
	if !utils.Exists(ioStatPath) {
		return 0, 0, utils.GetTimestamp()
	}
	lines, ts, _ := utils.FileLines(ioStatPath)
	for _, line := range lines {
		for _, p := range strings.Fields(line) {
			if v, ok := strings.CutPrefix(p, "rbytes="); ok {
				r += utils.ParseInt64(v)
			}
			if v, ok := strings.CutPrefix(p, "wbytes="); ok {
				w += utils.ParseInt64(v)
			}
		}
	}
	return r, w, ts
}

func findCgroupV2Path() string {
	lines, _, _ := utils.FileLines("/proc/self/cgroup")
	for _, line := range lines {
		parts := strings.SplitN(line, fieldSeparatorColon, 3)
		if len(parts) == 3 && parts[0] == ("0") {
			path := ("/sys/fs/cgroup") + parts[2]
			if utils.IsDir(path) {
				return path
			}
		}
	}
	return "/sys/fs/cgroup"
}
