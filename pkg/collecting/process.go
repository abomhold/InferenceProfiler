package collecting

import (
	"InferenceProfiler/pkg/metrics"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"InferenceProfiler/pkg/probing"
)

//const JiffiesPerSecond = 100

type Process struct {
	pageSize int64
}

func NewProcess() *Process {
	return &Process{
		pageSize: int64(syscall.Getpagesize()),
	}
}

func (c *Process) Name() string       { return "Process" }
func (c *Process) Close() error       { return nil }
func (c *Process) CollectStatic() any { return nil }

func (c *Process) CollectDynamic() any {
	m := &metrics.ProcessDynamic{}

	dirs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return m
	}

	var procs []metrics.Process
	for _, pidPath := range dirs {
		if proc := c.collectSingleProcess(pidPath); proc != nil {
			procs = append(procs, *proc)
		}
	}

	if len(procs) > 0 {
		data, _ := json.Marshal(procs)
		m.ProcessesJSON = string(data)
	}

	return m
}

func (c *Process) collectSingleProcess(pidPath string) *metrics.Process {
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

	return &metrics.Process{
		PID:                          pid,
		PIDT:                         tStat,
		Name:                         pName,
		NameT:                        tStat,
		Cmdline:                      strings.ReplaceAll(cmdlineVal, "\x00", " "),
		CmdlineT:                     tCmd,
		NumThreads:                   probing.ParseInt64(statParts[17]),
		NumThreadsT:                  tStat,
		CPUTimeUserMode:              probing.ParseInt64(statParts[11]) * JiffiesPerSecond,
		CPUTimeUserModeT:             tStat,
		CPUTimeKernelMode:            probing.ParseInt64(statParts[12]) * JiffiesPerSecond,
		CPUTimeKernelModeT:           tStat,
		ChildrenUserMode:             probing.ParseInt64(statParts[13]) * JiffiesPerSecond,
		ChildrenUserModeT:            tStat,
		ChildrenKernelMode:           probing.ParseInt64(statParts[14]) * JiffiesPerSecond,
		ChildrenKernelModeT:          tStat,
		VoluntaryContextSwitches:     probing.ParseInt64(status["voluntary_ctxt_switches"]),
		VoluntaryContextSwitchesT:    tStatus,
		NonvoluntaryContextSwitches:  probing.ParseInt64(status["nonvoluntary_ctxt_switches"]),
		NonvoluntaryContextSwitchesT: tStatus,
		BlockIODelays:                probing.ParseInt64(statParts[39]) * JiffiesPerSecond,
		BlockIODelaysT:               tStat,
		VirtualMemoryBytes:           probing.ParseInt64(statParts[20]),
		VirtualMemoryBytesT:          tStat,
		ResidentSetSize:              rssBytes,
		ResidentSetSizeT:             tStatm,
	}
}
