// Package types provides shared metric types for collectors.
package collectors

// =============================================================================
// Base Types
// =============================================================================

// BaseStatic contains session identification.
type BaseStatic struct {
	UUID     string `json:"uuid"`
	VMID     string `json:"vId"`
	Hostname string `json:"vHostname"`
	BootTime int64  `json:"vBootTime"`
}

// BaseDynamic contains the collection timestamp.
type BaseDynamic struct {
	Timestamp int64 `json:"timestamp"`
}

// =============================================================================
// CPU Types
// =============================================================================

// CPUStatic contains static CPU information.
type CPUStatic struct {
	NumProcessors       int     `json:"vNumProcessors"`
	CPUType             string  `json:"vCpuType"`
	CPUCache            string  `json:"vCpuCache"`
	KernelInfo          string  `json:"vKernelInfo"`
	TimeSynced          bool    `json:"vTimeSynced"`
	TimeOffsetSeconds   float64 `json:"vTimeOffsetSeconds"`
	TimeMaxErrorSeconds float64 `json:"vTimeMaxErrorSeconds"`
}

// CPUDynamic contains dynamic CPU metrics.
type CPUDynamic struct {
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
}

// =============================================================================
// Memory Types
// =============================================================================

// MemoryStatic contains static memory information.
type MemoryStatic struct {
	MemoryTotalBytes int64 `json:"vMemoryTotalBytes"`
	SwapTotalBytes   int64 `json:"vSwapTotalBytes"`
}

// MemoryDynamic contains dynamic memory metrics.
type MemoryDynamic struct {
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
}

// =============================================================================
// Disk Types
// =============================================================================

// DiskInfo contains static info for a single disk.
type DiskInfo struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Vendor    string `json:"vendor"`
	SizeBytes int64  `json:"sizeBytes"`
}

// DiskStatic contains static disk information.
type DiskStatic struct {
	DisksJSON string `json:"disks,omitempty"`
}

// DiskDynamic contains dynamic disk metrics.
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

// =============================================================================
// Network Types
// =============================================================================

// NetworkInterface contains static info for a network interface.
type NetworkInterface struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	State     string `json:"state"`
	MTU       int64  `json:"mtu"`
	SpeedMbps int64  `json:"speedMbps,omitempty"`
}

// NetworkStatic contains static network information.
type NetworkStatic struct {
	NetworkInterfacesJSON string `json:"networkInterfaces,omitempty"`
}

// NetworkDynamic contains dynamic network metrics.
type NetworkDynamic struct {
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
}

// =============================================================================
// Container Types
// =============================================================================

// ContainerStatic contains static container information.
type ContainerStatic struct {
	ContainerID      string `json:"cId"`
	ContainerNumCPUs int64  `json:"cNumProcessors"`
	CgroupVersion    int64  `json:"cCgroupVersion"`
}

// ContainerDynamic contains dynamic container metrics.
type ContainerDynamic struct {
	ContainerCPUTime            int64  `json:"cCpuTime"`
	ContainerCPUTimeT           int64  `json:"cCpuTimeT"`
	ContainerCPUTimeUserMode    int64  `json:"cCpuTimeUserMode"`
	ContainerCPUTimeUserModeT   int64  `json:"cCpuTimeUserModeT"`
	ContainerCPUTimeKernelMode  int64  `json:"cCpuTimeKernelMode"`
	ContainerCPUTimeKernelModeT int64  `json:"cCpuTimeKernelModeT"`
	ContainerPerCPUTimesJSON    string `json:"cCpuPerCpuJson,omitempty"`
	ContainerPerCPUTimesT       int64  `json:"cCpuPerCpuT,omitempty"`
	ContainerMemoryUsed         int64  `json:"cMemoryUsed"`
	ContainerMemoryUsedT        int64  `json:"cMemoryUsedT"`
	ContainerMemoryMaxUsed      int64  `json:"cMemoryMaxUsed"`
	ContainerMemoryMaxUsedT     int64  `json:"cMemoryMaxUsedT"`
	ContainerPgFault            int64  `json:"cPgFault"`
	ContainerPgFaultT           int64  `json:"cPgFaultT"`
	ContainerMajorPgFault       int64  `json:"cMajorPgFault"`
	ContainerMajorPgFaultT      int64  `json:"cMajorPgFaultT"`
	ContainerDiskReadBytes      int64  `json:"cDiskReadBytes"`
	ContainerDiskReadBytesT     int64  `json:"cDiskReadBytesT"`
	ContainerDiskWriteBytes     int64  `json:"cDiskWriteBytes"`
	ContainerDiskWriteBytesT    int64  `json:"cDiskWriteBytesT"`
	ContainerDiskSectorIO       int64  `json:"cDiskSectorIO,omitempty"`
	ContainerDiskSectorIOT      int64  `json:"cDiskSectorIOT,omitempty"`
	ContainerNetworkBytesRecvd  int64  `json:"cNetworkBytesRecvd"`
	ContainerNetworkBytesRecvdT int64  `json:"cNetworkBytesRecvdT"`
	ContainerNetworkBytesSent   int64  `json:"cNetworkBytesSent"`
	ContainerNetworkBytesSentT  int64  `json:"cNetworkBytesSentT"`
	ContainerNumProcesses       int64  `json:"cNumProcesses"`
	ContainerNumProcessesT      int64  `json:"cNumProcessesT"`
}

// =============================================================================
// NVIDIA GPU Types
// =============================================================================

// NvidiaStatic contains static NVIDIA driver information.
type NvidiaStatic struct {
	NvidiaDriverVersion string `json:"nvidiaDriverVersion,omitempty"`
	NvidiaCudaVersion   string `json:"nvidiaCudaVersion,omitempty"`
	NvmlVersion         string `json:"nvmlVersion,omitempty"`
	NvidiaGPUCount      int    `json:"nvidiaGpuCount,omitempty"`
	NvidiaGPUsJSON      string `json:"nvidiaGpus,omitempty"`
}

// GPUInfo contains static information for a single GPU.
type GPUInfo struct {
	Index               int    `json:"index"`
	Name                string `json:"name"`
	UUID                string `json:"uuid"`
	Serial              string `json:"serial,omitempty"`
	BoardPartNumber     string `json:"boardPartNumber,omitempty"`
	Brand               string `json:"brand,omitempty"`
	Architecture        string `json:"architecture,omitempty"`
	VbiosVersion        string `json:"vbiosVersion,omitempty"`
	InforomImageVersion string `json:"inforomImageVersion,omitempty"`
	InforomOemVersion   string `json:"inforomOemVersion,omitempty"`
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
	EncoderCapacityHEVC int    `json:"encoderCapacityHevc,omitempty"`
	EncoderCapacityAV1  int    `json:"encoderCapacityAv1,omitempty"`
	NvLinkCount         int    `json:"nvlinkCount,omitempty"`
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
	EncoderSamplingPeriodUs   int64  `json:"encoderSamplingPeriodUs,omitempty"`
	DecoderSamplingPeriodUs   int64  `json:"decoderSamplingPeriodUs,omitempty"`
	MemoryUsedBytes           int64  `json:"memoryUsedBytes"`
	MemoryUsedBytesT          int64  `json:"memoryUsedBytesT"`
	MemoryFreeBytes           int64  `json:"memoryFreeBytes"`
	MemoryFreeBytesT          int64  `json:"memoryFreeBytesT"`
	MemoryTotalBytes          int64  `json:"memoryTotalBytes"`
	MemoryReservedBytes       int64  `json:"memoryReservedBytes,omitempty"`
	MemoryReservedBytesT      int64  `json:"memoryReservedBytesT,omitempty"`
	Bar1UsedBytes             int64  `json:"bar1UsedBytes"`
	Bar1UsedBytesT            int64  `json:"bar1UsedBytesT"`
	Bar1FreeBytes             int64  `json:"bar1FreeBytes"`
	Bar1FreeBytesT            int64  `json:"bar1FreeBytesT"`
	Bar1TotalBytes            int64  `json:"bar1TotalBytes"`
	TemperatureGpuC           int64  `json:"temperatureGpuC"`
	TemperatureGpuCT          int64  `json:"temperatureGpuCT"`
	TemperatureMemoryC        int64  `json:"temperatureMemoryC,omitempty"`
	TemperatureMemoryCT       int64  `json:"temperatureMemoryCT,omitempty"`
	FanSpeedPercent           int64  `json:"fanSpeedPercent"`
	FanSpeedPercentT          int64  `json:"fanSpeedPercentT"`
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

// GPUProcess contains information about a process using the GPU.
type GPUProcess struct {
	PID             uint32 `json:"pid"`
	Name            string `json:"name"`
	UsedMemoryBytes int64  `json:"usedMemoryBytes"`
	Type            string `json:"type,omitempty"`
}

// GPUProcessUtilization contains utilization info for a GPU process.
type GPUProcessUtilization struct {
	PID         uint32 `json:"pid"`
	SmUtil      int    `json:"smUtil"`
	MemUtil     int    `json:"memUtil"`
	EncUtil     int    `json:"encUtil"`
	DecUtil     int    `json:"decUtil"`
	TimestampUs int64  `json:"timestampUs"`
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

// =============================================================================
// vLLM Types
// =============================================================================

// VLLMDynamic contains dynamic vLLM metrics.
type VLLMDynamic struct {
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
}

// VLLMHistograms contains vLLM histogram data.
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

// =============================================================================
// Process Types
// =============================================================================

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

// =============================================================================
// Composite Types (the big structs)
// =============================================================================

// StaticMetrics contains all static system information collected once at startup.
type StaticMetrics struct {
	BaseStatic
	CPUStatic
	MemoryStatic
	DiskStatic
	NetworkStatic
	ContainerStatic
	NvidiaStatic
}

// DynamicMetrics contains all dynamic metrics collected during profiling.
type DynamicMetrics struct {
	BaseDynamic
	CPUDynamic
	MemoryDynamic
	DiskDynamic
	NetworkDynamic
	ContainerDynamic
	VLLMDynamic

	// Slice data (handled separately during export)
	GPUs      []GPUDynamic  `json:"-"`
	Processes []ProcessInfo `json:"-"`
}
