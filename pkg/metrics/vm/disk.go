package vm

type DiskInfo struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Vendor    string `json:"vendor"`
	SizeBytes int64  `json:"sizeBytes"`
}

type DiskStatic struct {
	DisksJSON string `json:"disks,omitempty"`
}

type DiskDynamic struct {
	DiskSectorReads       int64 `json:"vDiskSectorReads"`
	DiskSectorReadsT      int64 `json:"vDiskSectorReadsT"`
	DiskSectorWrites      int64 `json:"vDiskSectorWrites"`
	DiskSectorWritesT     int64 `json:"vDiskSectorWritesT"`
	DiskReadBytes         int64 `json:"vDiskReadBytes"`
	DiskReadBytesT        int64 `json:"vDiskReadBytesT"`
	DiskWriteBytes        int64 `json:"vDiskWriteBytes"`
	DiskWriteBytesT       int64 `json:"vDiskWriteBytesT"`
	DiskSuccessfulReads   int64 `json:"vDiskSuccessfulReads"`
	DiskSuccessfulReadsT  int64 `json:"vDiskSuccessfulReadsT"`
	DiskSuccessfulWrites  int64 `json:"vDiskSuccessfulWrites"`
	DiskSuccessfulWritesT int64 `json:"vDiskSuccessfulWritesT"`
	DiskMergedReads       int64 `json:"vDiskMergedReads"`
	DiskMergedReadsT      int64 `json:"vDiskMergedReadsT"`
	DiskMergedWrites      int64 `json:"vDiskMergedWrites"`
	DiskMergedWritesT     int64 `json:"vDiskMergedWritesT"`
	DiskReadTime          int64 `json:"vDiskReadTime"`
	DiskReadTimeT         int64 `json:"vDiskReadTimeT"`
	DiskWriteTime         int64 `json:"vDiskWriteTime"`
	DiskWriteTimeT        int64 `json:"vDiskWriteTimeT"`
	DiskIOInProgress      int64 `json:"vDiskIOInProgress"`
	DiskIOInProgressT     int64 `json:"vDiskIOInProgressT"`
	DiskIOTime            int64 `json:"vDiskIOTime"`
	DiskIOTimeT           int64 `json:"vDiskIOTimeT"`
	DiskWeightedIOTime    int64 `json:"vDiskWeightedIOTime"`
	DiskWeightedIOTimeT   int64 `json:"vDiskWeightedIOTimeT"`
}
