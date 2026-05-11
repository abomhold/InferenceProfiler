package container

import (
	"InferenceProfiler/pkg/collecting/base"
	"path/filepath"
	"strings"

	"InferenceProfiler/pkg/utils"
)

func collectV1(m *Dynamic) {
	cpuPath := filepath.Join(cgroupDir, "cpuacct")
	memPath := filepath.Join(cgroupDir, "memory")
	blkioPath := filepath.Join(cgroupDir, "blkio")
	pidsPath := filepath.Join(cgroupDir, "pids")

	cpuUsage, tCpu, err := utils.FileInt(filepath.Join(cpuPath, "cpuacct.usage"))
	if err != nil {
		utils.Debugf("container/v1: cpuacct.usage read failed: %v", err)
	}
	cpuStat, tCpuStat, err := utils.FileKV(filepath.Join(cpuPath, "cpuacct.stat"), utils.FieldSeparatorSpace)
	if err != nil {
		utils.Debugf("container/v1: cpuacct.stat read failed: %v", err)
	}
	usage, tU, err := utils.FileInt(filepath.Join(memPath, "memory.usage_in_bytes"))
	if err != nil {
		utils.Debugf("container/v1: memory.usage_in_bytes read failed: %v", err)
	}
	maxMem, tM, err := utils.FileInt(filepath.Join(memPath, "memory.max_usage_in_bytes"))
	if err != nil {
		utils.Debugf("container/v1: memory.max_usage_in_bytes read failed: %v", err)
	}
	memStat, tMemStat, err := utils.FileKV(filepath.Join(memPath, "memory.stat"), utils.FieldSeparatorSpace)
	if err != nil {
		utils.Debugf("container/v1: memory.stat read failed: %v", err)
	}
	dr, dw, tBlk := getBlkioV1()
	sectorIO, tSector := getBlkioSectorsV1(blkioPath)
	numProcs, tProcs := getProcessCountV1(pidsPath)

	m.ContainerCPUTime = base.MetricInt{V: cpuUsage, T: tCpu}
	m.ContainerCPUTimeUserMode = base.MetricInt{V: utils.ParseInt64(cpuStat["user"]), T: tCpuStat}
	m.ContainerCPUTimeKernelMode = base.MetricInt{V: utils.ParseInt64(cpuStat["system"]), T: tCpuStat}
	m.ContainerMemoryUsed = base.MetricInt{V: usage, T: tU}
	m.ContainerMemoryMaxUsed = base.MetricInt{V: maxMem, T: tM}
	m.ContainerPgFault = base.MetricInt{V: utils.ParseInt64(memStat["pgfault"]), T: tMemStat}
	m.ContainerMajorPgFault = base.MetricInt{V: utils.ParseInt64(memStat["pgmajfault"]), T: tMemStat}
	m.ContainerDiskReadBytes = base.MetricInt{V: dr, T: tBlk}
	m.ContainerDiskWriteBytes = base.MetricInt{V: dw, T: tBlk}
	m.ContainerDiskSectorIO = base.MetricInt{V: sectorIO, T: tSector}
	m.ContainerNumProcesses = base.MetricInt{V: numProcs, T: tProcs}
}

func getBlkioV1() (r, w, ts int64) {
	path := filepath.Join(cgroupDir, "blkio", "blkio.throttle.io_service_bytes")
	lines, ts, _ := utils.FileLines(path)
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		val := utils.ParseInt64(f[2])
		if strings.EqualFold(f[1], "read") {
			r += val
		} else if strings.EqualFold(f[1], "write") {
			w += val
		}
	}
	return r, w, ts
}

func getBlkioSectorsV1(blkioPath string) (total, ts int64) {
	lines, ts, _ := utils.FileLines(filepath.Join(blkioPath, "blkio.sectors"))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) >= 2 {
			total += utils.ParseInt64(f[1])
		}
	}
	return total, ts
}

func getProcessCountV1(pidsPath string) (count, ts int64) {
	if count, ts, err := utils.FileInt(filepath.Join(pidsPath, "pids.current")); err == nil && count > 0 {
		return count, ts
	}
	lines, ts, _ := utils.FileLines(filepath.Join(cgroupDir, "cpuacct", "tasks"))
	return int64(len(lines)), ts
}
