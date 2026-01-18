package collecting

import (
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

var diskPattern = regexp.MustCompile(`^(sd[a-z]+|nvme\d+n\d+|vd[a-z]+|xvd[a-z]+|hd[a-z]+)$`)

// ============================================================================
// CPU Collector
// ============================================================================

type CPUCollector struct{}

func NewCPUCollector() *CPUCollector { return &CPUCollector{} }

func (c *CPUCollector) Name() string { return "CPU" }
func (c *CPUCollector) Close() error { return nil }

func (c *CPUCollector) CollectStatic(s *StaticMetrics) {
	s.NumProcessors = runtime.NumCPU()
	s.CPUType = getCPUType()
	s.CPUCache = getCPUCache()
	s.KernelInfo = getKernelInfo()
	s.TimeSynced, s.TimeOffsetSeconds, s.TimeMaxErrorSeconds = getNTPInfo()
}

func (c *CPUCollector) CollectDynamic(d *DynamicMetrics) {
	lines, tStat, _ := utils.FileLines("/proc/stat")
	mult := int64(time.Nanosecond / jiffiesPerSecond)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if fields[0] == ("cpu") && len(fields) >= 9 {
			d.CPUTimeUserMode = utils.ParseInt64(fields[1]) * mult
			d.CPUTimeUserModeT = tStat
			d.CPUNice = utils.ParseInt64(fields[2]) * mult
			d.CPUNiceT = tStat
			d.CPUTimeKernelMode = utils.ParseInt64(fields[3]) * mult
			d.CPUTimeKernelModeT = tStat
			d.CPUIdleTime = utils.ParseInt64(fields[4]) * mult
			d.CPUIdleTimeT = tStat
			d.CPUTimeIOWait = utils.ParseInt64(fields[5]) * mult
			d.CPUTimeIOWaitT = tStat
			d.CPUTimeIntSrvc = utils.ParseInt64(fields[6]) * mult
			d.CPUTimeIntSrvcT = tStat
			d.CPUTimeSoftIntSrvc = utils.ParseInt64(fields[7]) * mult
			d.CPUTimeSoftIntSrvcT = tStat
			d.CPUSteal = utils.ParseInt64(fields[8]) * mult
			d.CPUStealT = tStat
			d.CPUTime = d.CPUTimeUserMode + d.CPUTimeKernelMode
			d.CPUTimeT = tStat
		} else if fields[0] == ("ctxt") && len(fields) >= 2 {
			d.CPUContextSwitches = utils.ParseInt64(fields[1])
			d.CPUContextSwitchesT = tStat
		}
	}

	d.LoadAvg, d.LoadAvgT = getLoadAvg()
	d.CPUMhz, d.CPUMhzT = getCPUFreq()
}

func getCPUType() string {
	lines, _, _ := utils.FileLines("/proc/cpuinfo")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			if parts := strings.SplitN(line, fieldSeparatorColon, 2); len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return unknownValue
}

func getCPUCache() string {
	result := make(map[string]int64)
	seen := make(map[string]bool)

	dirs, _ := filepath.Glob(filepath.Join("/sys/devices/system/cpu", "cpu*/cache/index*"))
	for _, dir := range dirs {
		level, _, _ := utils.File(filepath.Join(dir, "level"))
		cType, _, _ := utils.File(filepath.Join(dir, "type"))
		sizeStr, _, _ := utils.File(filepath.Join(dir, "size"))
		shared, _, _ := utils.File(filepath.Join(dir, "shared_cpu_map"))

		cacheID := fmt.Sprintf("L%s-%s-%s", level, cType, shared)
		if seen[cacheID] || level == "" || sizeStr == "" {
			continue
		}
		seen[cacheID] = true

		var size int64
		var unit rune
		fmt.Sscanf(sizeStr, "%d%c", &size, &unit)
		switch unit {
		case 'K':
			size *= bytesPerKilobyte
		case 'M':
			size *= bytesPerMegaByte
		}

		suffix := ""
		if level == ("1") {
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
	for _, label := range []string{"L1d", "L1i", "L2", "L3", "L4"} {
		if size, ok := result[label]; ok && size > 0 {
			if size >= bytesPerMegaByte {
				parts = append(parts, fmt.Sprintf("%s:%dM", label, size/bytesPerMegaByte))
			} else {
				parts = append(parts, fmt.Sprintf("%s:%dK", label, size/bytesPerKilobyte))
			}
		}
	}
	return strings.Join(parts, fieldSeparatorSpace)
}

func getKernelInfo() string {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		log.Print("Failed to get kernel info: ", err)
		return ""
	}

	var fields []string
	fields = append(fields, utils.Int8SliceToString(uname.Sysname[:]))
	fields = append(fields, utils.Int8SliceToString(uname.Nodename[:]))
	fields = append(fields, utils.Int8SliceToString(uname.Release[:]))
	fields = append(fields, utils.Int8SliceToString(uname.Version[:]))
	fields = append(fields, utils.Int8SliceToString(uname.Machine[:]))
	return strings.Join(fields, " ")
}

func getNTPInfo() (bool, float64, float64) {
	var buf unix.Timex
	state, err := unix.Adjtimex(&buf)
	if err != nil {
		return false, 0, 0
	}

	synced := state == unix.TIME_OK || state == unix.TIME_INS || state == unix.TIME_DEL

	var offset float64
	if buf.Status&unix.STA_NANO != 0 {
		offset = float64(buf.Offset) / float64(time.Nanosecond)
	} else {
		offset = float64(buf.Offset) / float64(time.Microsecond)
	}

	maxErr := float64(buf.Maxerror) / float64(time.Microsecond)

	return synced, offset, maxErr
}

func getLoadAvg() (float64, int64) {
	val, ts, _ := utils.File("/proc/loadavg")
	if parts := strings.Fields(val); len(parts) > 0 {
		return utils.ParseFloat64(parts[0]), ts
	}
	return 0.0, ts
}

func getCPUFreq() (float64, int64) {
	ts := utils.GetTimestamp()
	files, err := filepath.Glob(filepath.Join("/sys/devices/system/cpu", "cpu*/cpufreq/scaling_cur_freq"))
	if err == nil && len(files) > 0 {
		var total, count int64
		for _, f := range files {
			if val, _, err := utils.FileInt(f); err == nil && val > 0 {
				total += val
				count++
			}
		}
		if count > 0 {
			return float64(total) / float64(count) / float64(bytesPerKilobyte), ts
		}
	}

	lines, ts, _ := utils.FileLines("/proc/cpuinfo")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu MHz") {
			if parts := strings.SplitN(line, fieldSeparatorColon, 2); len(parts) == 2 {
				return utils.ParseFloat64(parts[1]), ts
			}
		}
	}
	return 0.0, ts
}

// ============================================================================
// Memory Collector
// ============================================================================

type MemoryCollector struct{}

func NewMemoryCollector() *MemoryCollector { return &MemoryCollector{} }

func (c *MemoryCollector) Name() string { return "Memory" }
func (c *MemoryCollector) Close() error { return nil }

func (c *MemoryCollector) CollectStatic(s *StaticMetrics) {
	lines, _, _ := utils.FileLines("/proc/meminfo")
	for _, line := range lines {
		parts := strings.SplitN(line, fieldSeparatorColon, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := utils.ParseInt64(strings.Fields(parts[1])[0]) * bytesPerKilobyte

		switch key {
		case "MemTotal":
			s.MemoryTotalBytes = val
		case "SwapTotal":
			s.SwapTotalBytes = val
		}
	}
}

func (c *MemoryCollector) CollectDynamic(d *DynamicMetrics) {
	lines, ts, _ := utils.FileLines("/proc/meminfo")
	for _, line := range lines {
		parts := strings.SplitN(line, fieldSeparatorColon, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) == 0 {
			continue
		}
		val := utils.ParseInt64(fields[0]) * bytesPerKilobyte

		switch key {
		case "MemTotal":
			d.MemoryTotal, d.MemoryTotalT = val, ts
		case "MemFree":
			d.MemoryFree, d.MemoryFreeT = val, ts
		case "Buffers":
			d.MemoryBuffers, d.MemoryBuffersT = val, ts
		case "Cached":
			d.MemoryCached, d.MemoryCachedT = val, ts
		case "SwapTotal":
			d.MemorySwapTotal, d.MemorySwapTotalT = val, ts
		case "SwapFree":
			d.MemorySwapFree, d.MemorySwapFreeT = val, ts
		}
	}

	d.MemoryUsed = d.MemoryTotal - d.MemoryFree - d.MemoryBuffers - d.MemoryCached
	d.MemoryUsedT = ts
	d.MemorySwapUsed = d.MemorySwapTotal - d.MemorySwapFree
	d.MemorySwapUsedT = ts
	if d.MemoryTotal > 0 {
		d.MemoryPercent = float64(d.MemoryUsed) / float64(d.MemoryTotal) * 100
		d.MemoryPercentT = ts
	}

	vmLines, vmTs, _ := utils.FileLines("/proc/vmstat")
	for _, line := range vmLines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "pgfault":
			d.MemoryPgFault, d.MemoryPgFaultT = utils.ParseInt64(fields[1]), vmTs
		case "pgmajfault":
			d.MemoryMajorPageFault, d.MemoryMajorPageFaultT = utils.ParseInt64(fields[1]), vmTs
		}
	}
}

// ============================================================================
// Disk Collector
// ============================================================================

type DiskCollector struct{}

func NewDiskCollector() *DiskCollector { return &DiskCollector{} }

func (c *DiskCollector) Name() string { return "Disk" }
func (c *DiskCollector) Close() error { return nil }

func (c *DiskCollector) CollectStatic(s *StaticMetrics) {
	type diskInfo struct {
		Name       string `json:"name"`
		Model      string `json:"model,omitempty"`
		Size       int64  `json:"size"`
		Rotational bool   `json:"rotational"`
	}

	var disks []diskInfo
	entries, _ := filepath.Glob(filepath.Join("/sys/class/block", "*"))
	for _, entry := range entries {
		name := filepath.Base(entry)
		if !diskPattern.MatchString(name) {
			continue
		}

		di := diskInfo{Name: name}
		if model, _, _ := utils.File(filepath.Join(entry, "device/model")); model != "" {
			di.Model = model
		}
		if size, _, err := utils.FileInt(filepath.Join(entry, "size")); err == nil {
			di.Size = size * sectorBytes
		}
		if rot, _, _ := utils.File(filepath.Join(entry, "queue/rotational")); rot == ("1") {
			di.Rotational = true
		}
		disks = append(disks, di)
	}

	if len(disks) > 0 {
		data, _ := json.Marshal(disks)
		s.DisksJSON = string(data)
	}
}

func (c *DiskCollector) CollectDynamic(d *DynamicMetrics) {
	lines, ts, _ := utils.FileLines("/proc/diskstats")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		name := fields[2]
		if !diskPattern.MatchString(name) {
			continue
		}

		d.DiskSuccessfulReads += utils.ParseInt64(fields[3])
		d.DiskMergedReads += utils.ParseInt64(fields[4])
		d.DiskSectorReads += utils.ParseInt64(fields[5])
		d.DiskReadTime += utils.ParseInt64(fields[6])
		d.DiskSuccessfulWrites += utils.ParseInt64(fields[7])
		d.DiskMergedWrites += utils.ParseInt64(fields[8])
		d.DiskSectorWrites += utils.ParseInt64(fields[9])
		d.DiskWriteTime += utils.ParseInt64(fields[10])
		d.DiskIOInProgress += utils.ParseInt64(fields[11])
		d.DiskIOTime += utils.ParseInt64(fields[12])
		d.DiskWeightedIOTime += utils.ParseInt64(fields[13])
	}

	d.DiskReadBytes = d.DiskSectorReads * sectorBytes
	d.DiskWriteBytes = d.DiskSectorWrites * sectorBytes

	d.DiskSuccessfulReadsT = ts
	d.DiskMergedReadsT = ts
	d.DiskSectorReadsT = ts
	d.DiskReadTimeT = ts
	d.DiskReadBytesT = ts
	d.DiskSuccessfulWritesT = ts
	d.DiskMergedWritesT = ts
	d.DiskSectorWritesT = ts
	d.DiskWriteTimeT = ts
	d.DiskWriteBytesT = ts
	d.DiskIOInProgressT = ts
	d.DiskIOTimeT = ts
	d.DiskWeightedIOTimeT = ts
}

// ============================================================================
// Network Collector
// ============================================================================

type NetworkCollector struct{}

func NewNetworkCollector() *NetworkCollector { return &NetworkCollector{} }

func (c *NetworkCollector) Name() string { return "Network" }
func (c *NetworkCollector) Close() error { return nil }

func (c *NetworkCollector) CollectStatic(s *StaticMetrics) {
	type ifaceInfo struct {
		Name  string `json:"name"`
		MAC   string `json:"mac,omitempty"`
		MTU   int64  `json:"mtu,omitempty"`
		Speed int64  `json:"speed,omitempty"`
	}

	var ifaces []ifaceInfo
	entries, _ := filepath.Glob(filepath.Join("/sys/class/net", "*"))
	for _, entry := range entries {
		name := filepath.Base(entry)
		if name == ("lo") {
			continue
		}

		info := ifaceInfo{Name: name}
		if mac, _, _ := utils.File(filepath.Join(entry, "address")); mac != "" && mac != ("00:00:00:00:00:00") {
			info.MAC = mac
		}
		if mtu, _, err := utils.FileInt(filepath.Join(entry, "mtu")); err == nil {
			info.MTU = mtu
		}
		if speed, _, err := utils.FileInt(filepath.Join(entry, "speed")); err == nil && speed > 0 {
			info.Speed = speed
		}
		ifaces = append(ifaces, info)
	}

	if len(ifaces) > 0 {
		data, _ := json.Marshal(ifaces)
		s.NetworkInterfacesJSON = string(data)
	}
}

func (c *NetworkCollector) CollectDynamic(d *DynamicMetrics) {
	lines, ts, _ := utils.FileLines("/proc/net/dev")
	for _, line := range lines {
		if !strings.Contains(line, fieldSeparatorColon) {
			continue
		}
		parts := strings.SplitN(line, fieldSeparatorColon, 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		if name == ("lo") {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}

		d.NetworkBytesRecvd += utils.ParseInt64(fields[0])
		d.NetworkPacketsRecvd += utils.ParseInt64(fields[1])
		d.NetworkErrorsRecvd += utils.ParseInt64(fields[2])
		d.NetworkDropsRecvd += utils.ParseInt64(fields[3])
		d.NetworkBytesSent += utils.ParseInt64(fields[8])
		d.NetworkPacketsSent += utils.ParseInt64(fields[9])
		d.NetworkErrorsSent += utils.ParseInt64(fields[10])
		d.NetworkDropsSent += utils.ParseInt64(fields[11])
	}

	d.NetworkBytesRecvdT = ts
	d.NetworkPacketsRecvdT = ts
	d.NetworkErrorsRecvdT = ts
	d.NetworkDropsRecvdT = ts
	d.NetworkBytesSentT = ts
	d.NetworkPacketsSentT = ts
	d.NetworkErrorsSentT = ts
	d.NetworkDropsSentT = ts
}
