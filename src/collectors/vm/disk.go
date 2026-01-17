package vm

import (
	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/collectors/types"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DiskCollector collects disk I/O metrics from /proc/diskstats
type DiskCollector struct {
	collectors.BaseCollector
	diskPattern *regexp.Regexp
}

// NewDiskCollector creates a new disk collector
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{
		diskPattern: regexp.MustCompile(collectors.DiskRegex),
	}
}

func (c *DiskCollector) Name() string {
	return "Disk"
}

func (c *DiskCollector) CollectStatic(m *types.StaticMetrics) {
	entries, _ := os.ReadDir("/sys/class/block/")
	var disks []types.DiskStatic

	for _, entry := range entries {
		devName := entry.Name()
		if !c.diskPattern.MatchString(devName) {
			continue
		}

		basePath := filepath.Join("/sys/class/block", devName)

		model, _ := collectors.ProbeFile(filepath.Join(basePath, "device/model"))
		vendor, _ := collectors.ProbeFile(filepath.Join(basePath, "device/vendor"))
		sizeSectors, _ := collectors.ProbeFileInt(filepath.Join(basePath, "size"))

		disks = append(disks, types.DiskStatic{
			Name:      devName,
			Model:     strings.TrimSpace(model),
			Vendor:    strings.TrimSpace(vendor),
			SizeBytes: sizeSectors * collectors.SectorSize,
		})
	}

	if len(disks) > 0 {
		if data, err := json.Marshal(disks); err == nil {
			m.DisksJSON = string(data)
		}
	}
}

func (c *DiskCollector) CollectDynamic(m *types.DynamicMetrics) {
	lines, ts := collectors.ProbeFileLines("/proc/diskstats")

	var readCount, mergedReads, sectorReads, readTimeMs int64
	var writeCount, mergedWrites, sectorWrites, writeTimeMs int64
	var ioInProgress, ioTimeMs, weightedIOTimeMs int64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		devName := fields[2]
		if !c.diskPattern.MatchString(devName) {
			continue
		}

		readCount += collectors.parseInt64(fields[3])
		mergedReads += collectors.parseInt64(fields[4])
		sectorReads += collectors.parseInt64(fields[5])
		readTimeMs += collectors.parseInt64(fields[6])
		writeCount += collectors.parseInt64(fields[7])
		mergedWrites += collectors.parseInt64(fields[8])
		sectorWrites += collectors.parseInt64(fields[9])
		writeTimeMs += collectors.parseInt64(fields[10])
		ioInProgress += collectors.parseInt64(fields[11])
		ioTimeMs += collectors.parseInt64(fields[12])
		weightedIOTimeMs += collectors.parseInt64(fields[13])
	}

	m.DiskSectorReads = sectorReads
	m.DiskSectorReadsT = ts
	m.DiskSectorWrites = sectorWrites
	m.DiskSectorWritesT = ts
	m.DiskReadBytes = sectorReads * collectors.SectorSize
	m.DiskReadBytesT = ts
	m.DiskWriteBytes = sectorWrites * collectors.SectorSize
	m.DiskWriteBytesT = ts
	m.DiskSuccessfulReads = readCount
	m.DiskSuccessfulReadsT = ts
	m.DiskSuccessfulWrites = writeCount
	m.DiskSuccessfulWritesT = ts
	m.DiskMergedReads = mergedReads
	m.DiskMergedReadsT = ts
	m.DiskMergedWrites = mergedWrites
	m.DiskMergedWritesT = ts
	m.DiskReadTime = readTimeMs
	m.DiskReadTimeT = ts
	m.DiskWriteTime = writeTimeMs
	m.DiskWriteTimeT = ts
	m.DiskIOInProgress = ioInProgress
	m.DiskIOInProgressT = ts
	m.DiskIOTime = ioTimeMs
	m.DiskIOTimeT = ts
	m.DiskWeightedIOTime = weightedIOTimeMs
	m.DiskWeightedIOTimeT = ts
}
