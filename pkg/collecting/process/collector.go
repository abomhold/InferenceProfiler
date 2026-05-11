package process

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const procDir = "/proc"

type Dynamic struct {
	PID                         int64          `json:"Id"`
	Name                        base.MetricStr `json:"Name"`
	CmdLine                     base.MetricStr `json:"CmdLine"`
	NumThreads                  base.MetricInt `json:"NumThreads"`
	CPUTimeUserMode             base.MetricInt `json:"CpuTimeUserMode"`
	CPUTimeKernelMode           base.MetricInt `json:"CpuTimeKernelMode"`
	ChildrenUserMode            base.MetricInt `json:"ChildrenUserMode"`
	ChildrenKernelMode          base.MetricInt `json:"ChildrenKernelMode"`
	VoluntaryContextSwitches    base.MetricInt `json:"VoluntaryContextSwitches"`
	NonvoluntaryContextSwitches base.MetricInt `json:"NonvoluntaryContextSwitches"`
	BlockIODelays               base.MetricInt `json:"BlockIODelays"`
	VirtualMemoryBytes          base.MetricInt `json:"VirtualMemoryBytes"`
	ResidentSetSize             base.MetricInt `json:"ResidentSetSize"`
}

type Collector struct{}

func New() *Collector { return &Collector{} }

func (c *Collector) Name() string               { return "Process" }
func (c *Collector) Init(_ *utils.Config) error { return nil }
func (c *Collector) Static() any                { return nil }
func (c *Collector) Close() error               { return nil }

func (c *Collector) Poll(_ context.Context) any {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}

	processes := make([]Dynamic, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.ParseInt(entry.Name(), 10, 64)
		if err != nil {
			continue
		}
		if proc, ok := parseProcess(entry.Name(), pid); ok {
			processes = append(processes, proc)
		}
	}
	return processes
}

func parseProcess(pidStr string, pid int64) (Dynamic, bool) {
	procPath := filepath.Join(procDir, pidStr)
	proc := Dynamic{PID: pid}

	if data, ts, err := utils.File(filepath.Join(procPath, "stat")); err == nil {
		parseStat([]byte(data), &proc, ts)
	} else {
		return proc, false
	}

	if data, ts, err := utils.File(filepath.Join(procPath, "cmdline")); err == nil {
		proc.CmdLine = base.MetricStr{V: strings.ReplaceAll(data, "\x00", " "), T: ts}
	}

	if data, ts, err := utils.File(filepath.Join(procPath, "status")); err == nil {
		volContextSwitches, nonvolContextSwitches := parseStatusCtxSwitches([]byte(data))
		proc.VoluntaryContextSwitches = base.MetricInt{V: volContextSwitches, T: ts}
		proc.NonvoluntaryContextSwitches = base.MetricInt{V: nonvolContextSwitches, T: ts}
	}

	return proc, true
}

func parseStat(data []byte, proc *Dynamic, ts int64) {
	start := bytes.IndexByte(data, '(')
	end := bytes.LastIndexByte(data, ')')
	if start == -1 || end == -1 || end <= start {
		return
	}

	proc.Name = base.MetricStr{V: string(data[start+1 : end]), T: ts}

	rest := data[end+2:]
	fields := bytes.Fields(rest)
	if len(fields) < 40 {
		return
	}

	proc.CPUTimeUserMode = base.MetricInt{V: utils.ParseInt64Bytes(fields[11]), T: ts}
	proc.CPUTimeKernelMode = base.MetricInt{V: utils.ParseInt64Bytes(fields[12]), T: ts}
	proc.ChildrenUserMode = base.MetricInt{V: utils.ParseInt64Bytes(fields[13]), T: ts}
	proc.ChildrenKernelMode = base.MetricInt{V: utils.ParseInt64Bytes(fields[14]), T: ts}
	proc.NumThreads = base.MetricInt{V: utils.ParseInt64Bytes(fields[17]), T: ts}
	proc.VirtualMemoryBytes = base.MetricInt{V: utils.ParseInt64Bytes(fields[20]), T: ts}
	proc.ResidentSetSize = base.MetricInt{V: utils.ParseInt64Bytes(fields[21]), T: ts}
	proc.BlockIODelays = base.MetricInt{V: utils.ParseInt64Bytes(fields[39]), T: ts}
}

func parseStatusCtxSwitches(data []byte) (voluntary, nonvoluntary int64) {
	for len(data) > 0 {
		lineEnd := bytes.IndexByte(data, '\n')
		if lineEnd == -1 {
			lineEnd = len(data)
		}
		line := data[:lineEnd]

		if bytes.HasPrefix(line, []byte("voluntary_ctxt_switches:")) {
			voluntary = utils.ParseInt64Bytes(bytes.TrimSpace(line[24:]))
		} else if bytes.HasPrefix(line, []byte("nonvoluntary_ctxt_switches:")) {
			nonvoluntary = utils.ParseInt64Bytes(bytes.TrimSpace(line[27:]))
			return
		}

		if lineEnd+1 >= len(data) {
			break
		}
		data = data[lineEnd+1:]
	}
	return
}
