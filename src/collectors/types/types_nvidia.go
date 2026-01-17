package types

// NvidiaGPUDynamic contains dynamic metrics for a single NVIDIA GPU
type NvidiaGPUDynamic struct {
	Index int `json:"index"`

	// =========================================================================
	// Utilization (0-100 percent)
	// =========================================================================

	UtilizationGPU          int64 `json:"utilizationGpu"`
	UtilizationGPUT         int64 `json:"utilizationGpuT"`
	UtilizationMemory       int64 `json:"utilizationMemory"`
	UtilizationMemoryT      int64 `json:"utilizationMemoryT"`
	UtilizationEncoder      int64 `json:"utilizationEncoder"`
	UtilizationEncoderT     int64 `json:"utilizationEncoderT"`
	UtilizationDecoder      int64 `json:"utilizationDecoder"`
	UtilizationDecoderT     int64 `json:"utilizationDecoderT"`
	UtilizationJpeg         int64 `json:"utilizationJpeg,omitempty"` // Turing+
	UtilizationJpegT        int64 `json:"utilizationJpegT,omitempty"`
	UtilizationOfa          int64 `json:"utilizationOfa,omitempty"` // Optical Flow, Turing+
	UtilizationOfaT         int64 `json:"utilizationOfaT,omitempty"`
	EncoderSamplingPeriodUs int64 `json:"encoderSamplingPeriodUs,omitempty"`
	DecoderSamplingPeriodUs int64 `json:"decoderSamplingPeriodUs,omitempty"`

	// =========================================================================
	// Memory (bytes)
	// =========================================================================

	MemoryUsedBytes      int64 `json:"memoryUsedBytes"`
	MemoryUsedBytesT     int64 `json:"memoryUsedBytesT"`
	MemoryFreeBytes      int64 `json:"memoryFreeBytes"`
	MemoryFreeBytesT     int64 `json:"memoryFreeBytesT"`
	MemoryTotalBytes     int64 `json:"memoryTotalBytes"`
	MemoryReservedBytes  int64 `json:"memoryReservedBytes,omitempty"`
	MemoryReservedBytesT int64 `json:"memoryReservedBytesT,omitempty"`

	// BAR1 memory (CPU-accessible GPU memory window)
	Bar1UsedBytes  int64 `json:"bar1UsedBytes"`
	Bar1UsedBytesT int64 `json:"bar1UsedBytesT"`
	Bar1FreeBytes  int64 `json:"bar1FreeBytes"`
	Bar1FreeBytesT int64 `json:"bar1FreeBytesT"`
	Bar1TotalBytes int64 `json:"bar1TotalBytes"`

	// =========================================================================
	// Temperature (Celsius)
	// =========================================================================

	TemperatureGpuC     int64 `json:"temperatureGpuC"`
	TemperatureGpuCT    int64 `json:"temperatureGpuCT"`
	TemperatureMemoryC  int64 `json:"temperatureMemoryC,omitempty"` // HBM, Ampere+
	TemperatureMemoryCT int64 `json:"temperatureMemoryCT,omitempty"`

	// =========================================================================
	// Fan
	// =========================================================================

	FanSpeedPercent  int64  `json:"fanSpeedPercent"`
	FanSpeedPercentT int64  `json:"fanSpeedPercentT"`
	FanSpeedsJSON    string `json:"fanSpeedsJson,omitempty"` // Per-fan speeds array

	// =========================================================================
	// Clock Speeds (MHz)
	// =========================================================================

	ClockGraphicsMhz    int64 `json:"clockGraphicsMhz"`
	ClockGraphicsMhzT   int64 `json:"clockGraphicsMhzT"`
	ClockSmMhz          int64 `json:"clockSmMhz"`
	ClockSmMhzT         int64 `json:"clockSmMhzT"`
	ClockMemoryMhz      int64 `json:"clockMemoryMhz"`
	ClockMemoryMhzT     int64 `json:"clockMemoryMhzT"`
	ClockVideoMhz       int64 `json:"clockVideoMhz"`
	ClockVideoMhzT      int64 `json:"clockVideoMhzT"`
	AppClockGraphicsMhz int64 `json:"appClockGraphicsMhz,omitempty"` // User-configured
	AppClockMemoryMhz   int64 `json:"appClockMemoryMhz,omitempty"`
	AppClocksT          int64 `json:"appClocksT,omitempty"`

	// =========================================================================
	// Performance State
	// =========================================================================

	// PerformanceState is the P-state (P0=highest performance, P15=lowest)
	PerformanceState  int   `json:"performanceState"`
	PerformanceStateT int64 `json:"performanceStateT"`

	// =========================================================================
	// Power (milliwatts)
	// =========================================================================

	PowerUsageMw          int64 `json:"powerUsageMw"`
	PowerUsageMwT         int64 `json:"powerUsageMwT"`
	PowerLimitMw          int64 `json:"powerLimitMw"`
	PowerLimitMwT         int64 `json:"powerLimitMwT"`
	PowerEnforcedLimitMw  int64 `json:"powerEnforcedLimitMw"`
	PowerEnforcedLimitMwT int64 `json:"powerEnforcedLimitMwT"`

	// EnergyConsumptionMj is cumulative energy since driver load (millijoules)
	EnergyConsumptionMj  int64 `json:"energyConsumptionMj"`
	EnergyConsumptionMjT int64 `json:"energyConsumptionMjT"`

	// =========================================================================
	// PCIe
	// =========================================================================

	PcieTxBytesPerSec     int64 `json:"pcieTxBytesPerSec"` // bytes/sec (converted from KB/s)
	PcieTxBytesPerSecT    int64 `json:"pcieTxBytesPerSecT"`
	PcieRxBytesPerSec     int64 `json:"pcieRxBytesPerSec"`
	PcieRxBytesPerSecT    int64 `json:"pcieRxBytesPerSecT"`
	PcieCurrentLinkGen    int   `json:"pcieCurrentLinkGen"`
	PcieCurrentLinkGenT   int64 `json:"pcieCurrentLinkGenT"`
	PcieCurrentLinkWidth  int   `json:"pcieCurrentLinkWidth"`
	PcieCurrentLinkWidthT int64 `json:"pcieCurrentLinkWidthT"`
	PcieReplayCounter     int64 `json:"pcieReplayCounter"` // Error count
	PcieReplayCounterT    int64 `json:"pcieReplayCounterT"`

	// =========================================================================
	// Throttling / Clock Event Reasons
	// =========================================================================

	ClocksEventReasons    uint64   `json:"clocksEventReasons"` // Raw bitmask
	ClocksEventReasonsT   int64    `json:"clocksEventReasonsT"`
	ThrottleReasonsActive []string `json:"throttleReasonsActive,omitempty"` // Decoded names

	// Violation times (nanoseconds spent in throttled state, cumulative)
	ViolationPowerNs        int64 `json:"violationPowerNs"`
	ViolationPowerNsT       int64 `json:"violationPowerNsT"`
	ViolationThermalNs      int64 `json:"violationThermalNs"`
	ViolationThermalNsT     int64 `json:"violationThermalNsT"`
	ViolationReliabilityNs  int64 `json:"violationReliabilityNs,omitempty"`
	ViolationReliabilityNsT int64 `json:"violationReliabilityNsT,omitempty"`
	ViolationBoardLimitNs   int64 `json:"violationBoardLimitNs,omitempty"`
	ViolationBoardLimitNsT  int64 `json:"violationBoardLimitNsT,omitempty"`
	ViolationLowUtilNs      int64 `json:"violationLowUtilNs,omitempty"`
	ViolationLowUtilNsT     int64 `json:"violationLowUtilNsT,omitempty"`
	ViolationSyncBoostNs    int64 `json:"violationSyncBoostNs,omitempty"`
	ViolationSyncBoostNsT   int64 `json:"violationSyncBoostNsT,omitempty"`

	// =========================================================================
	// ECC Errors
	// =========================================================================

	// Volatile errors (since last driver load)
	EccVolatileSbe  int64 `json:"eccVolatileSbe"` // Single-bit errors (corrected)
	EccVolatileSbeT int64 `json:"eccVolatileSbeT"`
	EccVolatileDbe  int64 `json:"eccVolatileDbe"` // Double-bit errors (uncorrected)
	EccVolatileDbeT int64 `json:"eccVolatileDbeT"`

	// Aggregate errors (lifetime)
	EccAggregateSbe  int64 `json:"eccAggregateSbe"`
	EccAggregateSbeT int64 `json:"eccAggregateSbeT"`
	EccAggregateDbe  int64 `json:"eccAggregateDbe"`
	EccAggregateDbeT int64 `json:"eccAggregateDbeT"`

	// Memory page retirement
	RetiredPagesSbe int64 `json:"retiredPagesSbe,omitempty"`
	RetiredPagesDbe int64 `json:"retiredPagesDbe,omitempty"`
	RetiredPagesT   int64 `json:"retiredPagesT,omitempty"`
	RetiredPending  bool  `json:"retiredPending,omitempty"`
	RetiredPendingT int64 `json:"retiredPendingT,omitempty"`

	// Remapped rows (Ampere+ HBM health)
	RemappedRowsCorrectable   int64 `json:"remappedRowsCorrectable,omitempty"`
	RemappedRowsUncorrectable int64 `json:"remappedRowsUncorrectable,omitempty"`
	RemappedRowsPending       bool  `json:"remappedRowsPending,omitempty"`
	RemappedRowsFailure       bool  `json:"remappedRowsFailure,omitempty"`
	RemappedRowsT             int64 `json:"remappedRowsT,omitempty"`

	// =========================================================================
	// Encoder/Decoder Session Statistics
	// =========================================================================

	EncoderSessionCount int   `json:"encoderSessionCount"`
	EncoderAvgFps       int   `json:"encoderAvgFps"`
	EncoderAvgLatencyUs int   `json:"encoderAvgLatencyUs"`
	EncoderStatsT       int64 `json:"encoderStatsT"`

	FbcSessionCount int   `json:"fbcSessionCount"`
	FbcAvgFps       int   `json:"fbcAvgFps"`
	FbcAvgLatencyUs int   `json:"fbcAvgLatencyUs"`
	FbcStatsT       int64 `json:"fbcStatsT"`

	// =========================================================================
	// NVLink (Multi-GPU)
	// =========================================================================

	NvLinkBandwidthJSON string `json:"nvlinkBandwidthJson,omitempty"`
	NvLinkErrorsJSON    string `json:"nvlinkErrorsJson,omitempty"`

	// =========================================================================
	// Running Processes
	// =========================================================================

	ProcessCount           int64  `json:"processCount"`
	ProcessCountT          int64  `json:"processCountT"`
	ProcessesJSON          string `json:"processesJson,omitempty"`
	ProcessUtilizationJSON string `json:"processUtilizationJson,omitempty"`
}

// GPUProcess represents a process using GPU resources
type GPUProcess struct {
	PID             uint32 `json:"pid"`
	Name            string `json:"name"`
	UsedMemoryBytes int64  `json:"usedMemoryBytes"`
	Type            string `json:"type,omitempty"` // "compute", "graphics", "mps"
}

// GPUProcessUtilization represents per-process GPU utilization sample
type GPUProcessUtilization struct {
	PID         uint32 `json:"pid"`
	SmUtil      int    `json:"smUtil"`
	MemUtil     int    `json:"memUtil"`
	EncUtil     int    `json:"encUtil"`
	DecUtil     int    `json:"decUtil"`
	TimestampUs int64  `json:"timestampUs"`
}

// NvLinkBandwidth represents NVLink throughput for a single link
type NvLinkBandwidth struct {
	Link         int   `json:"link"`
	TxBytes      int64 `json:"txBytes"`
	RxBytes      int64 `json:"rxBytes"`
	ThroughputTx int64 `json:"throughputTx"` // bytes/sec
	ThroughputRx int64 `json:"throughputRx"` // bytes/sec
}

// NvLinkErrors represents error counts for a single NVLink
type NvLinkErrors struct {
	Link          int   `json:"link"`
	CrcErrors     int64 `json:"crcErrors"`
	EccErrors     int64 `json:"eccErrors"`
	ReplayErrors  int64 `json:"replayErrors"`
	RecoveryCount int64 `json:"recoveryCount"`
}
