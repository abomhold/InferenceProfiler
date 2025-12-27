package collectors

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	SectorSize = 512
	DiskRegex  = `^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`
)

// --- Static Metrics ---

func CollectDiskStatic() StaticMetrics {
	results := make(StaticMetrics)
	diskPattern := regexp.MustCompile(DiskRegex)
	entries, _ := os.ReadDir("/sys/class/block/")

	idx := 0
	for _, entry := range entries {
		devName := entry.Name()
		if !diskPattern.MatchString(devName) {
			continue
		}

		basePath := filepath.Join("/sys/class/block", devName)

		model, _ := ProbeFile(filepath.Join(basePath, "device/model"))
		vendor, _ := ProbeFile(filepath.Join(basePath, "device/vendor"))
		sizeSectors, _ := ProbeFileInt(filepath.Join(basePath, "size"))

		prefix := "vDisk" + strconv.Itoa(idx)
		results[prefix+"Name"] = devName
		results[prefix+"Model"] = strings.TrimSpace(model)
		results[prefix+"Vendor"] = strings.TrimSpace(vendor)
		results[prefix+"SizeBytes"] = sizeSectors * SectorSize

		idx++
	}

	return results
}

// --- Dynamic Metrics ---

func CollectDiskDynamic() DynamicMetrics {
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

	return DynamicMetrics{
		"vDiskSectorReads":      NewMetricWithTime(sectorReads, ts),
		"vDiskSectorWrites":     NewMetricWithTime(sectorWrites, ts),
		"vDiskReadBytes":        NewMetricWithTime(sectorReads*SectorSize, ts),
		"vDiskWriteBytes":       NewMetricWithTime(sectorWrites*SectorSize, ts),
		"vDiskSuccessfulReads":  NewMetricWithTime(readCount, ts),
		"vDiskSuccessfulWrites": NewMetricWithTime(writeCount, ts),
		"vDiskMergedReads":      NewMetricWithTime(mergedReads, ts),
		"vDiskMergedWrites":     NewMetricWithTime(mergedWrites, ts),
		"vDiskReadTime":         NewMetricWithTime(readTimeMs, ts),
		"vDiskWriteTime":        NewMetricWithTime(writeTimeMs, ts),
		"vDiskIOInProgress":     NewMetricWithTime(ioInProgress, ts),
		"vDiskIOTime":           NewMetricWithTime(ioTimeMs, ts),
		"vDiskWeightedIOTime":   NewMetricWithTime(weightedIOTimeMs, ts),
	}
}
