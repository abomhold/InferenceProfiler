package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type Violations struct {
	Power   base.MetricInt `json:"Power"`
	Thermal base.MetricInt `json:"Thermal"`
}

func CollectViolationsDynamic(d nvml.Device, v *Violations) {
	if status, ret := d.GetViolationStatus(nvml.PERF_POLICY_POWER); ret == nvml.SUCCESS {
		v.Power = base.MetricInt{V: int64(status.ViolationTime), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/violations: GetViolationStatus(POWER) failed: %v", ret)
	}
	if status, ret := d.GetViolationStatus(nvml.PERF_POLICY_THERMAL); ret == nvml.SUCCESS {
		v.Thermal = base.MetricInt{V: int64(status.ViolationTime), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/violations: GetViolationStatus(THERMAL) failed: %v", ret)
	}
}
