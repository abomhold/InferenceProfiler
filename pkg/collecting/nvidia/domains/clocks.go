package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type ClocksStatic struct {
	MaxGraphics int64
	MaxSM       int64
	MaxMemory   int64
}

func CollectClocksStatic(d nvml.Device, c *ClocksStatic) {
	if v, ret := d.GetMaxClockInfo(nvml.CLOCK_GRAPHICS); ret == nvml.SUCCESS {
		c.MaxGraphics = int64(v)
	} else {
		utils.Debugf("nvidia/clocks: GetMaxClockInfo(GRAPHICS) failed: %v", ret)
	}
	if v, ret := d.GetMaxClockInfo(nvml.CLOCK_SM); ret == nvml.SUCCESS {
		c.MaxSM = int64(v)
	} else {
		utils.Debugf("nvidia/clocks: GetMaxClockInfo(SM) failed: %v", ret)
	}
	if v, ret := d.GetMaxClockInfo(nvml.CLOCK_MEM); ret == nvml.SUCCESS {
		c.MaxMemory = int64(v)
	} else {
		utils.Debugf("nvidia/clocks: GetMaxClockInfo(MEM) failed: %v", ret)
	}
	utils.Debugf("nvidia/clocks: static maxGraphics=%d maxSM=%d maxMem=%d",
		c.MaxGraphics, c.MaxSM, c.MaxMemory)
}

type ClocksDynamic struct {
	Graphics        base.MetricInt `json:"Graphics"`
	SM              base.MetricInt `json:"SM"`
	Memory          base.MetricInt `json:"Memory"`
	Video           base.MetricInt `json:"Video"`
	PState          base.MetricInt `json:"PState"`
	ThrottleReasons base.MetricInt `json:"ThrottleReasons"`
}

func CollectClocksDynamic(d nvml.Device, c *ClocksDynamic) {
	if v, ret := d.GetClockInfo(nvml.CLOCK_GRAPHICS); ret == nvml.SUCCESS {
		c.Graphics = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/clocks: GetClockInfo(GRAPHICS) failed: %v", ret)
	}
	if v, ret := d.GetClockInfo(nvml.CLOCK_SM); ret == nvml.SUCCESS {
		c.SM = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/clocks: GetClockInfo(SM) failed: %v", ret)
	}
	if v, ret := d.GetClockInfo(nvml.CLOCK_MEM); ret == nvml.SUCCESS {
		c.Memory = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/clocks: GetClockInfo(MEM) failed: %v", ret)
	}
	if v, ret := d.GetClockInfo(nvml.CLOCK_VIDEO); ret == nvml.SUCCESS {
		c.Video = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/clocks: GetClockInfo(VIDEO) failed: %v", ret)
	}
	if v, ret := d.GetPerformanceState(); ret == nvml.SUCCESS {
		c.PState = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/clocks: GetPerformanceState failed: %v", ret)
	}
	if v, ret := d.GetCurrentClocksEventReasons(); ret == nvml.SUCCESS {
		c.ThrottleReasons = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/clocks: GetCurrentClocksEventReasons failed: %v", ret)
	}
}
