package nvidia

import "InferenceProfiler/pkg/collectors/types"

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

// Static contains static GPU information.
type Static struct {
	NvidiaDriverVersion string
	NvidiaCudaVersion   string
	NvmlVersion         string
	NvidiaGPUCount      int
	NvidiaGPUsJSON      string
}

// ToRecord converts Static to a Record.
func (s *Static) ToRecord() types.Record {
	r := types.Record{
		"nvidiaGpuCount": s.NvidiaGPUCount,
	}
	if s.NvidiaDriverVersion != "" {
		r["nvidiaDriverVersion"] = s.NvidiaDriverVersion
	}
	if s.NvidiaCudaVersion != "" {
		r["nvidiaCudaVersion"] = s.NvidiaCudaVersion
	}
	if s.NvmlVersion != "" {
		r["nvmlVersion"] = s.NvmlVersion
	}
	if s.NvidiaGPUsJSON != "" {
		r["nvidiaGpus"] = s.NvidiaGPUsJSON
	}
	return r
}

// GPUDynamicMetrics contains dynamic metrics for a single GPU.
type GPUDynamicMetrics struct {
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

// Dynamic contains dynamic GPU metrics.
type Dynamic struct {
	NvidiaGPUsJSON string
}

// ToRecord converts Dynamic to a Record.
func (d *Dynamic) ToRecord() types.Record {
	r := types.Record{}
	if d.NvidiaGPUsJSON != "" {
		r["nvidiaGpusDynamic"] = d.NvidiaGPUsJSON
	}
	return r
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
