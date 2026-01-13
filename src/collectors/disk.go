package collectors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DiskCollector collects disk I/O metrics from /proc/diskstats
type DiskCollector struct {
	BaseCollector
	diskPattern *regexp.Regexp
}

// NewDiskCollector creates a new disk collector
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{
		diskPattern: regexp.MustCompile(DiskRegex),
	}
}

func (c *DiskCollector) Name() string {
	return "Disk"
}

func (c *DiskCollector) CollectStatic(m *StaticMetrics) {
	entries, _ := os.ReadDir("/sys/class/block/")
	var disks []DiskStatic

	for _, entry := range entries {
		devName := entry.Name()
		if !c.diskPattern.MatchString(devName) {
			continue
		}

		basePath := filepath.Join("/sys/class/block", devName)

		model, _ := ProbeFile(filepath.Join(basePath, "device/model"))
		vendor, _ := ProbeFile(filepath.Join(basePath, "device/vendor"))
		sizeSectors, _ := ProbeFileInt(filepath.Join(basePath, "size"))

		disks = append(disks, DiskStatic{
			Name:      devName,
			Model:     strings.TrimSpace(model),
			Vendor:    strings.TrimSpace(vendor),
			SizeBytes: sizeSectors * SectorSize,
		})
	}

	if len(disks) > 0 {
		if data, err := json.Marshal(disks); err == nil {
			m.DisksJSON = string(data)
		}
	}
}

func (c *DiskCollector) CollectDynamic(m *DynamicMetrics) {
	lines, ts := ProbeFileLines("/proc/diskstats")

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

		readCount += parseInt64(fields[3])
		mergedReads += parseInt64(fields[4])
		sectorReads += parseInt64(fields[5])
		readTimeMs += parseInt64(fields[6])
		writeCount += parseInt64(fields[7])
		mergedWrites += parseInt64(fields[8])
		sectorWrites += parseInt64(fields[9])
		writeTimeMs += parseInt64(fields[10])
		ioInProgress += parseInt64(fields[11])
		ioTimeMs += parseInt64(fields[12])
		weightedIOTimeMs += parseInt64(fields[13])
	}

	m.DiskSectorReads = sectorReads
	m.DiskSectorReadsT = ts
	m.DiskSectorWrites = sectorWrites
	m.DiskSectorWritesT = ts
	m.DiskReadBytes = sectorReads * SectorSize
	m.DiskReadBytesT = ts
	m.DiskWriteBytes = sectorWrites * SectorSize
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
