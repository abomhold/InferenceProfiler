package collectors

import (
	"InferenceProfiler/src/collectors/types"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ProcessCollector collects per-process metrics from /proc/[pid]/*
type ProcessCollector struct {
	BaseCollector
	pageSize int64
}

// NewProcessCollector creates a new process collector
func NewProcessCollector() *ProcessCollector {
	return &ProcessCollector{
		pageSize: int64(syscall.Getpagesize()),
	}
}

func (c *ProcessCollector) Name() string {
	return "Process"
}

func (c *ProcessCollector) CollectStatic(m *types.StaticMetrics) {
	// Process collector has no static metrics
}

func (c *ProcessCollector) CollectDynamic(m *types.DynamicMetrics) {
	dirs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return
	}

	for _, pidPath := range dirs {
		if proc := c.collectSingleProcess(pidPath); proc != nil {
			m.Processes = append(m.Processes, *proc)
		}
	}
}

func (c *ProcessCollector) collectSingleProcess(pidPath string) *types.ProcessMetrics {
	pid, err := strconv.ParseInt(filepath.Base(pidPath), 10, 64)
	if err != nil {
		return nil
	}

	statData, tStat := ProbeFile(filepath.Join(pidPath, "stat"))
	rparenIndex := strings.LastIndex(statData, ")")
	if rparenIndex == -1 || len(statData) < rparenIndex+3 {
		return nil
	}

	statParts := strings.Fields(statData[rparenIndex+2:])
	if len(statParts) < 40 {
		return nil
	}

	cmdline, tCmd := ProbeFile(filepath.Join(pidPath, "cmdline"))
	status, tStatus := ProbeFileKV(filepath.Join(pidPath, "status"), ":")
	statm, tStatm := ProbeFile(filepath.Join(pidPath, "statm"))

	var rssBytes int64
	if parts := strings.Fields(statm); len(parts) >= 2 {
		rssPages, _ := strconv.ParseInt(parts[1], 10, 64)
		rssBytes = rssPages * c.pageSize
	}

	pName := status["Name"]
	if pName == "" {
		if lp := strings.Index(statData, "("); lp != -1 && rparenIndex > lp {
			pName = statData[lp+1 : rparenIndex]
		}
	}

	return &types.ProcessMetrics{
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
}
