package collectors

// StaticMetrics contains all static system information - collected once at startup
type StaticMetrics struct {
	// Session identification
	UUID     string `json:"uuid"`
	VMID     string `json:"vId"`
	Hostname string `json:"vHostname"`
	BootTime int64  `json:"vBootTime"`

	// CPU static info
	NumProcessors int    `json:"vNumProcessors"`
	CPUType       string `json:"vCpuType"`
	CPUCacheL1d   int64  `json:"vCpuCacheL1d"`
	CPUCacheL1i   int64  `json:"vCpuCacheL1i"`
	CPUCacheL2    int64  `json:"vCpuCacheL2"`
	CPUCacheL3    int64  `json:"vCpuCacheL3"`

	// Kernel info
	SystemName    string `json:"vSystemName"`
	NodeName      string `json:"vNodeName"`
	KernelRelease string `json:"vRelease"`
	KernelVersion string `json:"vVersion"`
	Machine       string `json:"vMachine"`

	// Memory static info
	MemoryTotalBytes int64 `json:"vMemoryTotalBytes"`
	SwapTotalBytes   int64 `json:"vSwapTotalBytes"`

	// Container static info
	ContainerID      string `json:"cId"`
	ContainerNumCPUs int64  `json:"cNumProcessors"`
	CgroupVersion    int64  `json:"cCgroupVersion"`

	// Network static info (JSON string)
	NetworkInterfacesJSON string `json:"networkInterfaces,omitempty"`

	// NVIDIA static info
	NvidiaDriverVersion string `json:"nvidiaDriverVersion,omitempty"`
	NvidiaCudaVersion   string `json:"nvidiaCudaVersion,omitempty"`
	NvidiaGPUsJSON      string `json:"nvidiaGpus,omitempty"`

	// Disk static info (JSON string)
	DisksJSON string `json:"disks,omitempty"`
}

// NetworkInterfaceStatic contains static info for a network interface
type NetworkInterfaceStatic struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	State     string `json:"state"`
	MTU       int64  `json:"mtu"`
	SpeedMbps int64  `json:"speedMbps,omitempty"`
}

// NvidiaGPUStatic contains static info for a single GPU
type NvidiaGPUStatic struct {
	Index         int    `json:"index"`
	Name          string `json:"name"`
	UUID          string `json:"uuid"`
	TotalMemoryMb int64  `json:"totalMemoryMb"`
}

// DiskStatic contains static info for a single disk
type DiskStatic struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Vendor    string `json:"vendor"`
	SizeBytes int64  `json:"sizeBytes"`
}

// DynamicMetrics contains all dynamic metrics
// GPU metrics stored as slice during collection, flattened at export time
// Process metrics stored as JSON string
type DynamicMetrics struct {
	Timestamp int64 `json:"timestamp"`

	// CPU dynamic metrics
	CPUTimeUserMode     int64   `json:"vCpuTimeUserMode"`
	CPUTimeUserModeT    int64   `json:"vCpuTimeUserModeT"`
	CPUTimeKernelMode   int64   `json:"vCpuTimeKernelMode"`
	CPUTimeKernelModeT  int64   `json:"vCpuTimeKernelModeT"`
	CPUIdleTime         int64   `json:"vCpuIdleTime"`
	CPUIdleTimeT        int64   `json:"vCpuIdleTimeT"`
	CPUTimeIOWait       int64   `json:"vCpuTimeIOWait"`
	CPUTimeIOWaitT      int64   `json:"vCpuTimeIOWaitT"`
	CPUTimeIntSrvc      int64   `json:"vCpuTimeIntSrvc"`
	CPUTimeIntSrvcT     int64   `json:"vCpuTimeIntSrvcT"`
	CPUTimeSoftIntSrvc  int64   `json:"vCpuTimeSoftIntSrvc"`
	CPUTimeSoftIntSrvcT int64   `json:"vCpuTimeSoftIntSrvcT"`
	CPUNice             int64   `json:"vCpuNice"`
	CPUNiceT            int64   `json:"vCpuNiceT"`
	CPUSteal            int64   `json:"vCpuSteal"`
	CPUStealT           int64   `json:"vCpuStealT"`
	CPUTime             int64   `json:"vCpuTime"`
	CPUTimeT            int64   `json:"vCpuTimeT"`
	CPUContextSwitches  int64   `json:"vCpuContextSwitches"`
	CPUContextSwitchesT int64   `json:"vCpuContextSwitchesT"`
	LoadAvg             float64 `json:"vLoadAvg"`
	LoadAvgT            int64   `json:"vLoadAvgT"`
	CPUMhz              float64 `json:"vCpuMhz"`
	CPUMhzT             int64   `json:"vCpuMhzT"`

	// Memory dynamic metrics
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
	MemoryPgFault         int64   `json:"vMemoryPgFault"`
	MemoryPgFaultT        int64   `json:"vMemoryPgFaultT"`
	MemoryMajorPageFault  int64   `json:"vMemoryMajorPageFault"`
	MemoryMajorPageFaultT int64   `json:"vMemoryMajorPageFaultT"`
	MemorySwapTotal       int64   `json:"vMemorySwapTotal"`
	MemorySwapTotalT      int64   `json:"vMemorySwapTotalT"`
	MemorySwapFree        int64   `json:"vMemorySwapFree"`
	MemorySwapFreeT       int64   `json:"vMemorySwapFreeT"`
	MemorySwapUsed        int64   `json:"vMemorySwapUsed"`
	MemorySwapUsedT       int64   `json:"vMemorySwapUsedT"`

	// Disk dynamic metrics
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

	// Network dynamic metrics
	NetworkBytesRecvd    int64 `json:"vNetworkBytesRecvd"`
	NetworkBytesRecvdT   int64 `json:"vNetworkBytesRecvdT"`
	NetworkBytesSent     int64 `json:"vNetworkBytesSent"`
	NetworkBytesSentT    int64 `json:"vNetworkBytesSentT"`
	NetworkPacketsRecvd  int64 `json:"vNetworkPacketsRecvd"`
	NetworkPacketsRecvdT int64 `json:"vNetworkPacketsRecvdT"`
	NetworkPacketsSent   int64 `json:"vNetworkPacketsSent"`
	NetworkPacketsSentT  int64 `json:"vNetworkPacketsSentT"`
	NetworkErrorsRecvd   int64 `json:"vNetworkErrorsRecvd"`
	NetworkErrorsRecvdT  int64 `json:"vNetworkErrorsRecvdT"`
	NetworkErrorsSent    int64 `json:"vNetworkErrorsSent"`
	NetworkErrorsSentT   int64 `json:"vNetworkErrorsSentT"`
	NetworkDropsRecvd    int64 `json:"vNetworkDropsRecvd"`
	NetworkDropsRecvdT   int64 `json:"vNetworkDropsRecvdT"`
	NetworkDropsSent     int64 `json:"vNetworkDropsSent"`
	NetworkDropsSentT    int64 `json:"vNetworkDropsSentT"`

	// Container dynamic metrics
	ContainerNetworkBytesRecvd  int64 `json:"cNetworkBytesRecvd"`
	ContainerNetworkBytesRecvdT int64 `json:"cNetworkBytesRecvdT"`
	ContainerNetworkBytesSent   int64 `json:"cNetworkBytesSent"`
	ContainerNetworkBytesSentT  int64 `json:"cNetworkBytesSentT"`
	ContainerCPUTime            int64 `json:"cCpuTime"`
	ContainerCPUTimeT           int64 `json:"cCpuTimeT"`
	ContainerCPUTimeUserMode    int64 `json:"cCpuTimeUserMode"`
	ContainerCPUTimeUserModeT   int64 `json:"cCpuTimeUserModeT"`
	ContainerCPUTimeKernelMode  int64 `json:"cCpuTimeKernelMode"`
	ContainerCPUTimeKernelModeT int64 `json:"cCpuTimeKernelModeT"`
	ContainerMemoryUsed         int64 `json:"cMemoryUsed"`
	ContainerMemoryUsedT        int64 `json:"cMemoryUsedT"`
	ContainerMemoryMaxUsed      int64 `json:"cMemoryMaxUsed"`
	ContainerMemoryMaxUsedT     int64 `json:"cMemoryMaxUsedT"`
	ContainerDiskReadBytes      int64 `json:"cDiskReadBytes"`
	ContainerDiskReadBytesT     int64 `json:"cDiskReadBytesT"`
	ContainerDiskWriteBytes     int64 `json:"cDiskWriteBytes"`
	ContainerDiskWriteBytesT    int64 `json:"cDiskWriteBytesT"`

	// NVIDIA GPU metrics - stored as slice, flattened dynamically at export
	// Exported as nvidia{i}UtilizationGpu, nvidia{i}MemoryUsedMb, etc.
	NvidiaGPUs []NvidiaGPUDynamic `json:"-"`

	// vLLM metrics - scalar fields flattened, histograms as JSON
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

	// Process metrics - stored as slice, flattened dynamically at export
	// Exported as process{i}Pid, process{i}Name, process{i}CpuTimeUserMode, etc.
	Processes []ProcessMetrics `json:"-"`
}

// NvidiaGPUDynamic contains dynamic metrics for a single GPU
type NvidiaGPUDynamic struct {
	Index             int     `json:"index"`
	UtilizationGPU    int64   `json:"utilizationGpu"`
	UtilizationGPUT   int64   `json:"utilizationGpuT"`
	UtilizationMem    int64   `json:"utilizationMem"`
	UtilizationMemT   int64   `json:"utilizationMemT"`
	MemoryUsedMb      int64   `json:"memoryUsedMb"`
	MemoryUsedMbT     int64   `json:"memoryUsedMbT"`
	MemoryFreeMb      int64   `json:"memoryFreeMb"`
	MemoryFreeMbT     int64   `json:"memoryFreeMbT"`
	Bar1UsedMb        int64   `json:"bar1UsedMb"`
	Bar1UsedMbT       int64   `json:"bar1UsedMbT"`
	TemperatureC      int64   `json:"temperatureC"`
	TemperatureCT     int64   `json:"temperatureCT"`
	FanSpeed          int64   `json:"fanSpeed"`
	FanSpeedT         int64   `json:"fanSpeedT"`
	ClockGraphicsMhz  int64   `json:"clockGraphicsMhz"`
	ClockGraphicsMhzT int64   `json:"clockGraphicsMhzT"`
	ClockSmMhz        int64   `json:"clockSmMhz"`
	ClockSmMhzT       int64   `json:"clockSmMhzT"`
	ClockMemMhz       int64   `json:"clockMemMhz"`
	ClockMemMhzT      int64   `json:"clockMemMhzT"`
	PcieTxKbps        int64   `json:"pcieTxKbps"`
	PcieTxKbpsT       int64   `json:"pcieTxKbpsT"`
	PcieRxKbps        int64   `json:"pcieRxKbps"`
	PcieRxKbpsT       int64   `json:"pcieRxKbpsT"`
	PowerDrawW        float64 `json:"powerDrawW"`
	PowerDrawWT       int64   `json:"powerDrawWT"`
	PerfState         string  `json:"perfState"`
	PerfStateT        int64   `json:"perfStateT"`
	ProcessCount      int64   `json:"processCount"`
	ProcessCountT     int64   `json:"processCountT"`
	// GPU processes kept as JSON within each GPU
	ProcessesJSON string `json:"processesJson,omitempty"`
}

// GPUProcess represents a process using GPU memory
type GPUProcess struct {
	PID          uint32 `json:"pid"`
	Name         string `json:"name"`
	UsedMemoryMb int64  `json:"usedMemoryMb"`
}

// ProcessMetrics contains metrics for a single process
type ProcessMetrics struct {
	PID                          int64  `json:"pId"`
	PIDT                         int64  `json:"pIdT"`
	Name                         string `json:"pName"`
	NameT                        int64  `json:"pNameT"`
	Cmdline                      string `json:"pCmdline"`
	CmdlineT                     int64  `json:"pCmdlineT"`
	NumThreads                   int64  `json:"pNumThreads"`
	NumThreadsT                  int64  `json:"pNumThreadsT"`
	CPUTimeUserMode              int64  `json:"pCpuTimeUserMode"`
	CPUTimeUserModeT             int64  `json:"pCpuTimeUserModeT"`
	CPUTimeKernelMode            int64  `json:"pCpuTimeKernelMode"`
	CPUTimeKernelModeT           int64  `json:"pCpuTimeKernelModeT"`
	ChildrenUserMode             int64  `json:"pChildrenUserMode"`
	ChildrenUserModeT            int64  `json:"pChildrenUserModeT"`
	ChildrenKernelMode           int64  `json:"pChildrenKernelMode"`
	ChildrenKernelModeT          int64  `json:"pChildrenKernelModeT"`
	VoluntaryContextSwitches     int64  `json:"pVoluntaryContextSwitches"`
	VoluntaryContextSwitchesT    int64  `json:"pVoluntaryContextSwitchesT"`
	NonvoluntaryContextSwitches  int64  `json:"pNonvoluntaryContextSwitches"`
	NonvoluntaryContextSwitchesT int64  `json:"pNonvoluntaryContextSwitchesT"`
	BlockIODelays                int64  `json:"pBlockIODelays"`
	BlockIODelaysT               int64  `json:"pBlockIODelaysT"`
	VirtualMemoryBytes           int64  `json:"pVirtualMemoryBytes"`
	VirtualMemoryBytesT          int64  `json:"pVirtualMemoryBytesT"`
	ResidentSetSize              int64  `json:"pResidentSetSize"`
	ResidentSetSizeT             int64  `json:"pResidentSetSizeT"`
}

// VLLMHistograms holds all histogram data
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
