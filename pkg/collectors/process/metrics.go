package process

// Info contains information for a single process.
type Info struct {
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
