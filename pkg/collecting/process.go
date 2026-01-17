package collecting

import (
	"InferenceProfiler/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type ProcessCollector struct {
	concurrent bool
}

func NewProcessCollector(concurrent bool) *ProcessCollector {
	return &ProcessCollector{concurrent: concurrent}
}

func (c *ProcessCollector) Name() string { return "Process" }
func (c *ProcessCollector) Close() error { return nil }

func (c *ProcessCollector) CollectStatic(m *StaticMetrics) {}

func (c *ProcessCollector) CollectDynamic(m *DynamicMetrics) {
	if c.concurrent {
		c.collectConcurrent(m)
	} else {
		c.collectSequential(m)
	}
}

func (c *ProcessCollector) collectSequential(m *DynamicMetrics) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}

	ts := utils.GetTimestamp()
	mult := int64(utils.NanosecondsPerSec / utils.JiffiesPerSecond)

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

		if name, _, err := utils.File(filepath.Join(procPath, "comm")); err == nil {
			proc.Name, proc.NameT = name, ts
		}
		if cmdline, _, err := utils.File(filepath.Join(procPath, "cmdline")); err == nil {
			proc.Cmdline = strings.ReplaceAll(cmdline, "\x00", " ")
			proc.CmdlineT = ts
		}

		if stat, _, err := utils.File(filepath.Join(procPath, "stat")); err == nil {
			fields := strings.Fields(stat)
			if len(fields) >= 24 {
				proc.NumThreads = utils.ParseInt64(fields[19])
				proc.NumThreadsT = ts
				proc.CPUTimeUserMode = utils.ParseInt64(fields[13]) * mult
				proc.CPUTimeUserModeT = ts
				proc.CPUTimeKernelMode = utils.ParseInt64(fields[14]) * mult
				proc.CPUTimeKernelModeT = ts
				proc.ChildrenUserMode = utils.ParseInt64(fields[15]) * mult
				proc.ChildrenUserModeT = ts
				proc.ChildrenKernelMode = utils.ParseInt64(fields[16]) * mult
				proc.ChildrenKernelModeT = ts
				proc.VirtualMemoryBytes = utils.ParseInt64(fields[22])
				proc.VirtualMemoryBytesT = ts
				proc.ResidentSetSize = utils.ParseInt64(fields[23]) * 4096
				proc.ResidentSetSizeT = ts
			}
		}

		if status, _, err := utils.FileKV(filepath.Join(procPath, "status"), ":"); err == nil {
			proc.VoluntaryContextSwitches = utils.ParseInt64(status["voluntary_ctxt_switches"])
			proc.VoluntaryContextSwitchesT = ts
			proc.NonvoluntaryContextSwitches = utils.ParseInt64(status["nonvoluntary_ctxt_switches"])
			proc.NonvoluntaryContextSwitchesT = ts
		}

		if sched, _, err := utils.FileKV(filepath.Join(procPath, "sched"), ":"); err == nil {
			proc.BlockIODelays = int64(utils.ParseFloat64(sched["sum_sleep_runtime"]))
			proc.BlockIODelaysT = ts
		}

		processes = append(processes, proc)
	}

	m.Processes = processes
}

func (c *ProcessCollector) collectConcurrent(m *DynamicMetrics) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}

	// Pre-filter to get only numeric entries (PIDs)
	var pids []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.ParseInt(entry.Name(), 10, 64); err == nil {
			pids = append(pids, entry.Name())
		}
	}

	// Use worker pool pattern
	numWorkers := runtime.NumCPU()
	pidChan := make(chan string, len(pids))
	resultChan := make(chan ProcessInfo, len(pids))

	var wg sync.WaitGroup

	ts := utils.GetTimestamp()
	mult := int64(utils.NanosecondsPerSec / utils.JiffiesPerSecond)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pidStr := range pidChan {
				if proc, ok := c.collectProcess(pidStr, ts, mult); ok {
					resultChan <- proc
				}
			}
		}()
	}

	// Send PIDs to workers
	go func() {
		for _, pid := range pids {
			pidChan <- pid
		}
		close(pidChan)
	}()

	// Wait for all workers and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var processes []ProcessInfo
	for proc := range resultChan {
		processes = append(processes, proc)
	}

	m.Processes = processes
}

func (c *ProcessCollector) collectProcess(pidStr string, ts int64, mult int64) (ProcessInfo, bool) {
	pid, _ := strconv.ParseInt(pidStr, 10, 64)
	procPath := filepath.Join("/proc", pidStr)
	proc := ProcessInfo{PID: pid, PIDT: ts}

	if name, _, err := utils.File(filepath.Join(procPath, "comm")); err == nil {
		proc.Name, proc.NameT = name, ts
	}
	if cmdline, _, err := utils.File(filepath.Join(procPath, "cmdline")); err == nil {
		proc.Cmdline = strings.ReplaceAll(cmdline, "\x00", " ")
		proc.CmdlineT = ts
	}

	if stat, _, err := utils.File(filepath.Join(procPath, "stat")); err == nil {
		fields := strings.Fields(stat)
		if len(fields) >= 24 {
			proc.NumThreads = utils.ParseInt64(fields[19])
			proc.NumThreadsT = ts
			proc.CPUTimeUserMode = utils.ParseInt64(fields[13]) * mult
			proc.CPUTimeUserModeT = ts
			proc.CPUTimeKernelMode = utils.ParseInt64(fields[14]) * mult
			proc.CPUTimeKernelModeT = ts
			proc.ChildrenUserMode = utils.ParseInt64(fields[15]) * mult
			proc.ChildrenUserModeT = ts
			proc.ChildrenKernelMode = utils.ParseInt64(fields[16]) * mult
			proc.ChildrenKernelModeT = ts
			proc.VirtualMemoryBytes = utils.ParseInt64(fields[22])
			proc.VirtualMemoryBytesT = ts
			proc.ResidentSetSize = utils.ParseInt64(fields[23]) * 4096
			proc.ResidentSetSizeT = ts
		}
	}

	if status, _, err := utils.FileKV(filepath.Join(procPath, "status"), ":"); err == nil {
		proc.VoluntaryContextSwitches = utils.ParseInt64(status["voluntary_ctxt_switches"])
		proc.VoluntaryContextSwitchesT = ts
		proc.NonvoluntaryContextSwitches = utils.ParseInt64(status["nonvoluntary_ctxt_switches"])
		proc.NonvoluntaryContextSwitchesT = ts
	}

	if sched, _, err := utils.FileKV(filepath.Join(procPath, "sched"), ":"); err == nil {
		proc.BlockIODelays = int64(utils.ParseFloat64(sched["sum_sleep_runtime"]))
		proc.BlockIODelaysT = ts
	}

	return proc, true
}
