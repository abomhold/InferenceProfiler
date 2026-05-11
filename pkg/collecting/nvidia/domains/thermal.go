package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type ThermalStatic struct {
	ShutdownThreshold int64
	SlowdownThreshold int64
	MaxOperating      int64
}

type ThermalDynamic struct {
	GPU base.MetricInt `json:"GPU"`
}

func CollectThermalStatic(d nvml.Device, t *ThermalStatic) {
	if v, ret := d.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SHUTDOWN); ret == nvml.SUCCESS {
		t.ShutdownThreshold = int64(v)
	} else {
		utils.Debugf("nvidia/thermal: GetTemperatureThreshold(SHUTDOWN) failed: %v", ret)
	}
	if v, ret := d.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SLOWDOWN); ret == nvml.SUCCESS {
		t.SlowdownThreshold = int64(v)
	} else {
		utils.Debugf("nvidia/thermal: GetTemperatureThreshold(SLOWDOWN) failed: %v", ret)
	}
	if v, ret := d.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_GPU_MAX); ret == nvml.SUCCESS {
		t.MaxOperating = int64(v)
	} else {
		utils.Debugf("nvidia/thermal: GetTemperatureThreshold(GPU_MAX) failed: %v", ret)
	}
	utils.Debugf("nvidia/thermal: static shutdown=%d°C slowdown=%d°C maxOp=%d°C",
		t.ShutdownThreshold, t.SlowdownThreshold, t.MaxOperating)
}

func CollectThermalDynamic(d nvml.Device, t *ThermalDynamic) {
	if v, ret := d.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		t.GPU = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/thermal: GetTemperature(GPU) failed: %v", ret)
	}
}
