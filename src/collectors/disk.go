package collectors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	SectorSize = 512
	DiskRegex  = `^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`
)

// CollectDiskStatic populates static disk information
func CollectDiskStatic(m *StaticMetrics) {
	diskPattern := regexp.MustCompile(DiskRegex)
	entries, _ := os.ReadDir("/sys/class/block/")
	var disks []DiskStatic

	for _, entry := range entries {
		devName := entry.Name()
		if !diskPattern.MatchString(devName) {
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

// CollectDiskDynamic populates dynamic disk metrics
func CollectDiskDynamic(m *DynamicMetrics) {
	lines, ts := ProbeFileLines("/proc/diskstats")
	diskPattern := regexp.MustCompile(DiskRegex)

	var readCount, mergedReads, sectorReads, readTimeMs int64
	var writeCount, mergedWrites, sectorWrites, writeTimeMs int64
	var ioInProgress, ioTimeMs, weightedIOTimeMs int64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		devName := fields[2]
		if !diskPattern.MatchString(devName) {
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
