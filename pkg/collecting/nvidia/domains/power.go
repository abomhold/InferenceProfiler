package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"errors"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type PowerStatic struct {
	DefaultLimit  int64
	MinLimit      int64
	MaxLimit      int64
	Limit         int64
	EnforcedLimit int64
}

func CollectPowerStatic(d nvml.Device, p *PowerStatic) {
	if v, ret := d.GetPowerManagementDefaultLimit(); errors.Is(ret, nvml.SUCCESS) {
		p.DefaultLimit = int64(v)
	} else {
		utils.Debugf("nvidia/power: GetPowerManagementDefaultLimit failed: %v", ret)
	}
	if mini, maxi, ret := d.GetPowerManagementLimitConstraints(); errors.Is(ret, nvml.SUCCESS) {
		p.MinLimit = int64(mini)
		p.MaxLimit = int64(maxi)
	} else {
		utils.Debugf("nvidia/power: GetPowerManagementLimitConstraints failed: %v", ret)
	}
	if v, ret := d.GetPowerManagementLimit(); ret == nvml.SUCCESS {
		p.Limit = int64(v)
	} else {
		utils.Debugf("nvidia/power: GetPowerManagementLimit failed: %v", ret)
	}
	if v, ret := d.GetEnforcedPowerLimit(); ret == nvml.SUCCESS {
		p.EnforcedLimit = int64(v)
	} else {
		utils.Debugf("nvidia/power: GetEnforcedPowerLimit failed: %v", ret)
	}
	utils.Debugf("nvidia/power: static default=%dmW limit=%dmW enforced=%dmW range=[%d-%d]mW",
		p.DefaultLimit, p.Limit, p.EnforcedLimit, p.MinLimit, p.MaxLimit)
}

type PowerDynamic struct {
	Usage  base.MetricInt `json:"Usage"`
	Energy base.MetricInt `json:"Energy"`
}

func CollectPowerDynamic(d nvml.Device, p *PowerDynamic) {
	if v, ret := d.GetPowerUsage(); ret == nvml.SUCCESS {
		p.Usage = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/power: GetPowerUsage failed: %v", ret)
	}
	if v, ret := d.GetTotalEnergyConsumption(); ret == nvml.SUCCESS {
		p.Energy = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/power: GetTotalEnergyConsumption failed: %v", ret)
	}
}
