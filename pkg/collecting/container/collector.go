package container

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"context"
	"errors"
	"runtime"
)

type Static struct {
	ContainerID      string `json:"Id"`
	ContainerNumCPUs int64  `json:"NumProcessors"`
	CgroupVersion    int64  `json:"CgroupVersion"`
}

type Dynamic struct {
	ContainerCPUTime           base.MetricInt `json:"CpuTime"`
	ContainerCPUTimeUserMode   base.MetricInt `json:"CpuTimeUserMode"`
	ContainerCPUTimeKernelMode base.MetricInt `json:"CpuTimeKernelMode"`
	ContainerMemoryUsed        base.MetricInt `json:"MemoryUsed"`
	ContainerMemoryMaxUsed     base.MetricInt `json:"MemoryMaxUsed"`
	ContainerPgFault           base.MetricInt `json:"PgFault"`
	ContainerMajorPgFault      base.MetricInt `json:"MajorPgFault"`
	ContainerDiskReadBytes     base.MetricInt `json:"DiskReadBytes"`
	ContainerDiskWriteBytes    base.MetricInt `json:"DiskWriteBytes"`
	ContainerDiskSectorIO      base.MetricInt `json:"DiskSectorIO"`
	ContainerNetworkBytesRecvd base.MetricInt `json:"NetworkBytesRecvd"`
	ContainerNetworkBytesSent  base.MetricInt `json:"NetworkBytesSent"`
	ContainerNumProcesses      base.MetricInt `json:"ProcessCount"`
}

type Collector struct {
	static Static
}

func New() *Collector { return &Collector{} }

func (c *Collector) Name() string { return "Container" }

func (c *Collector) Init(_ *utils.Config) error {
	if v := detect(); v == 0 {
		return errors.New("no cgroup detected")
	}

	c.static = Static{
		ContainerID:      getContainerID(),
		ContainerNumCPUs: int64(runtime.NumCPU()),
		CgroupVersion:    int64(version()),
	}
	return nil
}

func (c *Collector) Static() any { return c.static }

func (c *Collector) Poll(_ context.Context) any {
	m := Dynamic{}
	recv, sent, netTs := getNetStats()
	m.ContainerNetworkBytesRecvd = base.MetricInt{V: recv, T: netTs}
	m.ContainerNetworkBytesSent = base.MetricInt{V: sent, T: netTs}

	switch version() {
	case 1:
		collectV1(&m)
	case 2:
		collectV2(&m)
	}

	return m
}

func (c *Collector) Close() error { return nil }
