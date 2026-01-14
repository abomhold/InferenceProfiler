package collectors

// DynamicMetrics contains all dynamic metrics collected during profiling.
type DynamicMetrics struct {
	// Timestamp is the overall collection cycle timestamp (nanoseconds since Unix epoch)
	Timestamp int64 `json:"timestamp"`

	// VM-Level
	CPUTime               int64   `json:"vCpuTime"`
	CPUTimeT              int64   `json:"vCpuTimeT"`
	CPUTimeUserMode       int64   `json:"vCpuTimeUserMode"`
	CPUTimeUserModeT      int64   `json:"vCpuTimeUserModeT"`
	CPUTimeKernelMode     int64   `json:"vCpuTimeKernelMode"`
	CPUTimeKernelModeT    int64   `json:"vCpuTimeKernelModeT"`
	CPUIdleTime           int64   `json:"vCpuIdleTime"`
	CPUIdleTimeT          int64   `json:"vCpuIdleTimeT"`
	CPUTimeIOWait         int64   `json:"vCpuTimeIOWait"`
	CPUTimeIOWaitT        int64   `json:"vCpuTimeIOWaitT"`
	CPUTimeIntSrvc        int64   `json:"vCpuTimeIntSrvc"`
	CPUTimeIntSrvcT       int64   `json:"vCpuTimeIntSrvcT"`
	CPUTimeSoftIntSrvc    int64   `json:"vCpuTimeSoftIntSrvc"`
	CPUTimeSoftIntSrvcT   int64   `json:"vCpuTimeSoftIntSrvcT"`
	CPUNice               int64   `json:"vCpuNice"`
	CPUNiceT              int64   `json:"vCpuNiceT"`
	CPUSteal              int64   `json:"vCpuSteal"`
	CPUStealT             int64   `json:"vCpuStealT"`
	CPUContextSwitches    int64   `json:"vCpuContextSwitches"`
	CPUContextSwitchesT   int64   `json:"vCpuContextSwitchesT"`
	LoadAvg               float64 `json:"vLoadAvg"`
	LoadAvgT              int64   `json:"vLoadAvgT"`
	CPUMhz                float64 `json:"vCpuMhz"`
	CPUMhzT               int64   `json:"vCpuMhzT"`
	MemoryTotal           int64   `json:"vMemoryTotal"`
	MemoryTotalT          int64   `json:"vMemoryTotalT"`
	MemoryFree            int64   `json:"vMemoryFree"`
	MemoryFreeT           int64   `json:"vMemoryFreeT"`
	MemoryUsed            int64   `json:"vMemoryUsed"`
	MemoryUsedT           int64   `json:"vMemoryUsedT"`
	MemoryBuffers         int64   `json:"vMemoryBuffers"`
	MemoryBuffersT        int64   `json:"vMemoryBuffersT"`
	MemoryCached          int64   `json:"vMemoryCached"`
	MemoryCachedT         int64   `json:"vMemoryCachedT"`
	MemoryPercent         float64 `json:"vMemoryPercent"`
	MemoryPercentT        int64   `json:"vMemoryPercentT"`
	MemorySwapTotal       int64   `json:"vMemorySwapTotal"`
	MemorySwapTotalT      int64   `json:"vMemorySwapTotalT"`
	MemorySwapFree        int64   `json:"vMemorySwapFree"`
	MemorySwapFreeT       int64   `json:"vMemorySwapFreeT"`
	MemorySwapUsed        int64   `json:"vMemorySwapUsed"`
	MemorySwapUsedT       int64   `json:"vMemorySwapUsedT"`
	MemoryPgFault         int64   `json:"vMemoryPgFault"`
	MemoryPgFaultT        int64   `json:"vMemoryPgFaultT"`
	MemoryMajorPageFault  int64   `json:"vMemoryMajorPageFault"`
	MemoryMajorPageFaultT int64   `json:"vMemoryMajorPageFaultT"`
	DiskSectorReads       int64   `json:"vDiskSectorReads"`
	DiskSectorReadsT      int64   `json:"vDiskSectorReadsT"`
	DiskSectorWrites      int64   `json:"vDiskSectorWrites"`
	DiskSectorWritesT     int64   `json:"vDiskSectorWritesT"`
	DiskReadBytes         int64   `json:"vDiskReadBytes"`
	DiskReadBytesT        int64   `json:"vDiskReadBytesT"`
	DiskWriteBytes        int64   `json:"vDiskWriteBytes"`
	DiskWriteBytesT       int64   `json:"vDiskWriteBytesT"`
	DiskSuccessfulReads   int64   `json:"vDiskSuccessfulReads"`
	DiskSuccessfulReadsT  int64   `json:"vDiskSuccessfulReadsT"`
	DiskSuccessfulWrites  int64   `json:"vDiskSuccessfulWrites"`
	DiskSuccessfulWritesT int64   `json:"vDiskSuccessfulWritesT"`
	DiskMergedReads       int64   `json:"vDiskMergedReads"`
	DiskMergedReadsT      int64   `json:"vDiskMergedReadsT"`
	DiskMergedWrites      int64   `json:"vDiskMergedWrites"`
	DiskMergedWritesT     int64   `json:"vDiskMergedWritesT"`
	DiskReadTime          int64   `json:"vDiskReadTime"`
	DiskReadTimeT         int64   `json:"vDiskReadTimeT"`
	DiskWriteTime         int64   `json:"vDiskWriteTime"`
	DiskWriteTimeT        int64   `json:"vDiskWriteTimeT"`
	DiskIOInProgress      int64   `json:"vDiskIOInProgress"`
	DiskIOInProgressT     int64   `json:"vDiskIOInProgressT"`
	DiskIOTime            int64   `json:"vDiskIOTime"`
	DiskIOTimeT           int64   `json:"vDiskIOTimeT"`
	DiskWeightedIOTime    int64   `json:"vDiskWeightedIOTime"`
	DiskWeightedIOTimeT   int64   `json:"vDiskWeightedIOTimeT"`
	NetworkBytesRecvd     int64   `json:"vNetworkBytesRecvd"`
	NetworkBytesRecvdT    int64   `json:"vNetworkBytesRecvdT"`
	NetworkBytesSent      int64   `json:"vNetworkBytesSent"`
	NetworkBytesSentT     int64   `json:"vNetworkBytesSentT"`
	NetworkPacketsRecvd   int64   `json:"vNetworkPacketsRecvd"`
	NetworkPacketsRecvdT  int64   `json:"vNetworkPacketsRecvdT"`
	NetworkPacketsSent    int64   `json:"vNetworkPacketsSent"`
	NetworkPacketsSentT   int64   `json:"vNetworkPacketsSentT"`
	NetworkErrorsRecvd    int64   `json:"vNetworkErrorsRecvd"`
	NetworkErrorsRecvdT   int64   `json:"vNetworkErrorsRecvdT"`
	NetworkErrorsSent     int64   `json:"vNetworkErrorsSent"`
	NetworkErrorsSentT    int64   `json:"vNetworkErrorsSentT"`
	NetworkDropsRecvd     int64   `json:"vNetworkDropsRecvd"`
	NetworkDropsRecvdT    int64   `json:"vNetworkDropsRecvdT"`
	NetworkDropsSent      int64   `json:"vNetworkDropsSent"`
	NetworkDropsSentT     int64   `json:"vNetworkDropsSentT"`

	// Container-Level Metrics
	ContainerCPUTime            int64  `json:"cCpuTime"`
	ContainerCPUTimeT           int64  `json:"cCpuTimeT"`
	ContainerCPUTimeUserMode    int64  `json:"cCpuTimeUserMode"`
	ContainerCPUTimeUserModeT   int64  `json:"cCpuTimeUserModeT"`
	ContainerCPUTimeKernelMode  int64  `json:"cCpuTimeKernelMode"`
	ContainerCPUTimeKernelModeT int64  `json:"cCpuTimeKernelModeT"`
	ContainerMemoryUsed         int64  `json:"cMemoryUsed"`
	ContainerMemoryUsedT        int64  `json:"cMemoryUsedT"`
	ContainerMemoryMaxUsed      int64  `json:"cMemoryMaxUsed"`
	ContainerMemoryMaxUsedT     int64  `json:"cMemoryMaxUsedT"`
	ContainerDiskReadBytes      int64  `json:"cDiskReadBytes"`
	ContainerDiskReadBytesT     int64  `json:"cDiskReadBytesT"`
	ContainerDiskWriteBytes     int64  `json:"cDiskWriteBytes"`
	ContainerDiskWriteBytesT    int64  `json:"cDiskWriteBytesT"`
	ContainerDiskSectorIO       int64  `json:"cDiskSectorIO"`
	ContainerDiskSectorIOT      int64  `json:"cDiskSectorIOT"`
	ContainerNetworkBytesRecvd  int64  `json:"cNetworkBytesRecvd"`
	ContainerNetworkBytesRecvdT int64  `json:"cNetworkBytesRecvdT"`
	ContainerNetworkBytesSent   int64  `json:"cNetworkBytesSent"`
	ContainerNetworkBytesSentT  int64  `json:"cNetworkBytesSentT"`
	ContainerPgFault            int64  `json:"cPgFault"`
	ContainerPgFaultT           int64  `json:"cPgFaultT"`
	ContainerMajorPgFault       int64  `json:"cMajorPgFault"`
	ContainerMajorPgFaultT      int64  `json:"cMajorPgFaultT"`
	ContainerNumProcesses       int64  `json:"cNumProcesses"`
	ContainerNumProcessesT      int64  `json:"cNumProcessesT"`
	ContainerPerCPUTimesJSON    string `json:"cCpuPerCpuJson,omitempty"`
	ContainerPerCPUTimesT       int64  `json:"cCpuPerCpuT,omitempty"`

	// Process-Level Metrics
	Processes []ProcessMetrics `json:"-"`

	// NVIDIA GPU Metrics
	NvidiaGPUs []NvidiaGPUDynamic `json:"-"`

	// vLLM Inference Server Metrics
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
	VLLMHistogramsJSON         string  `json:"vllmHistogramsJson"`
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
