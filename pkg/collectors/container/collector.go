package container

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

// Collector collects container metrics.
type Collector struct {
	cgroupVersion int
	cgroupPath    string
}

// New creates a new Container collector.
// Returns nil if not running in a container or cgroups are unavailable.
func New() *Collector {
	if !probing.IsDir(config.CgroupDir) {
		return nil
	}

	c := &Collector{}
	if probing.Exists(filepath.Join(config.CgroupDir, "cgroup.controllers")) {
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

// Name returns the collector name.
func (c *Collector) Name() string {
	return "Container"
}

// Close releases any resources.
func (c *Collector) Close() error {
	return nil
}

// CollectStatic collects static container information.
func (c *Collector) CollectStatic() types.Record {
	s := &Static{
		ContainerID:      getContainerID(),
		ContainerNumCPUs: int64(runtime.NumCPU()),
		CgroupVersion:    int64(c.cgroupVersion),
	}
	return s.ToRecord()
}

// CollectDynamic collects dynamic container metrics.
func (c *Collector) CollectDynamic() types.Record {
	d := &Dynamic{}

	netRecv, netSent, tNet := getContainerNetStats()
	d.ContainerNetworkBytesRecvd, d.ContainerNetworkBytesRecvdT = netRecv, tNet
	d.ContainerNetworkBytesSent, d.ContainerNetworkBytesSentT = netSent, tNet

	switch c.cgroupVersion {
	case 1:
		c.collectV1(d)
	case 2:
		c.collectV2(d)
	}

	return d.ToRecord()
}

func (c *Collector) collectV1(d *Dynamic) {
	cpuPath := filepath.Join(config.CgroupDir, "cpuacct")
	memPath := filepath.Join(config.CgroupDir, "memory")
	blkioPath := filepath.Join(config.CgroupDir, "blkio")
	pidsPath := filepath.Join(config.CgroupDir, "pids")

	cpuUsage, tCpu := probing.FileInt(filepath.Join(cpuPath, "cpuacct.usage"))
	cpuStat, tCpuStat := probing.FileKV(filepath.Join(cpuPath, "cpuacct.stat"), " ")
	perCpuJSON, tPerCpu := getPerCPUTimesV1(cpuPath)
	usage, tU := probing.FileInt(filepath.Join(memPath, "memory.usage_in_bytes"))
	maxMem, tM := probing.FileInt(filepath.Join(memPath, "memory.max_usage_in_bytes"))
	memStat, tMemStat := probing.FileKV(filepath.Join(memPath, "memory.stat"), " ")
	dr, dw, tBlk := getBlkioV1()
	sectorIO, tSector := getBlkioSectorsV1(blkioPath)
	numProcs, tProcs := getProcessCountV1(pidsPath)

	d.ContainerCPUTime, d.ContainerCPUTimeT = cpuUsage, tCpu
	d.ContainerCPUTimeUserMode = probing.ParseInt64(cpuStat["user"]) * config.JiffiesPerSecond
	d.ContainerCPUTimeUserModeT = tCpuStat
	d.ContainerCPUTimeKernelMode = probing.ParseInt64(cpuStat["system"]) * config.JiffiesPerSecond
	d.ContainerCPUTimeKernelModeT = tCpuStat
	d.ContainerPerCPUTimesJSON, d.ContainerPerCPUTimesT = perCpuJSON, tPerCpu
	d.ContainerMemoryUsed, d.ContainerMemoryUsedT = usage, tU
	d.ContainerMemoryMaxUsed, d.ContainerMemoryMaxUsedT = maxMem, tM
	d.ContainerPgFault = probing.ParseInt64(memStat["pgfault"])
	d.ContainerPgFaultT = tMemStat
	d.ContainerMajorPgFault = probing.ParseInt64(memStat["pgmajfault"])
	d.ContainerMajorPgFaultT = tMemStat
	d.ContainerDiskReadBytes, d.ContainerDiskReadBytesT = dr, tBlk
	d.ContainerDiskWriteBytes, d.ContainerDiskWriteBytesT = dw, tBlk
	d.ContainerDiskSectorIO, d.ContainerDiskSectorIOT = sectorIO, tSector
	d.ContainerNumProcesses, d.ContainerNumProcessesT = numProcs, tProcs
}

func (c *Collector) collectV2(d *Dynamic) {
	cpuStats, tCpu := probing.FileKV(filepath.Join(c.cgroupPath, "cpu.stat"), " ")
	memUsage, tMem := probing.FileInt(filepath.Join(c.cgroupPath, "memory.current"))
	memPeak, tPeak := probing.FileInt(filepath.Join(c.cgroupPath, "memory.peak"))
	memStat, tMemStat := probing.FileKV(filepath.Join(c.cgroupPath, "memory.stat"), " ")
	dr, dw, tIO := getIOStatV2(c.cgroupPath)
	numProcs, tProcs := probing.FileInt(filepath.Join(c.cgroupPath, "pids.current"))

	d.ContainerCPUTime = probing.ParseInt64(cpuStats["usage_usec"]) * 1000
	d.ContainerCPUTimeT = tCpu
	d.ContainerCPUTimeUserMode = probing.ParseInt64(cpuStats["user_usec"]) / 10000
	d.ContainerCPUTimeUserModeT = tCpu
	d.ContainerCPUTimeKernelMode = probing.ParseInt64(cpuStats["system_usec"]) / 10000
	d.ContainerCPUTimeKernelModeT = tCpu
	d.ContainerMemoryUsed, d.ContainerMemoryUsedT = memUsage, tMem
	d.ContainerMemoryMaxUsed, d.ContainerMemoryMaxUsedT = memPeak, tPeak
	d.ContainerPgFault = probing.ParseInt64(memStat["pgfault"])
	d.ContainerPgFaultT = tMemStat
	d.ContainerMajorPgFault = probing.ParseInt64(memStat["pgmajfault"])
	d.ContainerMajorPgFaultT = tMemStat
	d.ContainerDiskReadBytes, d.ContainerDiskReadBytesT = dr, tIO
	d.ContainerDiskWriteBytes, d.ContainerDiskWriteBytesT = dw, tIO
	d.ContainerNumProcesses, d.ContainerNumProcessesT = numProcs, tProcs
}

func getContainerID() string {
	lines, _ := probing.FileLines(config.ProcSelfCgroup)
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
	lines, ts := probing.FileLines(config.ProcNetDev)
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
	lines, ts := probing.FileLines(filepath.Join(config.CgroupDir, "cpuacct", "tasks"))
	return int64(len(lines)), ts
}

func getBlkioV1() (int64, int64, int64) {
	var r, w int64
	path := filepath.Join(config.CgroupDir, "blkio", "blkio.throttle.io_service_bytes")
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
	lines, _ := probing.FileLines(config.ProcSelfCgroup)
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
