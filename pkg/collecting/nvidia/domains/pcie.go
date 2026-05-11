package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type PCIeStatic struct {
	BusID        string
	MaxLinkGen   int
	MaxLinkWidth int
	LinkGen      int
	LinkWidth    int
}

func CollectPCIeStatic(d nvml.Device, p *PCIeStatic) {
	if pci, ret := d.GetPciInfo(); ret == nvml.SUCCESS {
		p.BusID = utils.ByteSliceToString(pci.BusId[:])
	} else {
		utils.Debugf("nvidia/pcie: GetPciInfo failed: %v", ret)
	}
	if v, ret := d.GetMaxPcieLinkGeneration(); ret == nvml.SUCCESS {
		p.MaxLinkGen = v
	} else {
		utils.Debugf("nvidia/pcie: GetMaxPcieLinkGeneration failed: %v", ret)
	}
	if v, ret := d.GetMaxPcieLinkWidth(); ret == nvml.SUCCESS {
		p.MaxLinkWidth = v
	} else {
		utils.Debugf("nvidia/pcie: GetMaxPcieLinkWidth failed: %v", ret)
	}
	if v, ret := d.GetCurrPcieLinkGeneration(); ret == nvml.SUCCESS {
		p.LinkGen = v
	} else {
		utils.Debugf("nvidia/pcie: GetCurrPcieLinkGeneration failed: %v", ret)
	}
	if v, ret := d.GetCurrPcieLinkWidth(); ret == nvml.SUCCESS {
		p.LinkWidth = v
	} else {
		utils.Debugf("nvidia/pcie: GetCurrPcieLinkWidth failed: %v", ret)
	}
	utils.Debugf("nvidia/pcie: bus=%s gen%d x%d (max gen%d x%d)",
		p.BusID, p.LinkGen, p.LinkWidth, p.MaxLinkGen, p.MaxLinkWidth)
}

type PCIeDynamic struct {
	TxThroughput  base.MetricInt `json:"TxThroughput"`
	RxThroughput  base.MetricInt `json:"RxThroughput"`
	ReplayCounter base.MetricInt `json:"ReplayCounter"`
}

func CollectPCIeReplayCounter(d nvml.Device, p *PCIeDynamic) {
	if v, ret := d.GetPcieReplayCounter(); ret == nvml.SUCCESS {
		p.ReplayCounter = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/pcie: GetPcieReplayCounter failed: %v", ret)
	}
}

func CollectPCIeThroughput(d nvml.Device, p *PCIeDynamic) {
	if v, ret := d.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); ret == nvml.SUCCESS {
		p.TxThroughput = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/pcie: GetPcieThroughput(TX) failed: %v", ret)
	}
	if v, ret := d.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); ret == nvml.SUCCESS {
		p.RxThroughput = base.MetricInt{V: int64(v), T: utils.GetTimestamp()}
	} else {
		utils.Debugf("nvidia/pcie: GetPcieThroughput(RX) failed: %v", ret)
	}
}
