package collectors

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

type ProcessCollector struct{}

func NewProcessCollector() *ProcessCollector {
	return &ProcessCollector{}
}

func (c *ProcessCollector) Name() string { return "Process" }
func (c *ProcessCollector) Close() error { return nil }

func (c *ProcessCollector) CollectStatic(m *StaticMetrics) {}

func (c *ProcessCollector) CollectDynamic(m *DynamicMetrics) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}

	ts := probing.GetTimestamp()
	mult := int64(config.NanosecondsPerSec / config.JiffiesPerSecond)

	var processes []ProcessInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.ParseInt(entry.Name(), 10, 64)
		if err != nil {
			continue
		}

		procPath := filepath.Join("/proc", entry.Name())
		proc := ProcessInfo{PID: pid, PIDT: ts}

		if name, _, err := probing.File(filepath.Join(procPath, "comm")); err == nil {
			proc.Name, proc.NameT = name, ts
		}
		if cmdline, _, err := probing.File(filepath.Join(procPath, "cmdline")); err == nil {
			proc.Cmdline = strings.ReplaceAll(cmdline, "\x00", " ")
			proc.CmdlineT = ts
		}

		if stat, _, err := probing.File(filepath.Join(procPath, "stat")); err == nil {
			fields := strings.Fields(stat)
			if len(fields) >= 24 {
				proc.NumThreads = probing.ParseInt64(fields[19])
				proc.NumThreadsT = ts
				proc.CPUTimeUserMode = probing.ParseInt64(fields[13]) * mult
				proc.CPUTimeUserModeT = ts
				proc.CPUTimeKernelMode = probing.ParseInt64(fields[14]) * mult
				proc.CPUTimeKernelModeT = ts
				proc.ChildrenUserMode = probing.ParseInt64(fields[15]) * mult
				proc.ChildrenUserModeT = ts
				proc.ChildrenKernelMode = probing.ParseInt64(fields[16]) * mult
				proc.ChildrenKernelModeT = ts
				proc.VirtualMemoryBytes = probing.ParseInt64(fields[22])
				proc.VirtualMemoryBytesT = ts
				proc.ResidentSetSize = probing.ParseInt64(fields[23]) * 4096
				proc.ResidentSetSizeT = ts
			}
		}

		if status, _, err := probing.FileKV(filepath.Join(procPath, "status"), ":"); err == nil {
			proc.VoluntaryContextSwitches = probing.ParseInt64(status["voluntary_ctxt_switches"])
			proc.VoluntaryContextSwitchesT = ts
			proc.NonvoluntaryContextSwitches = probing.ParseInt64(status["nonvoluntary_ctxt_switches"])
			proc.NonvoluntaryContextSwitchesT = ts
		}

		if sched, _, err := probing.FileKV(filepath.Join(procPath, "sched"), ":"); err == nil {
			proc.BlockIODelays = int64(probing.ParseFloat64(sched["sum_sleep_runtime"]))
			proc.BlockIODelaysT = ts
		}

		processes = append(processes, proc)
	}

	m.Processes = processes
}
