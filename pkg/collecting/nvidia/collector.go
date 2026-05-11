package nvidia

import (
	"InferenceProfiler/pkg/collecting/nvidia/domains"
	"InferenceProfiler/pkg/utils"
	"context"
	"fmt"
)

type Static struct {
	Device  domains.DeviceStatic
	Power   domains.PowerStatic
	Memory  domains.MemoryStatic
	Clocks  domains.ClocksStatic
	Thermal domains.ThermalStatic
	PCIe    domains.PCIeStatic
}

type Dynamic struct {
	Index       int
	Power       domains.PowerDynamic
	Memory      domains.MemoryDynamic
	Utilization domains.Utilization
	Clocks      domains.ClocksDynamic
	Thermal     domains.ThermalDynamic
	PCIe        domains.PCIeDynamic
	Violations  domains.Violations
	Processes   domains.Processes
}

type Collector struct {
	nvml             *NVML
	static           []Static
	collectProcesses bool
}

func New() *Collector { return &Collector{} }

func (c *Collector) Name() string { return "Nvidia" }

func (c *Collector) Init(cfg *utils.Config) error {
	c.collectProcesses = !cfg.DisableProcess

	n, err := NewNVML()
	if err != nil {
		return fmt.Errorf("nvidia init: %w", err)
	}
	c.nvml = n
	c.static = make([]Static, n.Count())

	for i, device := range n.Devices() {
		s := &c.static[i]
		domains.CollectDeviceStatic(device, i, &s.Device)
		domains.CollectPowerStatic(device, &s.Power)
		domains.CollectMemoryStatic(device, &s.Memory)
		domains.CollectClocksStatic(device, &s.Clocks)
		domains.CollectThermalStatic(device, &s.Thermal)
		domains.CollectPCIeStatic(device, &s.PCIe)
	}

	return nil
}

func (c *Collector) Static() any { return c.static }

func (c *Collector) Poll(_ context.Context) any {
	devices := c.nvml.Devices()
	result := make([]Dynamic, len(devices))

	for i, device := range devices {
		d := &result[i]
		d.Index = i

		domains.CollectMemoryDynamic(device, &d.Memory)
		domains.CollectUtilizationDynamic(device, &d.Utilization)
		domains.CollectClocksDynamic(device, &d.Clocks)
		domains.CollectThermalDynamic(device, &d.Thermal)
		domains.CollectPowerDynamic(device, &d.Power)
		domains.CollectViolationsDynamic(device, &d.Violations)
		domains.CollectPCIeReplayCounter(device, &d.PCIe)
		domains.CollectPCIeThroughput(device, &d.PCIe)

		if c.collectProcesses {
			domains.CollectProcessesDynamic(device, &d.Processes)
		}
	}

	return result
}

func (c *Collector) Close() error {
	if c.nvml != nil {
		return c.nvml.Close()
	}
	return nil
}
