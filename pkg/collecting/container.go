package collecting

import (
	"InferenceProfiler/pkg/metrics"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"InferenceProfiler/pkg/probing"
)

type Container struct {
	cgroupVersion int
	cgroupPath    string
}

func NewContainer() *Container {
	if !probing.IsDir(CgroupDir) {
		return nil
	}

	c := &Container{}
	if probing.Exists(filepath.Join(CgroupDir, "cgroup.controllers")) {
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

func (c *Container) Name() string { return "Container" }
func (c *Container) Close() error { return nil }

func (c *Container) CollectStatic() any {
	return &metrics.ContainerStatic{
		ContainerID:      getContainerID(),
		ContainerNumCPUs: int64(runtime.NumCPU()),
		CgroupVersion:    int64(c.cgroupVersion),
	}
}

func (c *Container) CollectDynamic() any {
	m := &metrics.ContainerDynamic{}
	netRecv, netSent, tNet := getContainerNetStats()
	m.ContainerNetworkBytesRecvd, m.ContainerNetworkBytesRecvdT = netRecv, tNet
	m.ContainerNetworkBytesSent, m.ContainerNetworkBytesSentT = netSent, tNet

	switch c.cgroupVersion {
	case 1:
		c.collectV1(m)
	case 2:
		c.collectV2(m)
	}

	return m
}

func (c *Container) collectV1(m *metrics.ContainerDynamic) {
	cpuPath := filepath.Join(CgroupDir, "cpuacct")
	memPath := filepath.Join(CgroupDir, "memory")
	blkioPath := filepath.Join(CgroupDir, "blkio")
	pidsPath := filepath.Join(CgroupDir, "pids")
	cpuUsage, tCpu := probing.FileInt(filepath.Join(cpuPath, "cpuacct.usage"))
	cpuStat, tCpuStat := probing.FileKV(filepath.Join(cpuPath, "cpuacct.stat"), " ")
	perCpuJSON, tPerCpu := getPerCPUTimesV1(cpuPath)
	usage, tU := probing.FileInt(filepath.Join(memPath, "memory.usage_in_bytes"))
	maxMem, tM := probing.FileInt(filepath.Join(memPath, "memory.max_usage_in_bytes"))
	memStat, tMemStat := probing.FileKV(filepath.Join(memPath, "memory.stat"), " ")
	dr, dw, tBlk := getBlkioV1()
	sectorIO, tSector := getBlkioSectorsV1(blkioPath)
	numProcs, tProcs := getProcessCountV1(pidsPath)

	m.ContainerCPUTime, m.ContainerCPUTimeT = cpuUsage, tCpu
	m.ContainerCPUTimeUserMode = probing.ParseInt64(cpuStat["user"]) * JiffiesPerSecond
	m.ContainerCPUTimeUserModeT = tCpuStat
	m.ContainerCPUTimeKernelMode = probing.ParseInt64(cpuStat["system"]) * JiffiesPerSecond
	m.ContainerCPUTimeKernelModeT = tCpuStat
	m.ContainerPerCPUTimesJSON, m.ContainerPerCPUTimesT = perCpuJSON, tPerCpu
	m.ContainerMemoryUsed, m.ContainerMemoryUsedT = usage, tU
	m.ContainerMemoryMaxUsed, m.ContainerMemoryMaxUsedT = maxMem, tM
	m.ContainerPgFault = probing.ParseInt64(memStat["pgfault"])
	m.ContainerPgFaultT = tMemStat
	m.ContainerMajorPgFault = probing.ParseInt64(memStat["pgmajfault"])
	m.ContainerMajorPgFaultT = tMemStat
	m.ContainerDiskReadBytes, m.ContainerDiskReadBytesT = dr, tBlk
	m.ContainerDiskWriteBytes, m.ContainerDiskWriteBytesT = dw, tBlk
	m.ContainerDiskSectorIO, m.ContainerDiskSectorIOT = sectorIO, tSector
	m.ContainerNumProcesses, m.ContainerNumProcessesT = numProcs, tProcs
}

func (c *Container) collectV2(m *metrics.ContainerDynamic) {
	cpuStats, tCpu := probing.FileKV(filepath.Join(c.cgroupPath, "cpu.stat"), " ")
	memUsage, tMem := probing.FileInt(filepath.Join(c.cgroupPath, "memory.current"))
	memPeak, tPeak := probing.FileInt(filepath.Join(c.cgroupPath, "memory.peak"))
	memStat, tMemStat := probing.FileKV(filepath.Join(c.cgroupPath, "memory.stat"), " ")
	dr, dw, tIO := getIOStatV2(c.cgroupPath)
	numProcs, tProcs := probing.FileInt(filepath.Join(c.cgroupPath, "pids.current"))

	m.ContainerCPUTime = probing.ParseInt64(cpuStats["usage_usec"]) * 1000
	m.ContainerCPUTimeT = tCpu
	m.ContainerCPUTimeUserMode = probing.ParseInt64(cpuStats["user_usec"]) / 10000
	m.ContainerCPUTimeUserModeT = tCpu
	m.ContainerCPUTimeKernelMode = probing.ParseInt64(cpuStats["system_usec"]) / 10000
	m.ContainerCPUTimeKernelModeT = tCpu
	m.ContainerMemoryUsed, m.ContainerMemoryUsedT = memUsage, tMem
	m.ContainerMemoryMaxUsed, m.ContainerMemoryMaxUsedT = memPeak, tPeak
	m.ContainerPgFault = probing.ParseInt64(memStat["pgfault"])
	m.ContainerPgFaultT = tMemStat
	m.ContainerMajorPgFault = probing.ParseInt64(memStat["pgmajfault"])
	m.ContainerMajorPgFaultT = tMemStat
	m.ContainerDiskReadBytes, m.ContainerDiskReadBytesT = dr, tIO
	m.ContainerDiskWriteBytes, m.ContainerDiskWriteBytesT = dw, tIO
	m.ContainerNumProcesses, m.ContainerNumProcessesT = numProcs, tProcs
}

func getContainerID() string {
	lines, _ := probing.FileLines("/proc/self/cgroup")
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
	lines, ts := probing.FileLines("/proc/net/dev")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		fields := strings.Fields(strings.SplitN(line, ":", 2)[1])
		if len(fields) >= 9 && !strings.Contains(line, "lo:") {
			recv += probing.ParseInt64(fields[0])
			sent += probing.ParseInt64(fields[8])
		}
	}
	return recv, sent, ts
}

func getPerCPUTimesV1(cpuPath string) (string, int64) {
	val, ts := probing.File(filepath.Join(cpuPath, "cpuacct.usage_percpu"))
	fields := strings.Fields(val)
	times := make([]int64, 0, len(fields))
	for _, f := range fields {
		times = append(times, probing.ParseInt64(f))
	}
	if len(times) > 0 {
		data, _ := json.Marshal(times)
		return string(data), ts
	}
	return "", ts
}

func getBlkioSectorsV1(blkioPath string) (int64, int64) {
	var total int64
	lines, ts := probing.FileLines(filepath.Join(blkioPath, "blkio.sectors"))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) >= 2 {
			total += probing.ParseInt64(f[1])
		}
	}
	return total, ts
}

func getProcessCountV1(pidsPath string) (int64, int64) {
	if count, ts := probing.FileInt(filepath.Join(pidsPath, "pids.current")); count > 0 {
		return count, ts
	}
	lines, ts := probing.FileLines(filepath.Join(CgroupDir, "cpuacct", "tasks"))
	return int64(len(lines)), ts
}

func getBlkioV1() (int64, int64, int64) {
	var r, w int64
	path := filepath.Join(CgroupDir, "blkio", "blkio.throttle.io_service_bytes")
	lines, ts := probing.FileLines(path)
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		val := probing.ParseInt64(f[2])
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
	if !probing.Exists(ioStatPath) {
		return 0, 0, probing.GetTimestamp()
	}
	lines, ts := probing.FileLines(ioStatPath)
	for _, line := range lines {
		for _, p := range strings.Fields(line) {
			if v, ok := strings.CutPrefix(p, "rbytes="); ok {
				r += probing.ParseInt64(v)
			}
			if v, ok := strings.CutPrefix(p, "wbytes="); ok {
				w += probing.ParseInt64(v)
			}
		}
	}
	return r, w, ts
}

func findCgroupV2Path() string {
	lines, _ := probing.FileLines("/proc/self/cgroup")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) == 3 && parts[0] == "0" {
			path := "/sys/fs/cgroup" + parts[2]
			if probing.IsDir(path) {
				return path
			}
		}
	}
	return "/sys/fs/cgroup"
}
