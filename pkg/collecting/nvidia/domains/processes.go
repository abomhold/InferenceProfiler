package domains

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"errors"
	"sort"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type Processes struct {
	Count     int
	Timestamp int64
	List      []ProcessStats
}

type ProcessStats struct {
	PID        uint32         `json:"PID"`
	Type       base.MetricStr `json:"Type"`
	SM         base.MetricInt `json:"SM"`
	MemUtil    base.MetricInt `json:"MemUtil"`
	Encoder    base.MetricInt `json:"Encoder"`
	Decoder    base.MetricInt `json:"Decoder"`
	MemoryUsed base.MetricInt `json:"MemoryUsed"`
}

func CollectProcessesDynamic(d nvml.Device, p *Processes) {
	procMap := make(map[uint32]*ProcessStats)
	p.List = p.List[:0]

	addProcess := func(pid uint32, mem uint64, procType string, t int64) {
		if _, exists := procMap[pid]; !exists {
			procMap[pid] = &ProcessStats{
				PID:        pid,
				MemoryUsed: base.MetricInt{V: int64(mem), T: t},
				Type:       base.MetricStr{V: procType, T: t},
			}
		}
	}

	if list, ret := d.GetComputeRunningProcesses(); errors.Is(ret, nvml.SUCCESS) {
		now := utils.GetTimestamp()
		for _, proc := range list {
			addProcess(proc.Pid, proc.UsedGpuMemory, "compute", now)
		}
		utils.Debugf("nvidia/processes: %d compute processes", len(list))
	} else {
		utils.Debugf("nvidia/processes: GetComputeRunningProcesses failed: %v", ret)
	}

	if list, ret := d.GetGraphicsRunningProcesses(); ret == nvml.SUCCESS {
		now := utils.GetTimestamp()
		for _, proc := range list {
			addProcess(proc.Pid, proc.UsedGpuMemory, "graphics", now)
		}
		if len(list) > 0 {
			utils.Debugf("nvidia/processes: %d graphics processes", len(list))
		}
	} else {
		utils.Debugf("nvidia/processes: GetGraphicsRunningProcesses failed: %v", ret)
	}

	if list, ret := d.GetMPSComputeRunningProcesses(); ret == nvml.SUCCESS {
		now := utils.GetTimestamp()
		for _, proc := range list {
			addProcess(proc.Pid, proc.UsedGpuMemory, "mps", now)
		}
		if len(list) > 0 {
			utils.Debugf("nvidia/processes: %d MPS processes", len(list))
		}
	} else {
		utils.Debugf("nvidia/processes: GetMPSComputeRunningProcesses failed: %v", ret)
	}

	if samples, ret := d.GetProcessUtilization(0); ret == nvml.SUCCESS {
		for _, sample := range samples {
			if stats, exists := procMap[sample.Pid]; exists {
				ts := int64(sample.TimeStamp)
				stats.SM = base.MetricInt{V: int64(sample.SmUtil), T: ts}
				stats.MemUtil = base.MetricInt{V: int64(sample.MemUtil), T: ts}
				stats.Encoder = base.MetricInt{V: int64(sample.EncUtil), T: ts}
				stats.Decoder = base.MetricInt{V: int64(sample.DecUtil), T: ts}
			}
		}
		utils.Debugf("nvidia/processes: %d utilization samples", len(samples))
	} else {
		utils.Debugf("nvidia/processes: GetProcessUtilization failed: %v", ret)
	}

	for _, stats := range procMap {
		p.List = append(p.List, *stats)
	}

	sort.Slice(p.List, func(i, j int) bool {
		return p.List[i].PID < p.List[j].PID
	})

	p.Count = len(p.List)
	p.Timestamp = utils.GetTimestamp()
}
