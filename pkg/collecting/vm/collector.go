package vm

import (
	"InferenceProfiler/pkg/utils"
	"context"
)

type Static struct {
	CPU     CpuStatic  `json:"Cpu"`
	Memory  MemStatic  `json:"Mem"`
	Disk    DiskStatic `json:"Disk"`
	Network NetStatic  `json:"Net"`
}

type Dynamic struct {
	CPU     CpuDynamic  `json:"Cpu"`
	Memory  MemDynamic  `json:"Mem"`
	Disk    DiskDynamic `json:"Disk"`
	Network NetDynamic  `json:"Net"`
}

type Collector struct {
	static Static
}

func New() *Collector { return &Collector{} }

func (c *Collector) Name() string { return "Vm" }

func (c *Collector) Init(_ *utils.Config) error {
	collectCpuStatic(&c.static.CPU)
	collectMemStatic(&c.static.Memory)
	collectDiskStatic(&c.static.Disk)
	collectNetStatic(&c.static.Network)
	return nil
}

func (c *Collector) Static() any { return c.static }

func (c *Collector) Poll(_ context.Context) any {
	d := Dynamic{}
	collectCpuDynamic(&d.CPU)
	collectMemDynamic(&d.Memory)
	collectDiskDynamic(&d.Disk)
	collectNetDynamic(&d.Network)
	return d
}

func (c *Collector) Close() error { return nil }
