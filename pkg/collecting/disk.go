package collecting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"InferenceProfiler/pkg/metrics/vm"
	"InferenceProfiler/pkg/probing"
)

type Disk struct {
	diskPattern *regexp.Regexp
}

func NewDisk() *Disk {
	return &Disk{
		diskPattern: regexp.MustCompile(DiskRegex),
	}
}

func (c *Disk) Name() string { return "VM-Disk" }
func (c *Disk) Close() error { return nil }

func (c *Disk) CollectStatic() any {
	m := &vm.DiskStatic{}

	entries, _ := os.ReadDir("/sys/class/block/")
	var disks []vm.DiskInfo

	for _, entry := range entries {
		devName := entry.Name()
		if !c.diskPattern.MatchString(devName) {
			continue
		}

		basePath := filepath.Join("/sys/class/block", devName)
		devicePath := filepath.Join(basePath, "device")

		var model, vendor string

		// NVMe drives have different sysfs structure
		if strings.HasPrefix(devName, "nvme") {
			model = readFileSilent(filepath.Join(devicePath, "model"))
			// NVMe vendor info is in different locations
			vendor = readFileSilent(filepath.Join(devicePath, "subsystem_vendor"))
			if vendor == "" {
				// Try reading from nvme identify
				vendor = readFileSilent(filepath.Join(devicePath, "vendor"))
			}
		} else {
			// SATA/SAS/other drives
			model = readFileSilent(filepath.Join(devicePath, "model"))
			vendor = readFileSilent(filepath.Join(devicePath, "vendor"))
		}

		sizeSectors, _ := probing.FileInt(filepath.Join(basePath, "size"))

		disks = append(disks, vm.DiskInfo{
			Name:      devName,
			Model:     strings.TrimSpace(model),
			Vendor:    strings.TrimSpace(vendor),
			SizeBytes: sizeSectors * SectorSize,
		})
	}

	if len(disks) > 0 {
		data, _ := json.Marshal(disks)
		m.DisksJSON = string(data)
	}

	return m
}

// readFileSilent reads a file and returns empty string on error (no logging).
func readFileSilent(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (c *Disk) CollectDynamic() any {
	m := &vm.DiskDynamic{}

	lines, tDisk := probing.FileLines("/proc/diskstats")

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

	m.DiskSectorReads, m.DiskSectorReadsT = sectorReads, tDisk
	m.DiskSectorWrites, m.DiskSectorWritesT = sectorWrites, tDisk
	m.DiskReadBytes, m.DiskReadBytesT = sectorReads*SectorSize, tDisk
	m.DiskWriteBytes, m.DiskWriteBytesT = sectorWrites*SectorSize, tDisk
	m.DiskSuccessfulReads, m.DiskSuccessfulReadsT = readCount, tDisk
	m.DiskSuccessfulWrites, m.DiskSuccessfulWritesT = writeCount, tDisk
	m.DiskMergedReads, m.DiskMergedReadsT = mergedReads, tDisk
	m.DiskMergedWrites, m.DiskMergedWritesT = mergedWrites, tDisk
	m.DiskReadTime, m.DiskReadTimeT = readTimeMs, tDisk
	m.DiskWriteTime, m.DiskWriteTimeT = writeTimeMs, tDisk
	m.DiskIOInProgress, m.DiskIOInProgressT = ioInProgress, tDisk
	m.DiskIOTime, m.DiskIOTimeT = ioTimeMs, tDisk
	m.DiskWeightedIOTime, m.DiskWeightedIOTimeT = weightedIOTimeMs, tDisk

	return m
}
