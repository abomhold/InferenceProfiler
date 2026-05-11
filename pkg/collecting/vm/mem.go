package vm

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"strings"
)

var (
	procMeminfo = "/proc/meminfo"
	procVmstat  = "/proc/vmstat"
)

type MemStatic struct {
	TotalBytes     int64 `json:"Total"`
	SwapTotalBytes int64 `json:"SwapTotal"`
}

type MemDynamic struct {
	Total          base.MetricInt `json:"Total"`
	Free           base.MetricInt `json:"Free"`
	Buffers        base.MetricInt `json:"Buffers"`
	Cached         base.MetricInt `json:"Cached"`
	SwapTotal      base.MetricInt `json:"SwapTotal"`
	SwapFree       base.MetricInt `json:"SwapFree"`
	PgFault        base.MetricInt `json:"PgFault"`
	MajorPageFault base.MetricInt `json:"MajorPageFault"`
}

func collectMemStatic(s *MemStatic) {
	lines, _, err := utils.FileLines(procMeminfo)
	if err != nil {
		utils.Debugf("mem: failed to read %s: %v", procMeminfo, err)
		return
	}
	for _, line := range lines {
		parts := strings.SplitN(line, utils.FieldSeparatorColon, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := utils.ParseInt64(strings.Fields(parts[1])[0])

		switch key {
		case "MemTotal":
			s.TotalBytes = val
		case "SwapTotal":
			s.SwapTotalBytes = val
		}
	}
	utils.Debugf("mem: static total=%dKB swap=%dKB", s.TotalBytes, s.SwapTotalBytes)
}

func collectMemDynamic(d *MemDynamic) {
	lines, ts, err := utils.FileLines(procMeminfo)
	if err != nil {
		utils.Debugf("mem: failed to read %s: %v", procMeminfo, err)
		return
	}
	for _, line := range lines {
		parts := strings.SplitN(line, utils.FieldSeparatorColon, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) == 0 {
			continue
		}
		val := utils.ParseInt64(fields[0])

		switch key {
		case "MemTotal":
			d.Total = base.MetricInt{V: val, T: ts}
		case "MemFree":
			d.Free = base.MetricInt{V: val, T: ts}
		case "Buffers":
			d.Buffers = base.MetricInt{V: val, T: ts}
		case "Cached":
			d.Cached = base.MetricInt{V: val, T: ts}
		case "SwapTotal":
			d.SwapTotal = base.MetricInt{V: val, T: ts}
		case "SwapFree":
			d.SwapFree = base.MetricInt{V: val, T: ts}
		}
	}

	vmLines, vmTs, err := utils.FileLines(procVmstat)
	if err != nil {
		utils.Debugf("mem: failed to read %s: %v", procVmstat, err)
		return
	}
	for _, line := range vmLines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "pgfault":
			d.PgFault = base.MetricInt{V: utils.ParseInt64(fields[1]), T: vmTs}
		case "pgmajfault":
			d.MajorPageFault = base.MetricInt{V: utils.ParseInt64(fields[1]), T: vmTs}
		}
	}
}
