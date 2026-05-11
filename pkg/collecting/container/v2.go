package container

import (
	"InferenceProfiler/pkg/collecting/base"
	"path/filepath"
	"strings"

	"InferenceProfiler/pkg/utils"
)

func collectV2(m *Dynamic) {
	cpuStats, tCpu, err := utils.FileKV(filepath.Join(v2Path, "cpu.stat"), utils.FieldSeparatorSpace)
	if err != nil {
		utils.Debugf("container/v2: cpu.stat read failed: %v", err)
	}
	memUsage, tMem, err := utils.FileInt(filepath.Join(v2Path, "memory.current"))
	if err != nil {
		utils.Debugf("container/v2: memory.current read failed: %v", err)
	}
	memPeak, tPeak, err := utils.FileInt(filepath.Join(v2Path, "memory.peak"))
	if err != nil {
		utils.Debugf("container/v2: memory.peak read failed: %v", err)
	}
	memStat, tMemStat, err := utils.FileKV(filepath.Join(v2Path, "memory.stat"), utils.FieldSeparatorSpace)
	if err != nil {
		utils.Debugf("container/v2: memory.stat read failed: %v", err)
	}
	dr, dw, tIO := getIOStatV2()
	numProcs, tProcs, err := utils.FileInt(filepath.Join(v2Path, "pids.current"))
	if err != nil {
		utils.Debugf("container/v2: pids.current read failed: %v", err)
	}

	utils.Debugf("container/v2: cpu_usec=%s mem=%d peak=%d pids=%d io_r=%d io_w=%d",
		cpuStats["usage_usec"], memUsage, memPeak, numProcs, dr, dw)

	m.ContainerCPUTime = base.MetricInt{V: utils.ParseInt64(cpuStats["usage_usec"]), T: tCpu}
	m.ContainerCPUTimeUserMode = base.MetricInt{V: utils.ParseInt64(cpuStats["user_usec"]), T: tCpu}
	m.ContainerCPUTimeKernelMode = base.MetricInt{V: utils.ParseInt64(cpuStats["system_usec"]), T: tCpu}
	m.ContainerMemoryUsed = base.MetricInt{V: memUsage, T: tMem}
	m.ContainerMemoryMaxUsed = base.MetricInt{V: memPeak, T: tPeak}
	m.ContainerPgFault = base.MetricInt{V: utils.ParseInt64(memStat["pgfault"]), T: tMemStat}
	m.ContainerMajorPgFault = base.MetricInt{V: utils.ParseInt64(memStat["pgmajfault"]), T: tMemStat}
	m.ContainerDiskReadBytes = base.MetricInt{V: dr, T: tIO}
	m.ContainerDiskWriteBytes = base.MetricInt{V: dw, T: tIO}
	m.ContainerNumProcesses = base.MetricInt{V: numProcs, T: tProcs}
}

func getIOStatV2() (r, w, ts int64) {
	ioStatPath := filepath.Join(v2Path, "io.stat")
	if !utils.Exists(ioStatPath) {
		return 0, 0, utils.GetTimestamp()
	}
	lines, ts, _ := utils.FileLines(ioStatPath)
	for _, line := range lines {
		for _, p := range strings.Fields(line) {
			if v, ok := strings.CutPrefix(p, "rbytes="); ok {
				r += utils.ParseInt64(v)
			}
			if v, ok := strings.CutPrefix(p, "wbytes="); ok {
				w += utils.ParseInt64(v)
			}
		}
	}
	return r, w, ts
}
