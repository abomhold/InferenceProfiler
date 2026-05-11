package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type Utilization struct {
	GPU    base.MetricInt `json:"GPU"`
	Memory base.MetricInt `json:"Memory"`
}

func CollectUtilizationDynamic(d nvml.Device, u *Utilization) {
	if util, ret := d.GetUtilizationRates(); ret == nvml.SUCCESS {
		now := utils.GetTimestamp()
		u.GPU = base.MetricInt{V: int64(util.Gpu), T: now}
		u.Memory = base.MetricInt{V: int64(util.Memory), T: now}
	} else {
		utils.Debugf("nvidia/util: GetUtilizationRates failed: %v", ret)
	}
}
