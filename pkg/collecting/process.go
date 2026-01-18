package collecting

import (
	"InferenceProfiler/pkg/utils"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessCollector struct {
	concurrent      bool
	collectIODelays bool // sched file is expensive, make optional
}

func NewProcessCollector(concurrent bool) *ProcessCollector {
	return &ProcessCollector{
		concurrent: concurrent,
	}
}

func (c *ProcessCollector) Name() string { return "Process" }
func (c *ProcessCollector) Close() error { return nil }

func (c *ProcessCollector) CollectStatic(m *StaticMetrics) {}

// Buffer pool to avoid allocations on every file read
var procBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 4096)
		return &buf
	},
}

func (c *ProcessCollector) CollectDynamic(m *DynamicMetrics) {
	if c.concurrent {
		c.collectConcurrent(m)
	} else {
		c.collectSequential(m)
	}
}

func (c *ProcessCollector) collectSequential(m *DynamicMetrics) {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return
	}
	processes := make([]ProcessInfo, 0, len(entries)/2)
	ts := utils.GetTimestamp()
	mult := int64(time.Nanosecond / jiffiesPerSecond)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.ParseInt(entry.Name(), 10, 64)
		if err != nil {
			continue
		}

		if proc, ok := c.collectProcess(entry.Name(), pid, ts, mult); ok {
			processes = append(processes, proc)
		}
	}

	m.Processes = processes
}

func (c *ProcessCollector) collectConcurrent(m *DynamicMetrics) {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return
	}

	type pidEntry struct {
		name string
		pid  int64
	}
	pids := make([]pidEntry, 0, len(entries)/2)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if pid, err := strconv.ParseInt(entry.Name(), 10, 64); err == nil {
			pids = append(pids, pidEntry{entry.Name(), pid})
		}
	}

	numWorkers := runtime.NumCPU()
	pidChan := make(chan pidEntry, len(pids))
	resultChan := make(chan ProcessInfo, len(pids))

	var wg sync.WaitGroup
	ts := time.Now().UnixNano()
	mult := int64(time.Nanosecond / jiffiesPerSecond)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pe := range pidChan {
				if proc, ok := c.collectProcess(pe.name, pe.pid, ts, mult); ok {
					resultChan <- proc
				}
			}
		}()
	}

	go func() {
		for _, pe := range pids {
			pidChan <- pe
		}
		close(pidChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	processes := make([]ProcessInfo, 0, len(pids))
	for proc := range resultChan {
		processes = append(processes, proc)
	}

	m.Processes = processes
}

func (c *ProcessCollector) collectProcess(pidStr string, pid int64, ts int64, mult int64) (ProcessInfo, bool) {
	procPath := filepath.Join(procDir, pidStr)
	proc := ProcessInfo{PID: pid, PIDT: ts}

	if data, ok := readProcFile(filepath.Join(procPath, "stat")); ok {
		c.parseStat(data, &proc, ts, mult)
	}

	if data, ok := readProcFile(filepath.Join(procPath, "cmdline")); ok {
		proc.Cmdline = strings.ReplaceAll(string(data), "\x00", " ")
		proc.CmdlineT = ts
	}

	if data, ok := readProcFile(filepath.Join(procPath, "status")); ok {
		proc.VoluntaryContextSwitches, proc.NonvoluntaryContextSwitches = parseStatusCtxSwitches(data)
		proc.VoluntaryContextSwitchesT = ts
		proc.NonvoluntaryContextSwitchesT = ts
	}

	return proc, true
}

// readProcFile reads a proc file using pooled buffer
func readProcFile(path string) ([]byte, bool) {
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return nil, false
	}
	defer syscall.Close(fd)

	bufPtr := procBufPool.Get().(*[]byte)
	defer procBufPool.Put(bufPtr)

	n, err := syscall.Read(fd, *bufPtr)
	if err != nil || n == 0 {
		return nil, false
	}
	result := make([]byte, n)
	copy(result, (*bufPtr)[:n])
	return result, true
}

// parseStat extracts fields from /proc/[pid]/stat
func (c *ProcessCollector) parseStat(data []byte, proc *ProcessInfo, ts int64, mult int64) {
	start := bytes.IndexByte(data, '(')
	end := bytes.LastIndexByte(data, ')')
	if start == -1 || end == -1 || end <= start {
		return
	}
	proc.Name = string(data[start+1 : end])
	proc.NameT = ts
	rest := data[end+2:]
	fields := bytes.Fields(rest)

	if len(fields) < 22 { // need up to field 24 (index 21 after comm)
		return
	}
	proc.CPUTimeUserMode = utils.ToInt64(fields[11]) * mult
	proc.CPUTimeUserModeT = ts
	proc.CPUTimeKernelMode = utils.ToInt64(fields[12]) * mult
	proc.CPUTimeKernelModeT = ts
	proc.ChildrenUserMode = utils.ToInt64(fields[13]) * mult
	proc.ChildrenUserModeT = ts
	proc.ChildrenKernelMode = utils.ToInt64(fields[14]) * mult
	proc.ChildrenKernelModeT = ts
	proc.NumThreads = utils.ToInt64(fields[17])
	proc.NumThreadsT = ts
	proc.VirtualMemoryBytes = utils.ToInt64(fields[20])
	proc.VirtualMemoryBytesT = ts
	proc.ResidentSetSize = utils.ToInt64(fields[21]) * pageSize
	proc.ResidentSetSizeT = ts
}

// parseStatusCtxSwitches extracts only the two context switch fields
func parseStatusCtxSwitches(data []byte) (voluntary, nonvoluntary int64) {
	for len(data) > 0 {
		lineEnd := bytes.IndexByte(data, '\n')
		if lineEnd == -1 {
			lineEnd = len(data)
		}
		line := data[:lineEnd]

		if bytes.HasPrefix(line, []byte("voluntary_ctxt_switches:")) {
			voluntary = utils.ToInt64(bytes.TrimSpace(line[24:]))
		} else if bytes.HasPrefix(line, []byte("nonvoluntary_ctxt_switches:")) {
			nonvoluntary = utils.ToInt64(bytes.TrimSpace(line[27:]))
			return // early return
		}

		if lineEnd+1 >= len(data) {
			break
		}
		data = data[lineEnd+1:]
	}
	return
}
