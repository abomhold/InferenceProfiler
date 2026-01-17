package disk

import "InferenceProfiler/pkg/collectors/types"

// Info contains static disk information for a single disk.
type Info struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Vendor    string `json:"vendor"`
	SizeBytes int64  `json:"sizeBytes"`
}

// Static contains static disk information.
type Static struct {
	DisksJSON string
}

// ToRecord converts Static to a Record.
func (s *Static) ToRecord() types.Record {
	r := types.Record{}
	if s.DisksJSON != "" {
		r["disks"] = s.DisksJSON
	}
	return r
}

// Dynamic contains dynamic disk metrics.
type Dynamic struct {
	DiskSectorReads       int64
	DiskSectorReadsT      int64
	DiskSectorWrites      int64
	DiskSectorWritesT     int64
	DiskReadBytes         int64
	DiskReadBytesT        int64
	DiskWriteBytes        int64
	DiskWriteBytesT       int64
	DiskSuccessfulReads   int64
	DiskSuccessfulReadsT  int64
	DiskSuccessfulWrites  int64
	DiskSuccessfulWritesT int64
	DiskMergedReads       int64
	DiskMergedReadsT      int64
	DiskMergedWrites      int64
	DiskMergedWritesT     int64
	DiskReadTime          int64
	DiskReadTimeT         int64
	DiskWriteTime         int64
	DiskWriteTimeT        int64
	DiskIOInProgress      int64
	DiskIOInProgressT     int64
	DiskIOTime            int64
	DiskIOTimeT           int64
	DiskWeightedIOTime    int64
	DiskWeightedIOTimeT   int64
}

// ToRecord converts Dynamic to a Record.
func (d *Dynamic) ToRecord() types.Record {
	return types.Record{
		"vDiskSectorReads":       d.DiskSectorReads,
		"vDiskSectorReadsT":      d.DiskSectorReadsT,
		"vDiskSectorWrites":      d.DiskSectorWrites,
		"vDiskSectorWritesT":     d.DiskSectorWritesT,
		"vDiskReadBytes":         d.DiskReadBytes,
		"vDiskReadBytesT":        d.DiskReadBytesT,
		"vDiskWriteBytes":        d.DiskWriteBytes,
		"vDiskWriteBytesT":       d.DiskWriteBytesT,
		"vDiskSuccessfulReads":   d.DiskSuccessfulReads,
		"vDiskSuccessfulReadsT":  d.DiskSuccessfulReadsT,
		"vDiskSuccessfulWrites":  d.DiskSuccessfulWrites,
		"vDiskSuccessfulWritesT": d.DiskSuccessfulWritesT,
		"vDiskMergedReads":       d.DiskMergedReads,
		"vDiskMergedReadsT":      d.DiskMergedReadsT,
		"vDiskMergedWrites":      d.DiskMergedWrites,
		"vDiskMergedWritesT":     d.DiskMergedWritesT,
		"vDiskReadTime":          d.DiskReadTime,
		"vDiskReadTimeT":         d.DiskReadTimeT,
		"vDiskWriteTime":         d.DiskWriteTime,
		"vDiskWriteTimeT":        d.DiskWriteTimeT,
		"vDiskIOInProgress":      d.DiskIOInProgress,
		"vDiskIOInProgressT":     d.DiskIOInProgressT,
		"vDiskIOTime":            d.DiskIOTime,
		"vDiskIOTimeT":           d.DiskIOTimeT,
		"vDiskWeightedIOTime":    d.DiskWeightedIOTime,
		"vDiskWeightedIOTimeT":   d.DiskWeightedIOTimeT,
	}
}
