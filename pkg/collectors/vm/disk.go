package vm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

// Collector collects disk metrics.
type Collector struct {
	diskPattern *regexp.Regexp
}

// New creates a new Disk collector.
func NewDiskCollector() *Collector {
	return &Collector{
		diskPattern: regexp.MustCompile(config.DiskRegex),
	}
}

// Name returns the collector name.
func (c *Collector) Name() string {
	return "VM-Disk"
}

// Close releases any resources.
func (c *Collector) Close() error {
	return nil
}

// CollectStatic collects static disk information.
func (c *Collector) CollectStatic() types.Record {
	s := &Static{}

	entries, _ := os.ReadDir(config.SysClassBlock)
	var disks []Info

	for _, entry := range entries {
		devName := entry.Name()
		if !c.diskPattern.MatchString(devName) {
			continue
		}

		basePath := filepath.Join(config.SysClassBlock, devName)
		devicePath := filepath.Join(basePath, "device")

		var model, vendor string

		if strings.HasPrefix(devName, "nvme") {
			model = readFileSilent(filepath.Join(devicePath, "model"))
			vendor = readFileSilent(filepath.Join(devicePath, "subsystem_vendor"))
			if vendor == "" {
				vendor = readFileSilent(filepath.Join(devicePath, "vendor"))
			}
		} else {
			model = readFileSilent(filepath.Join(devicePath, "model"))
			vendor = readFileSilent(filepath.Join(devicePath, "vendor"))
		}

		sizeSectors, _ := probing.FileInt(filepath.Join(basePath, "size"))

		disks = append(disks, Info{
			Name:      devName,
			Model:     strings.TrimSpace(model),
			Vendor:    strings.TrimSpace(vendor),
			SizeBytes: sizeSectors * config.SectorSize,
		})
	}

	if len(disks) > 0 {
		data, _ := json.Marshal(disks)
		s.DisksJSON = string(data)
	}

	return s.ToRecord()
}

func readFileSilent(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// CollectDynamic collects dynamic disk metrics.
func (c *Collector) CollectDynamic() types.Record {
	d := &Dynamic{}

	lines, tDisk := probing.FileLines(config.ProcDiskstats)

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

		readCount += probing.ParseInt64(fields[3])
		mergedReads += probing.ParseInt64(fields[4])
		sectorReads += probing.ParseInt64(fields[5])
		readTimeMs += probing.ParseInt64(fields[6])
		writeCount += probing.ParseInt64(fields[7])
		mergedWrites += probing.ParseInt64(fields[8])
		sectorWrites += probing.ParseInt64(fields[9])
		writeTimeMs += probing.ParseInt64(fields[10])
		ioInProgress += probing.ParseInt64(fields[11])
		ioTimeMs += probing.ParseInt64(fields[12])
		weightedIOTimeMs += probing.ParseInt64(fields[13])
	}

	d.DiskSectorReads, d.DiskSectorReadsT = sectorReads, tDisk
	d.DiskSectorWrites, d.DiskSectorWritesT = sectorWrites, tDisk
	d.DiskReadBytes, d.DiskReadBytesT = sectorReads*config.SectorSize, tDisk
	d.DiskWriteBytes, d.DiskWriteBytesT = sectorWrites*config.SectorSize, tDisk
	d.DiskSuccessfulReads, d.DiskSuccessfulReadsT = readCount, tDisk
	d.DiskSuccessfulWrites, d.DiskSuccessfulWritesT = writeCount, tDisk
	d.DiskMergedReads, d.DiskMergedReadsT = mergedReads, tDisk
	d.DiskMergedWrites, d.DiskMergedWritesT = mergedWrites, tDisk
	d.DiskReadTime, d.DiskReadTimeT = readTimeMs, tDisk
	d.DiskWriteTime, d.DiskWriteTimeT = writeTimeMs, tDisk
	d.DiskIOInProgress, d.DiskIOInProgressT = ioInProgress, tDisk
	d.DiskIOTime, d.DiskIOTimeT = ioTimeMs, tDisk
	d.DiskWeightedIOTime, d.DiskWeightedIOTimeT = weightedIOTimeMs, tDisk

	return d.ToRecord()
}
