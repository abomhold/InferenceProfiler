package collectors

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ContainerCollector struct {
	BaseCollector
}

const cgroupDir = "/sys/fs/cgroup"

func (c *ContainerCollector) Collect() ContainerMetrics {
	if _, err := os.Stat(cgroupDir); err != nil {
		return ContainerMetrics{}
	}

	if c.isCgroupV2() {
		return c.collectCgroupV2()
	}
	return c.collectCgroupV1()
}

func (c *ContainerCollector) isCgroupV2() bool {
	_, err := os.Stat(filepath.Join(cgroupDir, "cgroup.controllers"))
	return err == nil
}

func (c *ContainerCollector) getContainerID() string {
	lines, _ := c.ReadLines("/proc/self/cgroup")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		path := parts[2]

		if idx := strings.Index(path, "/docker/"); idx >= 0 {
			id := path[idx+8:]
			if len(id) > 12 {
				id = id[:12]
			}
			return id
		}

	}
	hostname, _ := os.Hostname()
	if len(hostname) == 12 && c.isHex(hostname) {
		return hostname
	}

	return "unavailable"
}

func (c *ContainerCollector) isHex(s string) bool {
	for _, ch := range strings.ToLower(s) {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}

func (c *ContainerCollector) getNumCPU() int {
	matches, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if len(matches) > 0 {
		return len(matches)
	}
	return 1
}

// collectCgroupV1 gathers metrics from cgroup v1 hierarchy.
func (c *ContainerCollector) collectCgroupV1() ContainerMetrics {
	cpuUsage, cpuTS := c.ReadInt(filepath.Join(cgroupDir, "cpuacct", "cpuacct.usage"))

	cpuStat, cpuStatTS := c.ParseKV(filepath.Join(cgroupDir, "cpuacct", "cpuacct.stat"), ' ')
	userJiffies, _ := strconv.ParseInt(cpuStat["user"], 10, 64)
	systemJiffies, _ := strconv.ParseInt(cpuStat["system"], 10, 64)

	memUsage, memTS := c.ReadInt(filepath.Join(cgroupDir, "memory", "memory.usage_in_bytes"))
	memMax, memMaxTS := c.ReadInt(filepath.Join(cgroupDir, "memory", "memory.max_usage_in_bytes"))

	diskRead, diskWrite, blkioTS := c.parseBlkioV1(filepath.Join(cgroupDir, "blkio", "blkio.throttle.io_service_bytes"))

	netRecv, netSent, netTS := c.getContainerNetStats()

	perCPU := c.getPerCPUV1()

	return ContainerMetrics{
		ID:             c.getContainerID(),
		CgroupVersion:  1,
		CPUTime:        TimedAt(cpuUsage, cpuTS),
		CPUUserMode:    TimedAt(userJiffies*jiffiesPerSec, cpuStatTS),
		CPUKernelMode:  TimedAt(systemJiffies*jiffiesPerSec, cpuStatTS),
		NumProcessors:  c.getNumCPU(),
		PerCPU:         perCPU,
		MemoryUsed:     TimedAt(memUsage, memTS),
		MemoryMaxUsed:  TimedAt(memMax, memMaxTS),
		DiskReadBytes:  TimedAt(diskRead, blkioTS),
		DiskWriteBytes: TimedAt(diskWrite, blkioTS),
		NetBytesRecvd:  TimedAt(netRecv, netTS),
		NetBytesSent:   TimedAt(netSent, netTS),
	}
}

// collectCgroupV2 gathers metrics from cgroup v2 unified hierarchy.
func (c *ContainerCollector) collectCgroupV2() ContainerMetrics {
	cpuStat, cpuTS := c.ParseKV(filepath.Join(cgroupDir, "cpu.stat"), ' ')
	usageUsec, _ := strconv.ParseInt(cpuStat["usage_usec"], 10, 64)
	userUsec, _ := strconv.ParseInt(cpuStat["user_usec"], 10, 64)
	systemUsec, _ := strconv.ParseInt(cpuStat["system_usec"], 10, 64)

	memUsage, memTS := c.ReadInt(filepath.Join(cgroupDir, "memory.current"))
	memPeak, memPeakTS := c.ReadInt(filepath.Join(cgroupDir, "memory.peak"))

	diskRead, diskWrite, ioTS := c.parseIOStatV2(filepath.Join(cgroupDir, "io.stat"))

	netRecv, netSent, netTS := c.getContainerNetStats()

	return ContainerMetrics{
		ID:             c.getContainerID(),
		CgroupVersion:  2,
		CPUTime:        TimedAt(usageUsec*1000, cpuTS), // usec -> ns
		CPUUserMode:    TimedAt(userUsec/10000, cpuTS), // usec -> cs
		CPUKernelMode:  TimedAt(systemUsec/10000, cpuTS),
		NumProcessors:  c.getNumCPU(),
		MemoryUsed:     TimedAt(memUsage, memTS),
		MemoryMaxUsed:  TimedAt(memPeak, memPeakTS),
		DiskReadBytes:  TimedAt(diskRead, ioTS),
		DiskWriteBytes: TimedAt(diskWrite, ioTS),
		NetBytesRecvd:  TimedAt(netRecv, netTS),
		NetBytesSent:   TimedAt(netSent, netTS),
	}
}

// getPerCPUV1 reads per-CPU usage from cgroup v1.
func (c *ContainerCollector) getPerCPUV1() map[string]Timed[int64] {
	result := make(map[string]Timed[int64])
	content, ts := c.ReadFile(filepath.Join(cgroupDir, "cpuacct", "cpuacct.usage_percpu"))
	if content == "" {
		return result
	}

	fields := strings.Fields(content)
	for i, val := range fields {
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			continue
		}
		key := "cCpu" + strconv.Itoa(i) + "Time"
		result[key] = TimedAt(v, ts)
	}
	return result
}

// parseBlkioV1 parses blkio.throttle.io_service_bytes for v1.
func (c *ContainerCollector) parseBlkioV1(path string) (read, write, ts int64) {
	lines, ts := c.ReadLines(path)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		op := strings.ToLower(fields[1])
		val, _ := strconv.ParseInt(fields[2], 10, 64)
		switch op {
		case "read":
			read += val
		case "write":
			write += val
		}
	}
	return read, write, ts
}

// parseIOStatV2 parses io.stat for cgroup v2.
func (c *ContainerCollector) parseIOStatV2(path string) (read, write, ts int64) {
	lines, ts := c.ReadLines(path)
	for _, line := range lines {
		// Format: "major:minor rbytes=X wbytes=Y rios=Z wios=W ..."
		fields := strings.Fields(line)
		for _, f := range fields {
			if strings.HasPrefix(f, "rbytes=") {
				val, _ := strconv.ParseInt(f[7:], 10, 64)
				read += val
			} else if strings.HasPrefix(f, "wbytes=") {
				val, _ := strconv.ParseInt(f[7:], 10, 64)
				write += val
			}
		}
	}
	return read, write, ts
}

// getContainerNetStats reads /proc/net/dev from container's namespace.
func (c *ContainerCollector) getContainerNetStats() (recv, sent, ts int64) {
	lines, ts := c.ReadLines("/proc/net/dev")

	// Skip headers
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}

		iface := strings.TrimSpace(line[:colonIdx])
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(line[colonIdx+1:])
		if len(fields) < 9 {
			continue
		}

		r, _ := strconv.ParseInt(fields[0], 10, 64)
		s, _ := strconv.ParseInt(fields[8], 10, 64)
		recv += r
		sent += s
	}
	return recv, sent, ts
}
