// Package collectors provides metric collection functionality.
package collectors

// StaticMetrics contains all static system information collected once at startup.
type StaticMetrics struct {
	UUID     string `json:"uuid"`
	VMID     string `json:"vId"`
	Hostname string `json:"vHostname"`
	BootTime int64  `json:"vBootTime"`

	NumProcessors       int     `json:"vNumProcessors"`
	CPUType             string  `json:"vCpuType"`
	CPUCache            string  `json:"vCpuCache"`
	KernelInfo          string  `json:"vKernelInfo"`
	TimeSynced          bool    `json:"vTimeSynced"`
	TimeOffsetSeconds   float64 `json:"vTimeOffsetSeconds"`
	TimeMaxErrorSeconds float64 `json:"vTimeMaxErrorSeconds"`

	MemoryTotalBytes int64 `json:"vMemoryTotalBytes"`
	SwapTotalBytes   int64 `json:"vSwapTotalBytes"`

	DisksJSON             string `json:"disks,omitempty"`
	NetworkInterfacesJSON string `json:"networkInterfaces,omitempty"`

	ContainerID      string `json:"cId,omitempty"`
	ContainerNumCPUs int64  `json:"cNumProcessors,omitempty"`
	CgroupVersion    int64  `json:"cCgroupVersion,omitempty"`

	NvidiaDriverVersion string `json:"nvidiaDriverVersion,omitempty"`
	NvidiaCudaVersion   string `json:"nvidiaCudaVersion,omitempty"`
	NvmlVersion         string `json:"nvmlVersion,omitempty"`
	NvidiaGPUCount      int    `json:"nvidiaGpuCount,omitempty"`
	NvidiaGPUsJSON      string `json:"nvidiaGpus,omitempty"`
}

// DynamicMetrics contains all dynamic metrics collected during profiling.
type DynamicMetrics struct {
	Timestamp int64 `json:"timestamp"`

	// CPU metrics
	CPUTime             int64   `json:"vCpuTime"`
	CPUTimeT            int64   `json:"vCpuTimeT"`
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
	CPUContextSwitches  int64   `json:"vCpuContextSwitches"`
	CPUContextSwitchesT int64   `json:"vCpuContextSwitchesT"`
	LoadAvg             float64 `json:"vLoadAvg"`
	LoadAvgT            int64   `json:"vLoadAvgT"`
	CPUMhz              float64 `json:"vCpuMhz"`
	CPUMhzT             int64   `json:"vCpuMhzT"`

	// Memory metrics
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

	// Disk metrics
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

	// Network metrics
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

	// Container metrics
	ContainerCPUTime            int64  `json:"cCpuTime,omitempty"`
	ContainerCPUTimeT           int64  `json:"cCpuTimeT,omitempty"`
	ContainerCPUTimeUserMode    int64  `json:"cCpuTimeUserMode,omitempty"`
	ContainerCPUTimeUserModeT   int64  `json:"cCpuTimeUserModeT,omitempty"`
	ContainerCPUTimeKernelMode  int64  `json:"cCpuTimeKernelMode,omitempty"`
	ContainerCPUTimeKernelModeT int64  `json:"cCpuTimeKernelModeT,omitempty"`
	ContainerPerCPUTimesJSON    string `json:"cCpuPerCpuJson,omitempty"`
	ContainerPerCPUTimesT       int64  `json:"cCpuPerCpuT,omitempty"`
	ContainerMemoryUsed         int64  `json:"cMemoryUsed,omitempty"`
	ContainerMemoryUsedT        int64  `json:"cMemoryUsedT,omitempty"`
	ContainerMemoryMaxUsed      int64  `json:"cMemoryMaxUsed,omitempty"`
	ContainerMemoryMaxUsedT     int64  `json:"cMemoryMaxUsedT,omitempty"`
	ContainerPgFault            int64  `json:"cPgFault,omitempty"`
	ContainerPgFaultT           int64  `json:"cPgFaultT,omitempty"`
	ContainerMajorPgFault       int64  `json:"cMajorPgFault,omitempty"`
	ContainerMajorPgFaultT      int64  `json:"cMajorPgFaultT,omitempty"`
	ContainerDiskReadBytes      int64  `json:"cDiskReadBytes,omitempty"`
	ContainerDiskReadBytesT     int64  `json:"cDiskReadBytesT,omitempty"`
	ContainerDiskWriteBytes     int64  `json:"cDiskWriteBytes,omitempty"`
	ContainerDiskWriteBytesT    int64  `json:"cDiskWriteBytesT,omitempty"`
	ContainerDiskSectorIO       int64  `json:"cDiskSectorIO,omitempty"`
	ContainerDiskSectorIOT      int64  `json:"cDiskSectorIOT,omitempty"`
	ContainerNetworkBytesRecvd  int64  `json:"cNetworkBytesRecvd,omitempty"`
	ContainerNetworkBytesRecvdT int64  `json:"cNetworkBytesRecvdT,omitempty"`
	ContainerNetworkBytesSent   int64  `json:"cNetworkBytesSent,omitempty"`
	ContainerNetworkBytesSentT  int64  `json:"cNetworkBytesSentT,omitempty"`
	ContainerNumProcesses       int64  `json:"cProcessCount,omitempty"`
	ContainerNumProcessesT      int64  `json:"cProcessCountT,omitempty"`

	// GPU metrics (expanded during JSON marshal)
	GPUs []GPUDynamic `json:"-"`

	// vLLM metrics
	VLLMAvailable              bool    `json:"vllmAvailable,omitempty"`
	VLLMTimestamp              int64   `json:"vllmTimestamp,omitempty"`
	VLLMRequestsRunning        float64 `json:"vllmRequestsRunning,omitempty"`
	VLLMRequestsWaiting        float64 `json:"vllmRequestsWaiting,omitempty"`
	VLLMEngineSleepState       float64 `json:"vllmEngineSleepState,omitempty"`
	VLLMPreemptionsTotal       float64 `json:"vllmPreemptionsTotal,omitempty"`
	VLLMKvCacheUsagePercent    float64 `json:"vllmKvCacheUsagePercent,omitempty"`
	VLLMPrefixCacheHits        float64 `json:"vllmPrefixCacheHits,omitempty"`
	VLLMPrefixCacheQueries     float64 `json:"vllmPrefixCacheQueries,omitempty"`
	VLLMRequestsFinishedTotal  float64 `json:"vllmRequestsFinishedTotal,omitempty"`
	VLLMRequestsCorruptedTotal float64 `json:"vllmRequestsCorruptedTotal,omitempty"`
	VLLMTokensPromptTotal      float64 `json:"vllmTokensPromptTotal,omitempty"`
	VLLMTokensGenerationTotal  float64 `json:"vllmTokensGenerationTotal,omitempty"`
	VLLMLatencyTtftSum         float64 `json:"vllmLatencyTtftSum,omitempty"`
	VLLMLatencyTtftCount       float64 `json:"vllmLatencyTtftCount,omitempty"`
	VLLMLatencyE2eSum          float64 `json:"vllmLatencyE2eSum,omitempty"`
	VLLMLatencyE2eCount        float64 `json:"vllmLatencyE2eCount,omitempty"`
	VLLMLatencyQueueSum        float64 `json:"vllmLatencyQueueSum,omitempty"`
	VLLMLatencyQueueCount      float64 `json:"vllmLatencyQueueCount,omitempty"`
	VLLMLatencyInferenceSum    float64 `json:"vllmLatencyInferenceSum,omitempty"`
	VLLMLatencyInferenceCount  float64 `json:"vllmLatencyInferenceCount,omitempty"`
	VLLMLatencyPrefillSum      float64 `json:"vllmLatencyPrefillSum,omitempty"`
	VLLMLatencyPrefillCount    float64 `json:"vllmLatencyPrefillCount,omitempty"`
	VLLMLatencyDecodeSum       float64 `json:"vllmLatencyDecodeSum,omitempty"`
	VLLMLatencyDecodeCount     float64 `json:"vllmLatencyDecodeCount,omitempty"`
	VLLMHistogramsJSON         string  `json:"vllmHistogramsJson,omitempty"`

	// Process metrics (expanded during JSON marshal)
	Processes []ProcessInfo `json:"-"`
}

// GPUDynamic contains dynamic metrics for a single GPU.
type GPUDynamic struct {
	Index                     int    `json:"index"`
	UtilizationGPU            int64  `json:"utilizationGpu"`
	UtilizationGPUT           int64  `json:"utilizationGpuT"`
	UtilizationMemory         int64  `json:"utilizationMemory"`
	UtilizationMemoryT        int64  `json:"utilizationMemoryT"`
	UtilizationEncoder        int64  `json:"utilizationEncoder"`
	UtilizationEncoderT       int64  `json:"utilizationEncoderT"`
	UtilizationDecoder        int64  `json:"utilizationDecoder"`
	UtilizationDecoderT       int64  `json:"utilizationDecoderT"`
	UtilizationJpeg           int64  `json:"utilizationJpeg,omitempty"`
	UtilizationJpegT          int64  `json:"utilizationJpegT,omitempty"`
	UtilizationOfa            int64  `json:"utilizationOfa,omitempty"`
	UtilizationOfaT           int64  `json:"utilizationOfaT,omitempty"`
	EncoderSamplingPeriodUs   int64  `json:"encoderSamplingPeriodUs"`
	DecoderSamplingPeriodUs   int64  `json:"decoderSamplingPeriodUs"`
	MemoryUsedBytes           int64  `json:"memoryUsedBytes"`
	MemoryUsedBytesT          int64  `json:"memoryUsedBytesT"`
	MemoryFreeBytes           int64  `json:"memoryFreeBytes"`
	MemoryFreeBytesT          int64  `json:"memoryFreeBytesT"`
	MemoryTotalBytes          int64  `json:"memoryTotalBytes"`
	MemoryReservedBytes       int64  `json:"memoryReservedBytes"`
	MemoryReservedBytesT      int64  `json:"memoryReservedBytesT"`
	Bar1UsedBytes             int64  `json:"bar1UsedBytes,omitempty"`
	Bar1UsedBytesT            int64  `json:"bar1UsedBytesT,omitempty"`
	Bar1FreeBytes             int64  `json:"bar1FreeBytes,omitempty"`
	Bar1FreeBytesT            int64  `json:"bar1FreeBytesT,omitempty"`
	Bar1TotalBytes            int64  `json:"bar1TotalBytes,omitempty"`
	TemperatureGpuC           int64  `json:"temperatureGpuC"`
	TemperatureGpuCT          int64  `json:"temperatureGpuCT"`
	TemperatureMemoryC        int64  `json:"temperatureMemoryC"`
	TemperatureMemoryCT       int64  `json:"temperatureMemoryCT"`
	FanSpeedPercent           int64  `json:"fanSpeedPercent,omitempty"`
	FanSpeedPercentT          int64  `json:"fanSpeedPercentT,omitempty"`
	FanSpeedsJSON             string `json:"fanSpeedsJson,omitempty"`
	ClockGraphicsMhz          int64  `json:"clockGraphicsMhz"`
	ClockGraphicsMhzT         int64  `json:"clockGraphicsMhzT"`
	ClockSmMhz                int64  `json:"clockSmMhz"`
	ClockSmMhzT               int64  `json:"clockSmMhzT"`
	ClockMemoryMhz            int64  `json:"clockMemoryMhz"`
	ClockMemoryMhzT           int64  `json:"clockMemoryMhzT"`
	ClockVideoMhz             int64  `json:"clockVideoMhz"`
	ClockVideoMhzT            int64  `json:"clockVideoMhzT"`
	PerformanceState          int    `json:"performanceState"`
	PerformanceStateT         int64  `json:"performanceStateT"`
	PowerUsageMw              int64  `json:"powerUsageMw"`
	PowerUsageMwT             int64  `json:"powerUsageMwT"`
	PowerLimitMw              int64  `json:"powerLimitMw"`
	PowerLimitMwT             int64  `json:"powerLimitMwT"`
	PowerEnforcedLimitMw      int64  `json:"powerEnforcedLimitMw"`
	PowerEnforcedLimitMwT     int64  `json:"powerEnforcedLimitMwT"`
	EnergyConsumptionMj       int64  `json:"energyConsumptionMj"`
	EnergyConsumptionMjT      int64  `json:"energyConsumptionMjT"`
	PcieTxBytesPerSec         int64  `json:"pcieTxBytesPerSec"`
	PcieTxBytesPerSecT        int64  `json:"pcieTxBytesPerSecT"`
	PcieRxBytesPerSec         int64  `json:"pcieRxBytesPerSec"`
	PcieRxBytesPerSecT        int64  `json:"pcieRxBytesPerSecT"`
	PcieCurrentLinkGen        int    `json:"pcieCurrentLinkGen"`
	PcieCurrentLinkGenT       int64  `json:"pcieCurrentLinkGenT"`
	PcieCurrentLinkWidth      int    `json:"pcieCurrentLinkWidth"`
	PcieCurrentLinkWidthT     int64  `json:"pcieCurrentLinkWidthT"`
	PcieReplayCounter         int64  `json:"pcieReplayCounter"`
	PcieReplayCounterT        int64  `json:"pcieReplayCounterT"`
	ClocksEventReasons        uint64 `json:"clocksEventReasons"`
	ClocksEventReasonsT       int64  `json:"clocksEventReasonsT"`
	ViolationPowerNs          int64  `json:"violationPowerNs"`
	ViolationPowerNsT         int64  `json:"violationPowerNsT"`
	ViolationThermalNs        int64  `json:"violationThermalNs"`
	ViolationThermalNsT       int64  `json:"violationThermalNsT"`
	ViolationReliabilityNs    int64  `json:"violationReliabilityNs,omitempty"`
	ViolationReliabilityNsT   int64  `json:"violationReliabilityNsT,omitempty"`
	ViolationBoardLimitNs     int64  `json:"violationBoardLimitNs,omitempty"`
	ViolationBoardLimitNsT    int64  `json:"violationBoardLimitNsT,omitempty"`
	ViolationLowUtilNs        int64  `json:"violationLowUtilNs,omitempty"`
	ViolationLowUtilNsT       int64  `json:"violationLowUtilNsT,omitempty"`
	ViolationSyncBoostNs      int64  `json:"violationSyncBoostNs,omitempty"`
	ViolationSyncBoostNsT     int64  `json:"violationSyncBoostNsT,omitempty"`
	EccAggregateSbe           int64  `json:"eccAggregateSbe"`
	EccAggregateSbeT          int64  `json:"eccAggregateSbeT"`
	EccAggregateDbe           int64  `json:"eccAggregateDbe"`
	EccAggregateDbeT          int64  `json:"eccAggregateDbeT"`
	RetiredPagesSbe           int64  `json:"retiredPagesSbe,omitempty"`
	RetiredPagesDbe           int64  `json:"retiredPagesDbe,omitempty"`
	RetiredPagesT             int64  `json:"retiredPagesT,omitempty"`
	RetiredPending            bool   `json:"retiredPending,omitempty"`
	RetiredPendingT           int64  `json:"retiredPendingT,omitempty"`
	RemappedRowsCorrectable   int64  `json:"remappedRowsCorrectable,omitempty"`
	RemappedRowsUncorrectable int64  `json:"remappedRowsUncorrectable,omitempty"`
	RemappedRowsPending       bool   `json:"remappedRowsPending,omitempty"`
	RemappedRowsFailure       bool   `json:"remappedRowsFailure,omitempty"`
	RemappedRowsT             int64  `json:"remappedRowsT,omitempty"`
	NvLinkBandwidthJSON       string `json:"nvlinkBandwidthJson,omitempty"`
	NvLinkErrorsJSON          string `json:"nvlinkErrorsJson,omitempty"`
	ProcessCount              int64  `json:"processCount"`
	ProcessCountT             int64  `json:"processCountT"`
	ProcessesJSON             string `json:"processesJson,omitempty"`
	ProcessUtilizationJSON    string `json:"processUtilizationJson,omitempty"`
}

// ProcessInfo contains metrics for a single OS process.
type ProcessInfo struct {
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

// VLLMHistograms holds histogram data for vLLM metrics.
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

// GPUInfo contains static information about an NVIDIA GPU.
type GPUInfo struct {
	Index               int    `json:"index"`
	Name                string `json:"name"`
	UUID                string `json:"uuid"`
	Serial              string `json:"serial,omitempty"`
	BoardPartNumber     string `json:"boardPartNumber,omitempty"`
	Brand               string `json:"brand,omitempty"`
	Architecture        string `json:"architecture,omitempty"`
	CudaCapabilityMajor int    `json:"cudaCapabilityMajor,omitempty"`
	CudaCapabilityMinor int    `json:"cudaCapabilityMinor,omitempty"`
	MemoryTotalBytes    int64  `json:"memoryTotalBytes"`
	Bar1TotalBytes      int64  `json:"bar1TotalBytes,omitempty"`
	MemoryBusWidthBits  int    `json:"memoryBusWidthBits,omitempty"`
	NumCores            int    `json:"numCores,omitempty"`
	MaxClockGraphicsMhz int    `json:"maxClockGraphicsMhz,omitempty"`
	MaxClockMemoryMhz   int    `json:"maxClockMemoryMhz,omitempty"`
	MaxClockSmMhz       int    `json:"maxClockSmMhz,omitempty"`
	MaxClockVideoMhz    int    `json:"maxClockVideoMhz,omitempty"`
	PciBusId            string `json:"pciBusId,omitempty"`
	PciDeviceId         uint32 `json:"pciDeviceId,omitempty"`
	PciSubsystemId      uint32 `json:"pciSubsystemId,omitempty"`
	PcieMaxLinkGen      int    `json:"pcieMaxLinkGen,omitempty"`
	PcieMaxLinkWidth    int    `json:"pcieMaxLinkWidth,omitempty"`
	PowerDefaultLimitMw int    `json:"powerDefaultLimitMw,omitempty"`
	PowerMinLimitMw     int    `json:"powerMinLimitMw,omitempty"`
	PowerMaxLimitMw     int    `json:"powerMaxLimitMw,omitempty"`
	VbiosVersion        string `json:"vbiosVersion,omitempty"`
	InforomImageVersion string `json:"inforomImageVersion,omitempty"`
	InforomOemVersion   string `json:"inforomOemVersion,omitempty"`
	NumFans             int    `json:"numFans,omitempty"`
	TempShutdownC       int    `json:"tempShutdownC,omitempty"`
	TempSlowdownC       int    `json:"tempSlowdownC,omitempty"`
	TempMaxOperatingC   int    `json:"tempMaxOperatingC,omitempty"`
	TempTargetC         int    `json:"tempTargetC,omitempty"`
	EccModeEnabled      bool   `json:"eccModeEnabled,omitempty"`
	PersistenceModeOn   bool   `json:"persistenceModeOn,omitempty"`
	ComputeMode         string `json:"computeMode,omitempty"`
	IsMultiGpuBoard     bool   `json:"isMultiGpuBoard,omitempty"`
	DisplayModeEnabled  bool   `json:"displayModeEnabled,omitempty"`
	DisplayActive       bool   `json:"displayActive,omitempty"`
	MigModeEnabled      bool   `json:"migModeEnabled,omitempty"`
	EncoderCapacityH264 int    `json:"encoderCapacityH264,omitempty"`
	EncoderCapacityHEVC int    `json:"encoderCapacityHEVC,omitempty"`
	EncoderCapacityAV1  int    `json:"encoderCapacityAV1,omitempty"`
	NvLinkCount         int    `json:"nvlinkCount,omitempty"`
}

// GPUProcess contains information about a process using the GPU.
type GPUProcess struct {
	PID             uint32 `json:"pid"`
	Name            string `json:"name"`
	UsedMemoryBytes int64  `json:"usedMemoryBytes"`
	Type            string `json:"type,omitempty"`
}

// NvLinkBandwidth contains NVLink bandwidth metrics.
type NvLinkBandwidth struct {
	Link    int   `json:"link"`
	TxBytes int64 `json:"txBytes"`
	RxBytes int64 `json:"rxBytes"`
}

// NvLinkErrors contains NVLink error metrics.
type NvLinkErrors struct {
	Link          int   `json:"link"`
	CrcErrors     int64 `json:"crcErrors"`
	EccErrors     int64 `json:"eccErrors"`
	ReplayErrors  int64 `json:"replayErrors"`
	RecoveryCount int64 `json:"recoveryCount"`
}

// GPUProcessUtilization contains per-process GPU utilization data.
type GPUProcessUtilization struct {
	PID         uint32 `json:"pid"`
	SmUtil      int    `json:"smUtil"`
	MemUtil     int    `json:"memUtil"`
	EncUtil     int    `json:"encUtil"`
	DecUtil     int    `json:"decUtil"`
	TimestampUs int64  `json:"timestampUs"`
}

// Collector interface for all metric collectors.
type Collector interface {
	Name() string
	CollectStatic(s *StaticMetrics)
	CollectDynamic(d *DynamicMetrics)
	Close() error
}
