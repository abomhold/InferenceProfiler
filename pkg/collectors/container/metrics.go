package container

import "InferenceProfiler/pkg/collectors/types"

// Static contains static container information.
type Static struct {
	ContainerID      string
	ContainerNumCPUs int64
	CgroupVersion    int64
}

// ToRecord converts Static to a Record.
func (s *Static) ToRecord() types.Record {
	return types.Record{
		"cId":            s.ContainerID,
		"cNumProcessors": s.ContainerNumCPUs,
		"cCgroupVersion": s.CgroupVersion,
	}
}

// Dynamic contains dynamic container metrics.
type Dynamic struct {
	ContainerCPUTime            int64
	ContainerCPUTimeT           int64
	ContainerCPUTimeUserMode    int64
	ContainerCPUTimeUserModeT   int64
	ContainerCPUTimeKernelMode  int64
	ContainerCPUTimeKernelModeT int64
	ContainerPerCPUTimesJSON    string
	ContainerPerCPUTimesT       int64
	ContainerMemoryUsed         int64
	ContainerMemoryUsedT        int64
	ContainerMemoryMaxUsed      int64
	ContainerMemoryMaxUsedT     int64
	ContainerPgFault            int64
	ContainerPgFaultT           int64
	ContainerMajorPgFault       int64
	ContainerMajorPgFaultT      int64
	ContainerDiskReadBytes      int64
	ContainerDiskReadBytesT     int64
	ContainerDiskWriteBytes     int64
	ContainerDiskWriteBytesT    int64
	ContainerDiskSectorIO       int64
	ContainerDiskSectorIOT      int64
	ContainerNetworkBytesRecvd  int64
	ContainerNetworkBytesRecvdT int64
	ContainerNetworkBytesSent   int64
	ContainerNetworkBytesSentT  int64
	ContainerNumProcesses       int64
	ContainerNumProcessesT      int64
}

// ToRecord converts Dynamic to a Record.
func (d *Dynamic) ToRecord() types.Record {
	r := types.Record{
		"cCpuTime":            d.ContainerCPUTime,
		"cCpuTimeT":           d.ContainerCPUTimeT,
		"cCpuTimeUserMode":    d.ContainerCPUTimeUserMode,
		"cCpuTimeUserModeT":   d.ContainerCPUTimeUserModeT,
		"cCpuTimeKernelMode":  d.ContainerCPUTimeKernelMode,
		"cCpuTimeKernelModeT": d.ContainerCPUTimeKernelModeT,
		"cMemoryUsed":         d.ContainerMemoryUsed,
		"cMemoryUsedT":        d.ContainerMemoryUsedT,
		"cMemoryMaxUsed":      d.ContainerMemoryMaxUsed,
		"cMemoryMaxUsedT":     d.ContainerMemoryMaxUsedT,
		"cPgFault":            d.ContainerPgFault,
		"cPgFaultT":           d.ContainerPgFaultT,
		"cMajorPgFault":       d.ContainerMajorPgFault,
		"cMajorPgFaultT":      d.ContainerMajorPgFaultT,
		"cDiskReadBytes":      d.ContainerDiskReadBytes,
		"cDiskReadBytesT":     d.ContainerDiskReadBytesT,
		"cDiskWriteBytes":     d.ContainerDiskWriteBytes,
		"cDiskWriteBytesT":    d.ContainerDiskWriteBytesT,
		"cNetworkBytesRecvd":  d.ContainerNetworkBytesRecvd,
		"cNetworkBytesRecvdT": d.ContainerNetworkBytesRecvdT,
		"cNetworkBytesSent":   d.ContainerNetworkBytesSent,
		"cNetworkBytesSentT":  d.ContainerNetworkBytesSentT,
		"cNumProcesses":       d.ContainerNumProcesses,
		"cNumProcessesT":      d.ContainerNumProcessesT,
	}

	// Only include optional fields if they have values
	if d.ContainerPerCPUTimesJSON != "" {
		r["cCpuPerCpuJson"] = d.ContainerPerCPUTimesJSON
		r["cCpuPerCpuT"] = d.ContainerPerCPUTimesT
	}
	if d.ContainerDiskSectorIO != 0 {
		r["cDiskSectorIO"] = d.ContainerDiskSectorIO
		r["cDiskSectorIOT"] = d.ContainerDiskSectorIOT
	}

	return r
}
