package collectors

import "github.com/google/uuid"

type Timed[T any] struct {
	Value T     `json:"value"`
	Time  int64 `json:"time"`
}
type Snapshot struct {
	Timestamp int64              `json:"timestamp"`
	CPU       CPUMetrics         `json:"cpu"`
	Memory    MemoryMetrics      `json:"mem"`
	Disk      DiskMetrics        `json:"disk"`
	Network   NetMetrics         `json:"net"`
	Container ContainerMetrics   `json:"containers"`
	NVIDIA    []NvidiaMetrics    `json:"nvidia,omitempty"`
	VLLM      VLLMMetrics        `json:"vllm,omitempty"`
	Processes *ProcessCollection `json:"processes,omitempty"`
}
type StaticInfo struct {
	UUID   uuid.UUID     `json:"uuid"`
	VMID   string        `json:"vId"`
	Host   HostInfo      `json:"host"`
	NVIDIA *NvidiaStatic `json:"nvidia,omitempty"`
}
type HostInfo struct {
	Hostname      string           `json:"vHostname"`
	Kernel        string           `json:"vKernel"`
	BootTime      int64            `json:"vBootTime"`
	NumProcessors int              `json:"vNumProcessors"`
	CPUType       string           `json:"vCpuType"`
	CPUCache      map[string]int64 `json:"vCpuCache"`
	KernelInfo    string           `json:"vKernelInfo"`
	MemoryTotal   int64            `json:"vMemoryTotalBytes"`
	SwapTotal     int64            `json:"vSwapTotalBytes"`
}
type CPUMetrics struct {
	Time            Timed[int64]   `json:"vCpuTime"`
	UserMode        Timed[int64]   `json:"vCpuTimeUserMode"`
	KernelMode      Timed[int64]   `json:"vCpuTimeKernelMode"`
	Idle            Timed[int64]   `json:"vCpuIdleTime"`
	IOWait          Timed[int64]   `json:"vCpuTimeIOWait"`
	IRQ             Timed[int64]   `json:"vCpuTimeIntSrvc"`
	SoftIRQ         Timed[int64]   `json:"vCpuTimeSoftIntSrvc"`
	Nice            Timed[int64]   `json:"vCpuNice"`
	Steal           Timed[int64]   `json:"vCpuSteal"`
	ContextSwitches Timed[int64]   `json:"vCpuContextSwitches"`
	LoadAvg         Timed[float64] `json:"vLoadAvg"`
	FreqMHz         Timed[float64] `json:"vCpuMhz"`
}
type CPUStatic struct {
	NumProcessors int              `json:"vNumProcessors"`
	Model         string           `json:"vCpuType"`
	Cache         map[string]int64 `json:"vCpuCache"`
	KernelInfo    string           `json:"vKernelInfo"`
}
type MemoryMetrics struct {
	Total          Timed[int64]   `json:"vMemoryTotal"`
	Free           Timed[int64]   `json:"vMemoryFree"`
	Used           Timed[int64]   `json:"vMemoryUsed"`
	Buffers        Timed[int64]   `json:"vMemoryBuffers"`
	Cached         Timed[int64]   `json:"vMemoryCached"`
	Percent        Timed[float64] `json:"vMemoryPercent"`
	SwapTotal      Timed[int64]   `json:"vSwapTotal"`
	SwapFree       Timed[int64]   `json:"vSwapFree"`
	SwapUsed       Timed[int64]   `json:"vSwapUsed"`
	PageFault      Timed[int64]   `json:"vPgFault"`
	MajorPageFault Timed[int64]   `json:"vMajorPageFault"`
}
type MemoryStatic struct {
	TotalBytes     int64 `json:"vMemoryTotalBytes"`
	SwapTotalBytes int64 `json:"vSwapTotalBytes"`
}
type DiskMetrics struct {
	SectorReads      Timed[int64] `json:"vDiskSectorReads"`
	SectorWrites     Timed[int64] `json:"vDiskSectorWrites"`
	ReadBytes        Timed[int64] `json:"vDiskReadBytes"`
	WriteBytes       Timed[int64] `json:"vDiskWriteBytes"`
	SuccessfulReads  Timed[int64] `json:"vDiskSuccessfulReads"`
	SuccessfulWrites Timed[int64] `json:"vDiskSuccessfulWrites"`
	MergedReads      Timed[int64] `json:"vDiskMergedReads"`
	MergedWrites     Timed[int64] `json:"vDiskMergedWrites"`
	ReadTime         Timed[int64] `json:"vDiskReadTime"`
	WriteTime        Timed[int64] `json:"vDiskWriteTime"`
	IOTime           Timed[int64] `json:"vDiskIOTime"`
	WeightedIOTime   Timed[int64] `json:"vDiskWeightedIOTime"`
	IOInProgress     Timed[int64] `json:"vDiskIOInProgress"`
}
type NetMetrics struct {
	BytesRecvd   Timed[int64] `json:"vNetworkBytesRecvd"`
	BytesSent    Timed[int64] `json:"vNetworkBytesSent"`
	PacketsRecvd Timed[int64] `json:"vNetworkPacketsRecvd"`
	PacketsSent  Timed[int64] `json:"vNetworkPacketsSent"`
	ErrorsRecvd  Timed[int64] `json:"vNetworkErrorsRecvd"`
	ErrorsSent   Timed[int64] `json:"vNetworkErrorsSent"`
	DropsRecvd   Timed[int64] `json:"vNetworkDropsRecvd"`
	DropsSent    Timed[int64] `json:"vNetworkDropsSent"`
}
type ProcessMetrics struct {
	PID                     int          `json:"pId"`
	Name                    string       `json:"pName"`
	Cmdline                 string       `json:"pCmdline"`
	NumThreads              Timed[int64] `json:"pNumThreads"`
	CPUUserMode             Timed[int64] `json:"pCpuTimeUserMode"`
	CPUKernelMode           Timed[int64] `json:"pCpuTimeKernelMode"`
	ChildrenUserMode        Timed[int64] `json:"pChildrenUserMode"`
	ChildrenKernelMode      Timed[int64] `json:"pChildrenKernelMode"`
	VoluntaryCtxSwitches    Timed[int64] `json:"pVoluntaryContextSwitches"`
	NonvoluntaryCtxSwitches Timed[int64] `json:"pNonvoluntaryContextSwitches"`
	BlockIODelays           Timed[int64] `json:"pBlockIODelays"`
	VirtualMemory           Timed[int64] `json:"pVirtualMemoryBytes"`
	ResidentSetSize         Timed[int64] `json:"pResidentSetSize"`
}
type ContainerMetrics struct {
	ID             string                  `json:"cId"`
	CgroupVersion  int                     `json:"cCgroupVersion"`
	CPUTime        Timed[int64]            `json:"cCpuTime"`
	CPUUserMode    Timed[int64]            `json:"cCpuTimeUserMode"`
	CPUKernelMode  Timed[int64]            `json:"cCpuTimeKernelMode"`
	NumProcessors  int                     `json:"cNumProcessors"`
	PerCPU         map[string]Timed[int64] `json:"perCpu,omitempty"`
	MemoryUsed     Timed[int64]            `json:"cMemoryUsed"`
	MemoryMaxUsed  Timed[int64]            `json:"cMemoryMaxUsed"`
	DiskReadBytes  Timed[int64]            `json:"cDiskReadBytes"`
	DiskWriteBytes Timed[int64]            `json:"cDiskWriteBytes"`
	NetBytesRecvd  Timed[int64]            `json:"cNetworkBytesRecvd"`
	NetBytesSent   Timed[int64]            `json:"cNetworkBytesSent"`
}
type NvidiaMetrics struct {
	Index            int            `json:"gGpuIndex"`
	UtilizationGPU   Timed[int]     `json:"gUtilizationGpu"`
	UtilizationMem   Timed[int]     `json:"gUtilizationMem"`
	MemoryTotalMB    Timed[int64]   `json:"gMemoryTotalMb"`
	MemoryUsedMB     Timed[int64]   `json:"gMemoryUsedMb"`
	MemoryFreeMB     Timed[int64]   `json:"gMemoryFreeMb"`
	BAR1UsedMB       Timed[int64]   `json:"gBar1UsedMb"`
	BAR1FreeMB       Timed[int64]   `json:"gBar1FreeMb"`
	TemperatureC     Timed[int]     `json:"gTemperatureC"`
	FanSpeed         Timed[int]     `json:"gFanSpeed"`
	PowerDrawW       Timed[float64] `json:"gPowerDrawW"`
	PowerLimitW      Timed[float64] `json:"gPowerLimitW"`
	ClockGraphicsMHz Timed[int]     `json:"gClockGraphicsMhz"`
	ClockSMMHz       Timed[int]     `json:"gClockSmMhz"`
	ClockMemMHz      Timed[int]     `json:"gClockMemMhz"`
	PCIeTxKBps       Timed[int]     `json:"gPcieTxKbps"`
	PCIeRxKBps       Timed[int]     `json:"gPcieRxKbps"`
	PerfState        Timed[string]  `json:"gPerfState"`
	ECCSingleBit     Timed[int64]   `json:"gEccSingleBitErrors"`
	ECCDoubleBit     Timed[int64]   `json:"gEccDoubleBitErrors"`
	ProcessCount     int            `json:"gProcessCount"`
	Processes        []string       `json:"gProcesses"`
}
type NvidiaStatic struct {
	DriverVersion string             `json:"gDriverVersion"`
	CUDAVersion   int                `json:"gCudaVersion"`
	GPUs          []NvidiaStaticInfo `json:"gpus"`
}
type NvidiaStaticInfo struct {
	Name             string `json:"gName"`
	UUID             string `json:"gUuid"`
	TotalMemoryMB    int64  `json:"gTotalMemoryMb"`
	PCIBusID         string `json:"gPciBusId"`
	MaxGraphicsClock int    `json:"gMaxGraphicsClock"`
	MaxSMClock       int    `json:"gMaxSmClock"`
	MaxMemClock      int    `json:"gMaxMemClock"`
}
type VLLMMetrics struct {
	Timestamp              int64                         `json:"timestamp,omitempty"`
	RequestsRunning        float64                       `json:"system_requests_running,omitempty"`
	RequestsWaiting        float64                       `json:"system_requests_waiting,omitempty"`
	EngineSleepState       float64                       `json:"system_engine_sleep_state,omitempty"`
	PreemptionsTotal       float64                       `json:"system_preemptions_total,omitempty"`
	KVCacheUsagePercent    float64                       `json:"cache_kv_usage_percent,omitempty"`
	PrefixCacheHits        float64                       `json:"cache_prefix_hits,omitempty"`
	PrefixCacheQueries     float64                       `json:"cache_prefix_queries,omitempty"`
	MultimodalCacheHits    float64                       `json:"cache_multimodal_hits,omitempty"`
	MultimodalCacheQueries float64                       `json:"cache_multimodal_queries,omitempty"`
	RequestsFinished       float64                       `json:"requests_finished_total,omitempty"`
	RequestsCorrupted      float64                       `json:"requests_corrupted_total,omitempty"`
	PromptTokens           float64                       `json:"tokens_prompt_total,omitempty"`
	GenerationTokens       float64                       `json:"tokens_generation_total,omitempty"`
	LatencyTTFTSum         float64                       `json:"latency_ttft_s_sum,omitempty"`
	LatencyE2ESum          float64                       `json:"latency_e2e_s_sum,omitempty"`
	LatencyQueueSum        float64                       `json:"latency_queue_s_sum,omitempty"`
	LatencyInferenceSum    float64                       `json:"latency_inference_s_sum,omitempty"`
	LatencyPrefillSum      float64                       `json:"latency_prefill_s_sum,omitempty"`
	LatencyDecodeSum       float64                       `json:"latency_decode_s_sum,omitempty"`
	LatencyInterTokenSum   float64                       `json:"latency_inter_token_s_sum,omitempty"`
	Histograms             map[string]map[string]float64 `json:"histograms,omitempty"`
	Config                 map[string]interface{}        `json:"config,omitempty"`
}
