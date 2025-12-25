package collectors

import (
	"regexp"
	"strings"
)

// DiskCollector gathers disk I/O metrics.
type DiskCollector struct {
	BaseCollector
}

const sectorSize = 512

// diskPattern matches physical disk devices, excluding partitions and virtual devices.
// Matches: sda, hda, vda, xvda, nvme0n1, mmcblk0
// Excludes: sda1, loop0, ram0
var diskPattern = regexp.MustCompile(`^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`)

// Collect gathers disk I/O metrics from /proc/diskstats.
func (c *DiskCollector) Collect() DiskMetrics {
	lines, ts := c.ReadLines("/proc/diskstats")

	var (
		readCount, mergedReads, sectorReads, readTime     int64
		writeCount, mergedWrites, sectorWrites, writeTime int64
		ioInProgress, ioTime, weightedIOTime              int64
	)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		devName := fields[2]
		if !diskPattern.MatchString(devName) {
			continue
		}

		// /proc/diskstats field positions (0-indexed):
		// 0: major, 1: minor, 2: device name
		// 3: reads completed, 4: reads merged, 5: sectors read, 6: read time ms
		// 7: writes completed, 8: writes merged, 9: sectors written, 10: write time ms
		// 11: I/O in progress, 12: I/O time ms, 13: weighted I/O time ms
		readCount += c.ParseInt(fields[3])
		mergedReads += c.ParseInt(fields[4])
		sectorReads += c.ParseInt(fields[5])
		readTime += c.ParseInt(fields[6])
		writeCount += c.ParseInt(fields[7])
		mergedWrites += c.ParseInt(fields[8])
		sectorWrites += c.ParseInt(fields[9])
		writeTime += c.ParseInt(fields[10])
		ioInProgress += c.ParseInt(fields[11])
		ioTime += c.ParseInt(fields[12])
		weightedIOTime += c.ParseInt(fields[13])
	}

	return DiskMetrics{
		SectorReads:      TimedAt(sectorReads, ts),
		SectorWrites:     TimedAt(sectorWrites, ts),
		ReadBytes:        TimedAt(sectorReads*sectorSize, ts),
		WriteBytes:       TimedAt(sectorWrites*sectorSize, ts),
		SuccessfulReads:  TimedAt(readCount, ts),
		SuccessfulWrites: TimedAt(writeCount, ts),
		MergedReads:      TimedAt(mergedReads, ts),
		MergedWrites:     TimedAt(mergedWrites, ts),
		ReadTime:         TimedAt(readTime, ts),
		WriteTime:        TimedAt(writeTime, ts),
		IOTime:           TimedAt(ioTime, ts),
		WeightedIOTime:   TimedAt(weightedIOTime, ts),
		IOInProgress:     TimedAt(ioInProgress, ts),
	}
}
