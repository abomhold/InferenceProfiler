package collectors

import (
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// CollectProcessesDynamic populates per-process metrics into the slice
// Flattening to process{i}Field format happens at export time
func CollectProcessesDynamic(m *DynamicMetrics) {
	dirs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return
	}

	pageSize := int64(syscall.Getpagesize())

	for _, pidPath := range dirs {
		collectSingleProc(m, pidPath, pageSize)
	}
}

func collectSingleProc(m *DynamicMetrics, pidPath string, pageSize int64) {
	pid, err := strconv.ParseInt(filepath.Base(pidPath), 10, 64)
	if err != nil {
		return
	}

	statData, tStat := ProbeFile(filepath.Join(pidPath, "stat"))
	rparenIndex := strings.LastIndex(statData, ")")
	if rparenIndex == -1 || len(statData) < rparenIndex+3 {
		return
	}

	statParts := strings.Fields(statData[rparenIndex+2:])
	if len(statParts) < 40 {
		return
	}

	cmdline, tCmd := ProbeFile(filepath.Join(pidPath, "cmdline"))
	status, tStatus := ProbeFileKV(filepath.Join(pidPath, "status"), ":")
	statm, tStatm := ProbeFile(filepath.Join(pidPath, "statm"))

	var rssBytes int64
	if parts := strings.Fields(statm); len(parts) >= 2 {
		rssPages, _ := strconv.ParseInt(parts[1], 10, 64)
		rssBytes = rssPages * pageSize
	}

	pName := status["Name"]
	if pName == "" {
		if lp := strings.Index(statData, "("); lp != -1 && rparenIndex > lp {
			pName = statData[lp+1 : rparenIndex]
		}
	}

	proc := ProcessMetrics{
		PID:                          pid,
		PIDT:                         tStat,
		Name:                         pName,
		NameT:                        tStat,
		Cmdline:                      strings.ReplaceAll(cmdline, "\x00", " "),
		CmdlineT:                     tCmd,
		NumThreads:                   parseInt64(statParts[17]),
		NumThreadsT:                  tStat,
		CPUTimeUserMode:              parseInt64(statParts[11]) * JiffiesPerSecond,
		CPUTimeUserModeT:             tStat,
		CPUTimeKernelMode:            parseInt64(statParts[12]) * JiffiesPerSecond,
		CPUTimeKernelModeT:           tStat,
		ChildrenUserMode:             parseInt64(statParts[13]) * JiffiesPerSecond,
		ChildrenUserModeT:            tStat,
		ChildrenKernelMode:           parseInt64(statParts[14]) * JiffiesPerSecond,
		ChildrenKernelModeT:          tStat,
		VoluntaryContextSwitches:     parseInt64(status["voluntary_ctxt_switches"]),
		VoluntaryContextSwitchesT:    tStatus,
		NonvoluntaryContextSwitches:  parseInt64(status["nonvoluntary_ctxt_switches"]),
		NonvoluntaryContextSwitchesT: tStatus,
		BlockIODelays:                parseInt64(statParts[39]) * JiffiesPerSecond,
		BlockIODelaysT:               tStat,
		VirtualMemoryBytes:           parseInt64(statParts[20]),
		VirtualMemoryBytesT:          tStat,
		ResidentSetSize:              rssBytes,
		ResidentSetSizeT:             tStatm,
	}

	m.Processes = append(m.Processes, proc)
}
