package src

import (
	"InferenceProfiler/src"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/inference-profiler/utils"
)

// ProcessMetrics contains per-process measurements.
type ProcessMetrics struct {
	// Identification
	PID     int    `json:"pId"`
	Name    string `json:"pName"`
	Cmdline string `json:"pCmdline"`

	// Threading
	NumThreads exporter.Timed[int64] `json:"pNumThreads"`

	// CPU time (centiseconds)
	CPUUserMode        exporter.Timed[int64] `json:"pCpuTimeUserMode"`
	CPUKernelMode      exporter.Timed[int64] `json:"pCpuTimeKernelMode"`
	ChildrenUserMode   exporter.Timed[int64] `json:"pChildrenUserMode"`
	ChildrenKernelMode exporter.Timed[int64] `json:"pChildrenKernelMode"`

	// Context switches
	VoluntaryCtxSwitches    exporter.Timed[int64] `json:"pVoluntaryContextSwitches"`
	NonvoluntaryCtxSwitches exporter.Timed[int64] `json:"pNonvoluntaryContextSwitches"`

	// I/O delays (centiseconds)
	BlockIODelays exporter.Timed[int64] `json:"pBlockIODelays"`

	// Memory (bytes)
	VirtualMemory   exporter.Timed[int64] `json:"pVirtualMemoryBytes"`
	ResidentSetSize exporter.Timed[int64] `json:"pResidentSetSize"`
}

// ProcessCollection contains all process metrics plus summary.
type ProcessCollection struct {
	NumProcesses int              `json:"pNumProcesses"`
	Processes    []ProcessMetrics `json:"processes"`
}

// CollectProcesses gathers metrics for all running processes.
// This can be expensive on systems with many processes.
func CollectProcesses() ProcessCollection {
	pattern := "/proc/[0-9]*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return ProcessCollection{}
	}

	pageSize := int64(os.Getpagesize())
	var processes []ProcessMetrics

	for _, pidPath := range matches {
		pid, err := strconv.Atoi(filepath.Base(pidPath))
		if err != nil {
			continue
		}

		proc := collectProcess(pidPath, pid, pageSize)
		if proc != nil {
			processes = append(processes, *proc)
		}
	}

	return ProcessCollection{
		NumProcesses: len(processes),
		Processes:    processes,
	}
}

// collectProcess gathers metrics for a single process.
func collectProcess(pidPath string, pid int, pageSize int64) *ProcessMetrics {
	// Read /proc/[pid]/stat
	statData, statTS := utils.readFile(filepath.Join(pidPath, "stat"))
	if statData == "" {
		return nil
	}

	// Parse stat file - handle process names with parentheses
	// Format: pid (comm) state ppid pgrp session tty_nr tpgid flags ...
	rparenIdx := strings.LastIndex(statData, ")")
	if rparenIdx < 0 || rparenIdx+2 >= len(statData) {
		return nil
	}

	// Extract comm from between parentheses
	lparenIdx := strings.Index(statData, "(")
	var name string
	if lparenIdx >= 0 && rparenIdx > lparenIdx {
		name = statData[lparenIdx+1 : rparenIdx]
	}

	// Fields after ) - indexed from 0
	fields := strings.Fields(statData[rparenIdx+2:])
	if len(fields) < 40 {
		return nil
	}

	// Field indices (after comm):
	// 0: state, 1: ppid, 2: pgrp, ..., 11: utime, 12: stime, 13: cutime, 14: cstime
	// 17: num_threads, 20: vsize, 39: delayacct_blkio_ticks
	utime, _ := strconv.ParseInt(fields[11], 10, 64)
	stime, _ := strconv.ParseInt(fields[12], 10, 64)
	cutime, _ := strconv.ParseInt(fields[13], 10, 64)
	cstime, _ := strconv.ParseInt(fields[14], 10, 64)
	numThreads, _ := strconv.ParseInt(fields[17], 10, 64)
	vsize, _ := strconv.ParseInt(fields[20], 10, 64)
	blkioDelay, _ := strconv.ParseInt(fields[39], 10, 64)

	// Read /proc/[pid]/status for context switches
	status, statusTS := utils.parseKV(filepath.Join(pidPath, "status"), ':')
	volCtx, _ := strconv.ParseInt(status["voluntary_ctxt_switches"], 10, 64)
	nvolCtx, _ := strconv.ParseInt(status["nonvoluntary_ctxt_switches"], 10, 64)

	// Use name from status if available
	if statusName, ok := status["Name"]; ok && statusName != "" {
		name = statusName
	}

	// Read /proc/[pid]/statm for RSS
	statmData, statmTS := utils.readFile(filepath.Join(pidPath, "statm"))
	var rssPages int64
	if statmData != "" {
		statmFields := strings.Fields(statmData)
		if len(statmFields) >= 2 {
			rssPages, _ = strconv.ParseInt(statmFields[1], 10, 64)
		}
	}

	// Read /proc/[pid]/cmdline
	cmdline, _ := utils.readFile(filepath.Join(pidPath, "cmdline"))
	cmdline = strings.ReplaceAll(cmdline, "\x00", " ")
	cmdline = strings.TrimSpace(cmdline)

	return &ProcessMetrics{
		PID:                     pid,
		Name:                    name,
		Cmdline:                 cmdline,
		NumThreads:              exporter.TimedAt(numThreads, statTS),
		CPUUserMode:             exporter.TimedAt(utime*src.jiffiesPerSec, statTS),
		CPUKernelMode:           exporter.TimedAt(stime*src.jiffiesPerSec, statTS),
		ChildrenUserMode:        exporter.TimedAt(cutime*src.jiffiesPerSec, statTS),
		ChildrenKernelMode:      exporter.TimedAt(cstime*src.jiffiesPerSec, statTS),
		VoluntaryCtxSwitches:    exporter.TimedAt(volCtx, statusTS),
		NonvoluntaryCtxSwitches: exporter.TimedAt(nvolCtx, statusTS),
		BlockIODelays:           exporter.TimedAt(blkioDelay*src.jiffiesPerSec, statTS),
		VirtualMemory:           exporter.TimedAt(vsize, statTS),
		ResidentSetSize:         exporter.TimedAt(rssPages*pageSize, statmTS),
	}
}
