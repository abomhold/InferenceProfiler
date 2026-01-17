package collectors

import (
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

// Collector collects process metrics.
type Collector struct {
	pageSize int64
}

// New creates a new Process collector.
func New() *Collector {
	return &Collector{
		pageSize: int64(syscall.Getpagesize()),
	}
}

// Name returns the collector name.
func (c *Collector) Name() string {
	return "Process"
}

// Close releases any resources.
func (c *Collector) Close() error {
	return nil
}

// CollectStatic returns nil as process collector has no static data.
func (c *Collector) CollectStatic() types.Record {
	return nil
}

// CollectDynamic collects dynamic process metrics.
// Stores []Info under types.KeyProcesses for deferred serialization.
func (c *Collector) CollectDynamic() types.Record {
	dirs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return nil
	}

	// Pre-allocate with estimated capacity
	procs := make([]Info, 0, len(dirs))
	for _, pidPath := range dirs {
		if proc := c.collectSingleProcess(pidPath); proc != nil {
			procs = append(procs, *proc)
		}
	}

	if len(procs) == 0 {
		return nil
	}

	// Store slice directly - serialization deferred to export time
	return types.Record{types.KeyProcesses: procs}
}

func (c *Collector) collectSingleProcess(pidPath string) *Info {
	pid, err := strconv.ParseInt(filepath.Base(pidPath), 10, 64)
	if err != nil {
		return nil
	}

	statVal, tStat := probing.File(filepath.Join(pidPath, "stat"))
	statData := statVal

	rparenIndex := strings.LastIndex(statData, ")")
	if rparenIndex == -1 || len(statData) < rparenIndex+3 {
		return nil
	}

	statParts := strings.Fields(statData[rparenIndex+2:])
	if len(statParts) < 40 {
		return nil
	}

	cmdlineVal, tCmd := probing.File(filepath.Join(pidPath, "cmdline"))
	status, tStatus := probing.FileKV(filepath.Join(pidPath, "status"), ":")
	statmVal, tStatm := probing.File(filepath.Join(pidPath, "statm"))

	var rssBytes int64
	if parts := strings.Fields(statmVal); len(parts) >= 2 {
		rssPages := probing.ParseInt64(parts[1])
		rssBytes = rssPages * c.pageSize
	}

	pName := status["Name"]
	if pName == "" {
		if lp := strings.Index(statData, "("); lp != -1 && rparenIndex > lp {
			pName = statData[lp+1 : rparenIndex]
		}
	}

	return &Info{
		PID:                          pid,
		PIDT:                         tStat,
		Name:                         pName,
		NameT:                        tStat,
		Cmdline:                      strings.ReplaceAll(cmdlineVal, "\x00", " "),
		CmdlineT:                     tCmd,
		NumThreads:                   probing.ParseInt64(statParts[17]),
		NumThreadsT:                  tStat,
		CPUTimeUserMode:              probing.ParseInt64(statParts[11]) * config.JiffiesPerSecond,
		CPUTimeUserModeT:             tStat,
		CPUTimeKernelMode:            probing.ParseInt64(statParts[12]) * config.JiffiesPerSecond,
		CPUTimeKernelModeT:           tStat,
		ChildrenUserMode:             probing.ParseInt64(statParts[13]) * config.JiffiesPerSecond,
		ChildrenUserModeT:            tStat,
		ChildrenKernelMode:           probing.ParseInt64(statParts[14]) * config.JiffiesPerSecond,
		ChildrenKernelModeT:          tStat,
		VoluntaryContextSwitches:     probing.ParseInt64(status["voluntary_ctxt_switches"]),
		VoluntaryContextSwitchesT:    tStatus,
		NonvoluntaryContextSwitches:  probing.ParseInt64(status["nonvoluntary_ctxt_switches"]),
		NonvoluntaryContextSwitchesT: tStatus,
		BlockIODelays:                probing.ParseInt64(statParts[39]) * config.JiffiesPerSecond,
		BlockIODelaysT:               tStat,
		VirtualMemoryBytes:           probing.ParseInt64(statParts[20]),
		VirtualMemoryBytesT:          tStat,
		ResidentSetSize:              rssBytes,
		ResidentSetSizeT:             tStatm,
	}
}
