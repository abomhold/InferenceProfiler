package src

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ContainerMetrics contains container-level measurements from cgroups.
type ContainerMetrics struct {
	// Identification
	ID            string `json:"cId"`            // Container ID (Docker/K8s short ID)
	CgroupVersion int    `json:"cCgroupVersion"` // 1 or 2

	// CPU (time in nanoseconds for total, centiseconds for user/kernel)
	CPUTime       exporter.Timed[int64] `json:"cCpuTime"`           // Total CPU nanoseconds
	CPUUserMode   exporter.Timed[int64] `json:"cCpuTimeUserMode"`   // User mode centiseconds
	CPUKernelMode exporter.Timed[int64] `json:"cCpuTimeKernelMode"` // Kernel mode centiseconds
	NumProcessors int                   `json:"cNumProcessors"`     // Available CPUs

	// Per-CPU times (cgroup v1 only)
	PerCPU map[string]exporter.Timed[int64] `json:"perCpu,omitempty"` // cCpu{i}Time in nanoseconds

	// Memory (bytes)
	MemoryUsed    exporter.Timed[int64] `json:"cMemoryUsed"`    // Current usage
	MemoryMaxUsed exporter.Timed[int64] `json:"cMemoryMaxUsed"` // Peak usage

	// Disk I/O (bytes)
	DiskReadBytes  exporter.Timed[int64] `json:"cDiskReadBytes"`
	DiskWriteBytes exporter.Timed[int64] `json:"cDiskWriteBytes"`

	// Network (bytes, from container namespace)
	NetBytesRecvd exporter.Timed[int64] `json:"cNetworkBytesRecvd"`
	NetBytesSent  exporter.Timed[int64] `json:"cNetworkBytesSent"`
}

const cgroupDir = "/sys/fs/cgroup"

// CollectContainer gathers container metrics from cgroups.
// Auto-detects cgroup v1 vs v2.
func CollectContainer() ContainerMetrics {
	if _, err := os.Stat(cgroupDir); err != nil {
		return ContainerMetrics{}
	}

	if isCgroupV2() {
		return collectCgroupV2()
	}
	return collectCgroupV1()
}

// isCgroupV2 checks for cgroup v2 unified hierarchy.
func isCgroupV2() bool {
	_, err := os.Stat(filepath.Join(cgroupDir, "cgroup.controllers"))
	return err == nil
}

// getContainerID attempts to detect container ID from /proc/self/cgroup or hostname.
func getContainerID() string {
	lines, _ := exporter.readLines("/proc/self/cgroup")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		path := parts[2]

		// Docker format: /docker/<container_id>
		if idx := strings.Index(path, "/docker/"); idx >= 0 {
			id := path[idx+8:]
			if len(id) > 12 {
				id = id[:12]
			}
			return id
		}

		// Kubernetes format: /kubepods/.../<container_id>
		if strings.Contains(path, "/kubepods/") {
			segments := strings.Split(path, "/")
			if len(segments) > 0 {
				id := segments[len(segments)-1]
				if len(id) > 12 {
					id = id[:12]
				}
				return id
			}
		}
	}

	// Fallback: hostname (often set to container ID)
	hostname, _ := os.Hostname()
	if len(hostname) == 12 && isHex(hostname) {
		return hostname
	}

	return "unavailable"
}

func isHex(s string) bool {
	for _, c := range strings.ToLower(s) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// collectCgroupV1 gathers metrics from cgroup v1 hierarchy.
func collectCgroupV1() ContainerMetrics {
	cpuUsage, cpuTS := exporter.readInt(filepath.Join(cgroupDir, "cpuacct", "cpuacct.usage"))

	cpuStat, cpuStatTS := exporter.parseKV(filepath.Join(cgroupDir, "cpuacct", "cpuacct.stat"), ' ')
	userJiffies, _ := strconv.ParseInt(cpuStat["user"], 10, 64)
	systemJiffies, _ := strconv.ParseInt(cpuStat["system"], 10, 64)

	memUsage, memTS := exporter.readInt(filepath.Join(cgroupDir, "memory", "memory.usage_in_bytes"))
	memMax, memMaxTS := exporter.readInt(filepath.Join(cgroupDir, "memory", "memory.max_usage_in_bytes"))

	diskRead, diskWrite, blkioTS := parseBlkioV1(filepath.Join(cgroupDir, "blkio", "blkio.throttle.io_service_bytes"))

	netRecv, netSent, netTS := getContainerNetStats()

	perCPU := getPerCPUV1()

	return ContainerMetrics{
		ID:             getContainerID(),
		CgroupVersion:  1,
		CPUTime:        exporter.TimedAt(cpuUsage, cpuTS),
		CPUUserMode:    exporter.TimedAt(userJiffies*jiffiesPerSec, cpuStatTS),
		CPUKernelMode:  exporter.TimedAt(systemJiffies*jiffiesPerSec, cpuStatTS),
		NumProcessors:  getNumCPU(),
		PerCPU:         perCPU,
		MemoryUsed:     exporter.TimedAt(memUsage, memTS),
		MemoryMaxUsed:  exporter.TimedAt(memMax, memMaxTS),
		DiskReadBytes:  exporter.TimedAt(diskRead, blkioTS),
		DiskWriteBytes: exporter.TimedAt(diskWrite, blkioTS),
		NetBytesRecvd:  exporter.TimedAt(netRecv, netTS),
		NetBytesSent:   exporter.TimedAt(netSent, netTS),
	}
}

// collectCgroupV2 gathers metrics from cgroup v2 unified hierarchy.
func collectCgroupV2() ContainerMetrics {
	cpuStat, cpuTS := exporter.parseKV(filepath.Join(cgroupDir, "cpu.stat"), ' ')
	usageUsec, _ := strconv.ParseInt(cpuStat["usage_usec"], 10, 64)
	userUsec, _ := strconv.ParseInt(cpuStat["user_usec"], 10, 64)
	systemUsec, _ := strconv.ParseInt(cpuStat["system_usec"], 10, 64)

	memUsage, memTS := exporter.readInt(filepath.Join(cgroupDir, "memory.current"))
	memPeak, memPeakTS := exporter.readInt(filepath.Join(cgroupDir, "memory.peak"))

	diskRead, diskWrite, ioTS := parseIOStatV2(filepath.Join(cgroupDir, "io.stat"))

	netRecv, netSent, netTS := getContainerNetStats()

	return ContainerMetrics{
		ID:             getContainerID(),
		CgroupVersion:  2,
		CPUTime:        exporter.TimedAt(usageUsec*1000, cpuTS), // usec -> ns
		CPUUserMode:    exporter.TimedAt(userUsec/10000, cpuTS), // usec -> cs
		CPUKernelMode:  exporter.TimedAt(systemUsec/10000, cpuTS),
		NumProcessors:  getNumCPU(),
		MemoryUsed:     exporter.TimedAt(memUsage, memTS),
		MemoryMaxUsed:  exporter.TimedAt(memPeak, memPeakTS),
		DiskReadBytes:  exporter.TimedAt(diskRead, ioTS),
		DiskWriteBytes: exporter.TimedAt(diskWrite, ioTS),
		NetBytesRecvd:  exporter.TimedAt(netRecv, netTS),
		NetBytesSent:   exporter.TimedAt(netSent, netTS),
	}
}

// getPerCPUV1 reads per-CPU usage from cgroup v1.
func getPerCPUV1() map[string]exporter.Timed[int64] {
	result := make(map[string]exporter.Timed[int64])
	content, ts := exporter.readFile(filepath.Join(cgroupDir, "cpuacct", "cpuacct.usage_percpu"))
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
		result[key] = exporter.TimedAt(v, ts)
	}
	return result
}

// parseBlkioV1 parses blkio.throttle.io_service_bytes for v1.
func parseBlkioV1(path string) (read, write, ts int64) {
	lines, ts := exporter.readLines(path)
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
func parseIOStatV2(path string) (read, write, ts int64) {
	lines, ts := exporter.readLines(path)
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
func getContainerNetStats() (recv, sent, ts int64) {
	lines, ts := exporter.readLines("/proc/net/dev")

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
