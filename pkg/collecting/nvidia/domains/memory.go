package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type MemoryStatic struct {
	Total     int64
	Bar1Total int64
}

func CollectMemoryStatic(d nvml.Device, m *MemoryStatic) {
	if mem, ret := d.GetMemoryInfo(); ret == nvml.SUCCESS {
		m.Total = int64(mem.Total)
	} else {
		utils.Debugf("nvidia/memory: GetMemoryInfo failed: %v", ret)
	}
	if bar1, ret := d.GetBAR1MemoryInfo(); ret == nvml.SUCCESS {
		m.Bar1Total = int64(bar1.Bar1Total)
	} else {
		utils.Debugf("nvidia/memory: GetBAR1MemoryInfo failed: %v", ret)
	}
	utils.Debugf("nvidia/memory: static total=%dMB bar1=%dMB",
		m.Total/(1024*1024), m.Bar1Total/(1024*1024))
}

type MemoryDynamic struct {
	Used     base.MetricInt `json:"Used"`
	Free     base.MetricInt `json:"Free"`
	Bar1Used base.MetricInt `json:"Bar1Used"`
	Bar1Free base.MetricInt `json:"Bar1Free"`
}

func CollectMemoryDynamic(d nvml.Device, m *MemoryDynamic) {
	now := utils.GetTimestamp()
	if mem, ret := d.GetMemoryInfo(); ret == nvml.SUCCESS {
		m.Used = base.MetricInt{V: int64(mem.Used), T: now}
		m.Free = base.MetricInt{V: int64(mem.Free), T: now}
	} else {
		utils.Debugf("nvidia/memory: GetMemoryInfo dynamic failed: %v", ret)
	}

	now = utils.GetTimestamp()
	if bar1, ret := d.GetBAR1MemoryInfo(); ret == nvml.SUCCESS {
		m.Bar1Used = base.MetricInt{V: int64(bar1.Bar1Used), T: now}
		m.Bar1Free = base.MetricInt{V: int64(bar1.Bar1Free), T: now}
	} else {
		utils.Debugf("nvidia/memory: GetBAR1MemoryInfo dynamic failed: %v", ret)
	}
}
