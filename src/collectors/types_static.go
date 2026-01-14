package collectors

// StaticMetrics contains all static system information collected once at startup
type StaticMetrics struct {
	UUID                  string  `json:"uuid"`
	VMID                  string  `json:"vId"`
	Hostname              string  `json:"vHostname"`
	BootTime              int64   `json:"vBootTime"`
	NumProcessors         int     `json:"vNumProcessors"`
	CPUType               string  `json:"vCpuType"`
	CPUCache              string  `json:"vCpuCache"`
	KernelInfo            string  `json:"vKernelInfo"`
	TimeSynced            bool    `json:"vTimeSynced"`
	TimeOffsetSeconds     float64 `json:"vTimeOffsetSeconds"`
	TimeMaxErrorSeconds   float64 `json:"vTimeMaxErrorSeconds"`
	MemoryTotalBytes      int64   `json:"vMemoryTotalBytes"`
	SwapTotalBytes        int64   `json:"vSwapTotalBytes"`
	ContainerID           string  `json:"cId"`
	ContainerNumCPUs      int64   `json:"cNumProcessors"`
	CgroupVersion         int64   `json:"cCgroupVersion"`
	NetworkInterfacesJSON string  `json:"networkInterfaces,omitempty"`
	DisksJSON             string  `json:"disks,omitempty"`
	NvidiaDriverVersion   string  `json:"nvidiaDriverVersion,omitempty"`
	NvidiaCudaVersion     string  `json:"nvidiaCudaVersion,omitempty"`
	NvmlVersion           string  `json:"nvmlVersion,omitempty"`
	NvidiaGPUCount        int     `json:"nvidiaGpuCount,omitempty"`
	NvidiaGPUsJSON        string  `json:"nvidiaGpus,omitempty"`
}

// NetworkInterfaceStatic contains static info for a network interface
type NetworkInterfaceStatic struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	State     string `json:"state"`
	MTU       int64  `json:"mtu"`
	SpeedMbps int64  `json:"speedMbps,omitempty"`
}

// DiskStatic contains static info for a disk device
type DiskStatic struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Vendor    string `json:"vendor"`
	SizeBytes int64  `json:"sizeBytes"`
}

// NvidiaGPUStatic contains static information for a single NVIDIA GPU
type NvidiaGPUStatic struct {
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
	MultiGpuBoardId     uint   `json:"multiGpuBoardId,omitempty"`
	DisplayModeEnabled  bool   `json:"displayModeEnabled,omitempty"`
	DisplayActive       bool   `json:"displayActive,omitempty"`
	MigModeEnabled      bool   `json:"migModeEnabled,omitempty"`
	MaxMigInstances     int    `json:"maxMigInstances,omitempty"`
	EncoderCapacityH264 int    `json:"encoderCapacityH264,omitempty"`
	EncoderCapacityHEVC int    `json:"encoderCapacityHevc,omitempty"`
	EncoderCapacityAV1  int    `json:"encoderCapacityAv1,omitempty"`
	NvLinkCount         int    `json:"nvlinkCount,omitempty"`
}
