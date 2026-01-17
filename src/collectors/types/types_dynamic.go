package types

// DynamicMetrics contains all dynamic metrics collected during profiling.
// Fields ending in 'T' are timestamps (nanoseconds since Unix epoch) indicating
// when that specific metric was collected.
type DynamicMetrics struct {
	// Timestamp is the overall collection cycle timestamp (nanoseconds since Unix epoch)
	Timestamp int64 `json:"timestamp"`

	// ==========================================================================
	// VM-Level CPU Metrics (centiseconds unless noted)
	// Source: /proc/stat
	// ==========================================================================

	CPUTime             int64 `json:"vCpuTime"` // Total CPU time (user + kernel)
	CPUTimeT            int64 `json:"vCpuTimeT"`
	CPUTimeUserMode     int64 `json:"vCpuTimeUserMode"` // User-space execution time
	CPUTimeUserModeT    int64 `json:"vCpuTimeUserModeT"`
	CPUTimeKernelMode   int64 `json:"vCpuTimeKernelMode"` // Kernel-mode execution time
	CPUTimeKernelModeT  int64 `json:"vCpuTimeKernelModeT"`
	CPUIdleTime         int64 `json:"vCpuIdleTime"` // Idle time
	CPUIdleTimeT        int64 `json:"vCpuIdleTimeT"`
	CPUTimeIOWait       int64 `json:"vCpuTimeIOWait"` // I/O wait time
	CPUTimeIOWaitT      int64 `json:"vCpuTimeIOWaitT"`
	CPUTimeIntSrvc      int64 `json:"vCpuTimeIntSrvc"` // Hardware interrupt time
	CPUTimeIntSrvcT     int64 `json:"vCpuTimeIntSrvcT"`
	CPUTimeSoftIntSrvc  int64 `json:"vCpuTimeSoftIntSrvc"` // Software interrupt time
	CPUTimeSoftIntSrvcT int64 `json:"vCpuTimeSoftIntSrvcT"`
	CPUNice             int64 `json:"vCpuNice"` // Nice (low priority) time
	CPUNiceT            int64 `json:"vCpuNiceT"`
	CPUSteal            int64 `json:"vCpuSteal"` // Hypervisor stolen time
	CPUStealT           int64 `json:"vCpuStealT"`
	CPUContextSwitches  int64 `json:"vCpuContextSwitches"` // Total context switches since boot
	CPUContextSwitchesT int64 `json:"vCpuContextSwitchesT"`

	// LoadAvg is the 1-minute load average (ratio)
	LoadAvg  float64 `json:"vLoadAvg"`
	LoadAvgT int64   `json:"vLoadAvgT"`

	// CPUMhz is the current average CPU frequency (MHz)
	CPUMhz  float64 `json:"vCpuMhz"`
	CPUMhzT int64   `json:"vCpuMhzT"`

	// ==========================================================================
	// VM-Level Memory Metrics (bytes unless noted)
	// Source: /proc/meminfo, /proc/vmstat
	// ==========================================================================

	MemoryTotal           int64   `json:"vMemoryTotal"` // Total physical RAM
	MemoryTotalT          int64   `json:"vMemoryTotalT"`
	MemoryFree            int64   `json:"vMemoryFree"` // Available memory
	MemoryFreeT           int64   `json:"vMemoryFreeT"`
	MemoryUsed            int64   `json:"vMemoryUsed"` // Actively used memory
	MemoryUsedT           int64   `json:"vMemoryUsedT"`
	MemoryBuffers         int64   `json:"vMemoryBuffers"` // Kernel buffer memory
	MemoryBuffersT        int64   `json:"vMemoryBuffersT"`
	MemoryCached          int64   `json:"vMemoryCached"` // Page cache memory
	MemoryCachedT         int64   `json:"vMemoryCachedT"`
	MemoryPercent         float64 `json:"vMemoryPercent"` // RAM usage percentage
	MemoryPercentT        int64   `json:"vMemoryPercentT"`
	MemorySwapTotal       int64   `json:"vMemorySwapTotal"` // Total swap space
	MemorySwapTotalT      int64   `json:"vMemorySwapTotalT"`
	MemorySwapFree        int64   `json:"vMemorySwapFree"` // Free swap space
	MemorySwapFreeT       int64   `json:"vMemorySwapFreeT"`
	MemorySwapUsed        int64   `json:"vMemorySwapUsed"` // Used swap space
	MemorySwapUsedT       int64   `json:"vMemorySwapUsedT"`
	MemoryPgFault         int64   `json:"vMemoryPgFault"` // Minor page faults (count)
	MemoryPgFaultT        int64   `json:"vMemoryPgFaultT"`
	MemoryMajorPageFault  int64   `json:"vMemoryMajorPageFault"` // Major page faults (count)
	MemoryMajorPageFaultT int64   `json:"vMemoryMajorPageFaultT"`

	// ==========================================================================
	// VM-Level Disk I/O Metrics
	// Source: /proc/diskstats
	// ==========================================================================

	DiskSectorReads       int64 `json:"vDiskSectorReads"` // Sectors read (1 sector = 512B)
	DiskSectorReadsT      int64 `json:"vDiskSectorReadsT"`
	DiskSectorWrites      int64 `json:"vDiskSectorWrites"` // Sectors written
	DiskSectorWritesT     int64 `json:"vDiskSectorWritesT"`
	DiskReadBytes         int64 `json:"vDiskReadBytes"` // Total bytes read
	DiskReadBytesT        int64 `json:"vDiskReadBytesT"`
	DiskWriteBytes        int64 `json:"vDiskWriteBytes"` // Total bytes written
	DiskWriteBytesT       int64 `json:"vDiskWriteBytesT"`
	DiskSuccessfulReads   int64 `json:"vDiskSuccessfulReads"` // Completed read operations
	DiskSuccessfulReadsT  int64 `json:"vDiskSuccessfulReadsT"`
	DiskSuccessfulWrites  int64 `json:"vDiskSuccessfulWrites"` // Completed write operations
	DiskSuccessfulWritesT int64 `json:"vDiskSuccessfulWritesT"`
	DiskMergedReads       int64 `json:"vDiskMergedReads"` // Merged read requests
	DiskMergedReadsT      int64 `json:"vDiskMergedReadsT"`
	DiskMergedWrites      int64 `json:"vDiskMergedWrites"` // Merged write requests
	DiskMergedWritesT     int64 `json:"vDiskMergedWritesT"`
	DiskReadTime          int64 `json:"vDiskReadTime"` // Time spent reading (ms)
	DiskReadTimeT         int64 `json:"vDiskReadTimeT"`
	DiskWriteTime         int64 `json:"vDiskWriteTime"` // Time spent writing (ms)
	DiskWriteTimeT        int64 `json:"vDiskWriteTimeT"`
	DiskIOInProgress      int64 `json:"vDiskIOInProgress"` // I/O operations in flight
	DiskIOInProgressT     int64 `json:"vDiskIOInProgressT"`
	DiskIOTime            int64 `json:"vDiskIOTime"` // Total I/O time (ms)
	DiskIOTimeT           int64 `json:"vDiskIOTimeT"`
	DiskWeightedIOTime    int64 `json:"vDiskWeightedIOTime"` // Weighted I/O time (ms)
	DiskWeightedIOTimeT   int64 `json:"vDiskWeightedIOTimeT"`

	// ==========================================================================
	// VM-Level Network Metrics
	// Source: /proc/net/dev (excludes loopback)
	// ==========================================================================

	NetworkBytesRecvd    int64 `json:"vNetworkBytesRecvd"` // Bytes received
	NetworkBytesRecvdT   int64 `json:"vNetworkBytesRecvdT"`
	NetworkBytesSent     int64 `json:"vNetworkBytesSent"` // Bytes transmitted
	NetworkBytesSentT    int64 `json:"vNetworkBytesSentT"`
	NetworkPacketsRecvd  int64 `json:"vNetworkPacketsRecvd"` // Packets received
	NetworkPacketsRecvdT int64 `json:"vNetworkPacketsRecvdT"`
	NetworkPacketsSent   int64 `json:"vNetworkPacketsSent"` // Packets transmitted
	NetworkPacketsSentT  int64 `json:"vNetworkPacketsSentT"`
	NetworkErrorsRecvd   int64 `json:"vNetworkErrorsRecvd"` // Receive errors
	NetworkErrorsRecvdT  int64 `json:"vNetworkErrorsRecvdT"`
	NetworkErrorsSent    int64 `json:"vNetworkErrorsSent"` // Transmit errors
	NetworkErrorsSentT   int64 `json:"vNetworkErrorsSentT"`
	NetworkDropsRecvd    int64 `json:"vNetworkDropsRecvd"` // Dropped on receive
	NetworkDropsRecvdT   int64 `json:"vNetworkDropsRecvdT"`
	NetworkDropsSent     int64 `json:"vNetworkDropsSent"` // Dropped on transmit
	NetworkDropsSentT    int64 `json:"vNetworkDropsSentT"`

	// ==========================================================================
	// Container-Level Metrics
	// Source: cgroup v1 or v2
	// ==========================================================================

	ContainerCPUTime            int64  `json:"cCpuTime"` // Total CPU time (ns)
	ContainerCPUTimeT           int64  `json:"cCpuTimeT"`
	ContainerCPUTimeUserMode    int64  `json:"cCpuTimeUserMode"` // User mode CPU time (cs)
	ContainerCPUTimeUserModeT   int64  `json:"cCpuTimeUserModeT"`
	ContainerCPUTimeKernelMode  int64  `json:"cCpuTimeKernelMode"` // Kernel mode CPU time (cs)
	ContainerCPUTimeKernelModeT int64  `json:"cCpuTimeKernelModeT"`
	ContainerMemoryUsed         int64  `json:"cMemoryUsed"` // Current memory usage (bytes)
	ContainerMemoryUsedT        int64  `json:"cMemoryUsedT"`
	ContainerMemoryMaxUsed      int64  `json:"cMemoryMaxUsed"` // Peak memory usage (bytes)
	ContainerMemoryMaxUsedT     int64  `json:"cMemoryMaxUsedT"`
	ContainerDiskReadBytes      int64  `json:"cDiskReadBytes"` // Bytes read
	ContainerDiskReadBytesT     int64  `json:"cDiskReadBytesT"`
	ContainerDiskWriteBytes     int64  `json:"cDiskWriteBytes"` // Bytes written
	ContainerDiskWriteBytesT    int64  `json:"cDiskWriteBytesT"`
	ContainerDiskSectorIO       int64  `json:"cDiskSectorIO"` // Total sector I/O (v1 only)
	ContainerDiskSectorIOT      int64  `json:"cDiskSectorIOT"`
	ContainerNetworkBytesRecvd  int64  `json:"cNetworkBytesRecvd"` // Network bytes received
	ContainerNetworkBytesRecvdT int64  `json:"cNetworkBytesRecvdT"`
	ContainerNetworkBytesSent   int64  `json:"cNetworkBytesSent"` // Network bytes sent
	ContainerNetworkBytesSentT  int64  `json:"cNetworkBytesSentT"`
	ContainerPgFault            int64  `json:"cPgFault"` // Page faults
	ContainerPgFaultT           int64  `json:"cPgFaultT"`
	ContainerMajorPgFault       int64  `json:"cMajorPgFault"` // Major page faults
	ContainerMajorPgFaultT      int64  `json:"cMajorPgFaultT"`
	ContainerNumProcesses       int64  `json:"cNumProcesses"` // Process count
	ContainerNumProcessesT      int64  `json:"cNumProcessesT"`
	ContainerPerCPUTimesJSON    string `json:"cCpuPerCpuJson,omitempty"` // Per-CPU times (JSON, v1)
	ContainerPerCPUTimesT       int64  `json:"cCpuPerCpuT,omitempty"`

	// ==========================================================================
	// NVIDIA GPU Metrics
	// Collected separately, flattened for export
	// ==========================================================================

	NvidiaGPUs []NvidiaGPUDynamic `json:"-"`

	// ==========================================================================
	// vLLM Inference Server Metrics
	// Source: Prometheus endpoint
	// ==========================================================================

	VLLMAvailable              bool    `json:"vllmAvailable"`
	VLLMTimestamp              int64   `json:"vllmTimestamp"`
	VLLMRequestsRunning        float64 `json:"vllmRequestsRunning"`
	VLLMRequestsWaiting        float64 `json:"vllmRequestsWaiting"`
	VLLMEngineSleepState       float64 `json:"vllmEngineSleepState"`
	VLLMPreemptionsTotal       float64 `json:"vllmPreemptionsTotal"`
	VLLMKvCacheUsagePercent    float64 `json:"vllmKvCacheUsagePercent"`
	VLLMPrefixCacheHits        float64 `json:"vllmPrefixCacheHits"`
	VLLMPrefixCacheQueries     float64 `json:"vllmPrefixCacheQueries"`
	VLLMRequestsFinishedTotal  float64 `json:"vllmRequestsFinishedTotal"`
	VLLMRequestsCorruptedTotal float64 `json:"vllmRequestsCorruptedTotal"`
	VLLMTokensPromptTotal      float64 `json:"vllmTokensPromptTotal"`
	VLLMTokensGenerationTotal  float64 `json:"vllmTokensGenerationTotal"`
	VLLMLatencyTtftSum         float64 `json:"vllmLatencyTtftSum"`
	VLLMLatencyTtftCount       float64 `json:"vllmLatencyTtftCount"`
	VLLMLatencyE2eSum          float64 `json:"vllmLatencyE2eSum"`
	VLLMLatencyE2eCount        float64 `json:"vllmLatencyE2eCount"`
	VLLMLatencyQueueSum        float64 `json:"vllmLatencyQueueSum"`
	VLLMLatencyQueueCount      float64 `json:"vllmLatencyQueueCount"`
	VLLMLatencyInferenceSum    float64 `json:"vllmLatencyInferenceSum"`
	VLLMLatencyInferenceCount  float64 `json:"vllmLatencyInferenceCount"`
	VLLMLatencyPrefillSum      float64 `json:"vllmLatencyPrefillSum"`
	VLLMLatencyPrefillCount    float64 `json:"vllmLatencyPrefillCount"`
	VLLMLatencyDecodeSum       float64 `json:"vllmLatencyDecodeSum"`
	VLLMLatencyDecodeCount     float64 `json:"vllmLatencyDecodeCount"`
	VLLMHistogramsJSON         string  `json:"vllmHistogramsJson,omitempty"`

	// ==========================================================================
	// Process-Level Metrics
	// Collected separately, flattened for export
	// ==========================================================================

	Processes []ProcessMetrics `json:"-"`
}

// VLLMHistograms holds histogram data for vLLM metrics
type VLLMHistograms struct {
	TokensPerStep           map[string]float64 `json:"tokensPerStep,omitempty"`
	LatencyTtft             map[string]float64 `json:"latencyTtft,omitempty"`
	LatencyE2e              map[string]float64 `json:"latencyE2e,omitempty"`
	LatencyQueue            map[string]float64 `json:"latencyQueue,omitempty"`
	LatencyInference        map[string]float64 `json:"latencyInference,omitempty"`
	LatencyPrefill          map[string]float64 `json:"latencyPrefill,omitempty"`
	LatencyDecode           map[string]float64 `json:"latencyDecode,omitempty"`
	LatencyInterToken       map[string]float64 `json:"latencyInterToken,omitempty"`
	ReqSizePromptTokens     map[string]float64 `json:"reqSizePromptTokens,omitempty"`
	ReqSizeGenerationTokens map[string]float64 `json:"reqSizeGenerationTokens,omitempty"`
	ReqParamsMaxTokens      map[string]float64 `json:"reqParamsMaxTokens,omitempty"`
	ReqParamsN              map[string]float64 `json:"reqParamsN,omitempty"`
}
