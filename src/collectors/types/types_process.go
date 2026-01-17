package types

// ProcessMetrics contains metrics for a single OS process
// Source: /proc/[pid]/stat, /proc/[pid]/status, /proc/[pid]/statm
type ProcessMetrics struct {
	PID      int64  `json:"pId"`
	PIDT     int64  `json:"pIdT"`
	Name     string `json:"pName"`
	NameT    int64  `json:"pNameT"`
	Cmdline  string `json:"pCmdline"`
	CmdlineT int64  `json:"pCmdlineT"`

	// Thread count
	NumThreads  int64 `json:"pNumThreads"`
	NumThreadsT int64 `json:"pNumThreadsT"`

	// CPU time (centiseconds)
	CPUTimeUserMode     int64 `json:"pCpuTimeUserMode"`
	CPUTimeUserModeT    int64 `json:"pCpuTimeUserModeT"`
	CPUTimeKernelMode   int64 `json:"pCpuTimeKernelMode"`
	CPUTimeKernelModeT  int64 `json:"pCpuTimeKernelModeT"`
	ChildrenUserMode    int64 `json:"pChildrenUserMode"`
	ChildrenUserModeT   int64 `json:"pChildrenUserModeT"`
	ChildrenKernelMode  int64 `json:"pChildrenKernelMode"`
	ChildrenKernelModeT int64 `json:"pChildrenKernelModeT"`

	// Context switches (count)
	VoluntaryContextSwitches     int64 `json:"pVoluntaryContextSwitches"`
	VoluntaryContextSwitchesT    int64 `json:"pVoluntaryContextSwitchesT"`
	NonvoluntaryContextSwitches  int64 `json:"pNonvoluntaryContextSwitches"`
	NonvoluntaryContextSwitchesT int64 `json:"pNonvoluntaryContextSwitchesT"`

	// I/O delays (centiseconds)
	BlockIODelays  int64 `json:"pBlockIODelays"`
	BlockIODelaysT int64 `json:"pBlockIODelaysT"`

	// Memory (bytes)
	VirtualMemoryBytes  int64 `json:"pVirtualMemoryBytes"`
	VirtualMemoryBytesT int64 `json:"pVirtualMemoryBytesT"`
	ResidentSetSize     int64 `json:"pResidentSetSize"`
	ResidentSetSizeT    int64 `json:"pResidentSetSizeT"`
}
