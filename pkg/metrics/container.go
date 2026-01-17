package metrics

type ContainerStatic struct {
	ContainerID      string `json:"cId"`
	ContainerNumCPUs int64  `json:"cNumProcessors"`
	CgroupVersion    int64  `json:"cCgroupVersion"`
}

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
	ContainerDiskSectorIO       int64  `json:"cDiskSectorIO"`
	ContainerDiskSectorIOT      int64  `json:"cDiskSectorIOT"`
	ContainerNetworkBytesRecvd  int64  `json:"cNetworkBytesRecvd"`
	ContainerNetworkBytesRecvdT int64  `json:"cNetworkBytesRecvdT"`
	ContainerNetworkBytesSent   int64  `json:"cNetworkBytesSent"`
	ContainerNetworkBytesSentT  int64  `json:"cNetworkBytesSentT"`
	ContainerNumProcesses       int64  `json:"cNumProcesses"`
	ContainerNumProcessesT      int64  `json:"cNumProcessesT"`
}
