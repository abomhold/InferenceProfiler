package vm

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"path/filepath"
	"regexp"
	"strings"
)

type DiskInfo struct {
	Name       string `json:"Name"`
	Model      string `json:"Model"`
	Sectors    int64  `json:"Sectors"`
	Rotational bool   `json:"Rotational"`
}

type DiskStatic struct {
	Disks []DiskInfo `json:"Drives"`
}

type DiskDynamic struct {
	SectorReads      base.MetricInt `json:"SectorReads"`
	SectorWrites     base.MetricInt `json:"SectorWrites"`
	SuccessfulReads  base.MetricInt `json:"SuccessfulReads"`
	SuccessfulWrites base.MetricInt `json:"SuccessfulWrites"`
	MergedReads      base.MetricInt `json:"MergedReads"`
	MergedWrites     base.MetricInt `json:"MergedWrites"`
	ReadTime         base.MetricInt `json:"ReadTime"`
	WriteTime        base.MetricInt `json:"WriteTime"`
	IOInProgress     base.MetricInt `json:"IOInProgress"`
	IOTime           base.MetricInt `json:"IOTime"`
	WeightedIOTime   base.MetricInt `json:"WeightedIOTime"`
}

var (
	procDiskstats = "/proc/diskstats"
	sysBlockBase  = "/sys/class/block"
)

var diskPattern = regexp.MustCompile(`^(sd[a-z]+|nvme\d+n\d+|vd[a-z]+|xvd[a-z]+|hd[a-z]+)$`)

func collectDiskStatic(s *DiskStatic) {
	s.Disks = []DiskInfo{}

	entries, err := filepath.Glob(filepath.Join(sysBlockBase, "*"))
	if err != nil {
		utils.Debugf("disk: glob %s/* error: %v", sysBlockBase, err)
	}
	for _, entry := range entries {
		name := filepath.Base(entry)
		if !diskPattern.MatchString(name) {
			continue
		}

		di := DiskInfo{Name: name}
		if model, _, _ := utils.File(filepath.Join(entry, "device/model")); model != "" {
			di.Model = model
		}
		if sectors, _, err := utils.FileInt(filepath.Join(entry, "size")); err == nil {
			di.Sectors = sectors
		}
		if rot, _, _ := utils.File(filepath.Join(entry, "queue/rotational")); rot == "1" {
			di.Rotational = true
		}
		s.Disks = append(s.Disks, di)
	}
	utils.Debugf("disk: found %d block devices", len(s.Disks))
	for _, d := range s.Disks {
		utils.Debugf("disk: %s model=%q sectors=%d rotational=%v", d.Name, d.Model, d.Sectors, d.Rotational)
	}
}

func collectDiskDynamic(d *DiskDynamic) {
	lines, ts, err := utils.FileLines(procDiskstats)
	if err != nil {
		utils.Debugf("disk: failed to read %s: %v", procDiskstats, err)
		return
	}

	var srV, swV, rdV, wrV, mrV, mwV, rtV, wtV, ioV, iotV, wiotV int64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		name := fields[2]
		if !diskPattern.MatchString(name) {
			continue
		}

		rdV += utils.ParseInt64(fields[3])
		mrV += utils.ParseInt64(fields[4])
		srV += utils.ParseInt64(fields[5])
		rtV += utils.ParseInt64(fields[6])
		wrV += utils.ParseInt64(fields[7])
		mwV += utils.ParseInt64(fields[8])
		swV += utils.ParseInt64(fields[9])
		wtV += utils.ParseInt64(fields[10])
		ioV += utils.ParseInt64(fields[11])
		iotV += utils.ParseInt64(fields[12])
		wiotV += utils.ParseInt64(fields[13])
	}

	d.SectorReads = base.MetricInt{V: srV, T: ts}
	d.SectorWrites = base.MetricInt{V: swV, T: ts}
	d.SuccessfulReads = base.MetricInt{V: rdV, T: ts}
	d.SuccessfulWrites = base.MetricInt{V: wrV, T: ts}
	d.MergedReads = base.MetricInt{V: mrV, T: ts}
	d.MergedWrites = base.MetricInt{V: mwV, T: ts}
	d.ReadTime = base.MetricInt{V: rtV, T: ts}
	d.WriteTime = base.MetricInt{V: wtV, T: ts}
	d.IOInProgress = base.MetricInt{V: ioV, T: ts}
	d.IOTime = base.MetricInt{V: iotV, T: ts}
	d.WeightedIOTime = base.MetricInt{V: wiotV, T: ts}
}
