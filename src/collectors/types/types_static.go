package types

// StaticMetrics contains all static system information collected once at startup
type StaticMetrics struct {
	// Session identification
	UUID     string `json:"uuid"`
	VMID     string `json:"vId"`
	Hostname string `json:"vHostname"`

	// BootTime is the system boot time as Unix timestamp (seconds)
	BootTime int64 `json:"vBootTime"`

	// CPU static info
	NumProcessors int    `json:"vNumProcessors"` // Logical CPU count
	CPUType       string `json:"vCpuType"`       // Processor model name
	CPUCache      string `json:"vCpuCache"`      // Cache sizes (e.g., "L1d:32K L1i:32K L2:256K L3:8M")
	KernelInfo    string `json:"vKernelInfo"`    // Full kernel version string

	// Time synchronization (from adjtimex syscall)
	TimeSynced          bool    `json:"vTimeSynced"`          // Whether kernel clock is synchronized
	TimeOffsetSeconds   float64 `json:"vTimeOffsetSeconds"`   // Offset from reference (seconds)
	TimeMaxErrorSeconds float64 `json:"vTimeMaxErrorSeconds"` // Maximum time error (seconds)

	// Memory static info (bytes)
	MemoryTotalBytes int64 `json:"vMemoryTotalBytes"` // Total physical RAM
	SwapTotalBytes   int64 `json:"vSwapTotalBytes"`   // Total swap space

	// Container static info
	ContainerID      string `json:"cId"`            // Container ID or hostname
	ContainerNumCPUs int64  `json:"cNumProcessors"` // Available processors
	CgroupVersion    int64  `json:"cCgroupVersion"` // Cgroup version (1 or 2)

	// Network static info (JSON serialized)
	NetworkInterfacesJSON string `json:"networkInterfaces,omitempty"`

	// Disk static info (JSON serialized)
	DisksJSON string `json:"disks,omitempty"`

	// NVIDIA GPU static info
	NvidiaDriverVersion string `json:"nvidiaDriverVersion,omitempty"` // Driver version
	NvidiaCudaVersion   string `json:"nvidiaCudaVersion,omitempty"`   // CUDA driver version
	NvmlVersion         string `json:"nvmlVersion,omitempty"`         // NVML library version
	NvidiaGPUCount      int    `json:"nvidiaGpuCount,omitempty"`      // Number of GPUs
	NvidiaGPUsJSON      string `json:"nvidiaGpus,omitempty"`          // GPU details (JSON)
}

// NetworkInterfaceStatic contains static info for a network interface
type NetworkInterfaceStatic struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	State     string `json:"state"`               // Operational state (up/down)
	MTU       int64  `json:"mtu"`                 // Maximum transmission unit
	SpeedMbps int64  `json:"speedMbps,omitempty"` // Link speed (if available)
}

// DiskStatic contains static info for a disk device
type DiskStatic struct {
	Name      string `json:"name"`      // Device name (e.g., "sda", "nvme0n1")
	Model     string `json:"model"`     // Disk model
	Vendor    string `json:"vendor"`    // Disk vendor
	SizeBytes int64  `json:"sizeBytes"` // Total size
}

// NvidiaGPUStatic contains static information for a single NVIDIA GPU
type NvidiaGPUStatic struct {
	// Basic identification
	Index           int    `json:"index"`
	Name            string `json:"name"`
	UUID            string `json:"uuid"`
	Serial          string `json:"serial,omitempty"`
	BoardPartNumber string `json:"boardPartNumber,omitempty"`
	Brand           string `json:"brand,omitempty"`

	// Architecture and compute capability
	Architecture        string `json:"architecture,omitempty"`
	CudaCapabilityMajor int    `json:"cudaCapabilityMajor,omitempty"`
	CudaCapabilityMinor int    `json:"cudaCapabilityMinor,omitempty"`

	// Memory specifications (bytes)
	MemoryTotalBytes   int64 `json:"memoryTotalBytes"`
	Bar1TotalBytes     int64 `json:"bar1TotalBytes,omitempty"`
	MemoryBusWidthBits int   `json:"memoryBusWidthBits,omitempty"`

	// Compute specifications
	NumCores            int `json:"numCores,omitempty"`
	MaxClockGraphicsMhz int `json:"maxClockGraphicsMhz,omitempty"`
	MaxClockMemoryMhz   int `json:"maxClockMemoryMhz,omitempty"`
	MaxClockSmMhz       int `json:"maxClockSmMhz,omitempty"`
	MaxClockVideoMhz    int `json:"maxClockVideoMhz,omitempty"`

	// PCI information
	PciBusId         string `json:"pciBusId,omitempty"`
	PciDeviceId      uint32 `json:"pciDeviceId,omitempty"`
	PciSubsystemId   uint32 `json:"pciSubsystemId,omitempty"`
	PcieMaxLinkGen   int    `json:"pcieMaxLinkGen,omitempty"`
	PcieMaxLinkWidth int    `json:"pcieMaxLinkWidth,omitempty"`

	// Power specifications (milliwatts)
	PowerDefaultLimitMw int `json:"powerDefaultLimitMw,omitempty"`
	PowerMinLimitMw     int `json:"powerMinLimitMw,omitempty"`
	PowerMaxLimitMw     int `json:"powerMaxLimitMw,omitempty"`

	// Firmware versions
	VbiosVersion        string `json:"vbiosVersion,omitempty"`
	InforomImageVersion string `json:"inforomImageVersion,omitempty"`
	InforomOemVersion   string `json:"inforomOemVersion,omitempty"`

	// Thermal specifications (Celsius)
	NumFans           int `json:"numFans,omitempty"`
	TempShutdownC     int `json:"tempShutdownC,omitempty"`
	TempSlowdownC     int `json:"tempSlowdownC,omitempty"`
	TempMaxOperatingC int `json:"tempMaxOperatingC,omitempty"`
	TempTargetC       int `json:"tempTargetC,omitempty"`

	// Configuration state
	EccModeEnabled     bool   `json:"eccModeEnabled,omitempty"`
	PersistenceModeOn  bool   `json:"persistenceModeOn,omitempty"`
	ComputeMode        string `json:"computeMode,omitempty"`
	IsMultiGpuBoard    bool   `json:"isMultiGpuBoard,omitempty"`
	MultiGpuBoardId    uint   `json:"multiGpuBoardId,omitempty"`
	DisplayModeEnabled bool   `json:"displayModeEnabled,omitempty"`
	DisplayActive      bool   `json:"displayActive,omitempty"`

	// MIG capabilities
	MigModeEnabled  bool `json:"migModeEnabled,omitempty"`
	MaxMigInstances int  `json:"maxMigInstances,omitempty"`

	// Encoder/decoder capabilities
	EncoderCapacityH264 int `json:"encoderCapacityH264,omitempty"`
	EncoderCapacityHEVC int `json:"encoderCapacityHevc,omitempty"`
	EncoderCapacityAV1  int `json:"encoderCapacityAv1,omitempty"`

	// NVLink info
	NvLinkCount int `json:"nvlinkCount,omitempty"`
}
