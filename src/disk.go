package main

import (
	"regexp"
	"strings"
)

type DiskMetrics struct {
	SectorReads      Timed[int64] `json:"vDiskSectorReads"`
	SectorWrites     Timed[int64] `json:"vDiskSectorWrites"`
	ReadBytes        Timed[int64] `json:"vDiskReadBytes"`
	WriteBytes       Timed[int64] `json:"vDiskWriteBytes"`
	SuccessfulReads  Timed[int64] `json:"vDiskSuccessfulReads"`
	SuccessfulWrites Timed[int64] `json:"vDiskSuccessfulWrites"`
	MergedReads      Timed[int64] `json:"vDiskMergedReads"`
	MergedWrites     Timed[int64] `json:"vDiskMergedWrites"`
	ReadTime         Timed[int64] `json:"vDiskReadTime"`
	WriteTime        Timed[int64] `json:"vDiskWriteTime"`
	IOTime           Timed[int64] `json:"vDiskIOTime"`
	WeightedIOTime   Timed[int64] `json:"vDiskWeightedIOTime"`
	IOInProgress     Timed[int64] `json:"vDiskIOInProgress"`
}

type DiskCollector struct {
	BaseCollector
}

var diskPattern = regexp.MustCompile(`^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`)

func (c *DiskCollector) Collect() DiskMetrics {
	lines, ts := c.ReadLines("/proc/diskstats")
	var s DiskMetrics
	var rc, mr, sr, rt, wc, mw, sw, wt, ioip, iot, wiot int64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}
		if !diskPattern.MatchString(fields[2]) {
			continue
		}
		rc += c.ParseInt(fields[3])
		mr += c.ParseInt(fields[4])
		sr += c.ParseInt(fields[5])
		rt += c.ParseInt(fields[6])
		wc += c.ParseInt(fields[7])
		mw += c.ParseInt(fields[8])
		sw += c.ParseInt(fields[9])
		wt += c.ParseInt(fields[10])
		ioip += c.ParseInt(fields[11])
		iot += c.ParseInt(fields[12])
		wiot += c.ParseInt(fields[13])
	}

	s.SectorReads = TimedAt(sr, ts)
	s.SectorWrites = TimedAt(sw, ts)
	s.ReadBytes = TimedAt(sr*512, ts)
	s.WriteBytes = TimedAt(sw*512, ts)
	s.SuccessfulReads = TimedAt(rc, ts)
	s.SuccessfulWrites = TimedAt(wc, ts)
	s.MergedReads = TimedAt(mr, ts)
	s.MergedWrites = TimedAt(mw, ts)
	s.ReadTime = TimedAt(rt, ts)
	s.WriteTime = TimedAt(wt, ts)
	s.IOTime = TimedAt(iot, ts)
	s.WeightedIOTime = TimedAt(wiot, ts)
	s.IOInProgress = TimedAt(ioip, ts)
	return s
}
