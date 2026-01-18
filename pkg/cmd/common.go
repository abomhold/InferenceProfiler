package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"context"
	"flag"
	"log"
	"time"
)

// CmdContext holds initialized command resources
type CmdContext struct {
	Manager *collecting.Manager
	Config  *utils.Config
}

// InitCmd initializes common command resources: parses flags, creates manager, collects static metrics
func InitCmd(name string, args []string) (*CmdContext, func()) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	cfg := utils.NewConfig()
	applyFlags := utils.GetFlags(fs, cfg)
	fs.Parse(args)
	applyFlags()

	manager := collecting.NewManager(cfg)

	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	ctx := &CmdContext{
		Manager: manager,
		Config:  cfg,
	}

	cleanup := func() {
		manager.Close()
	}

	return ctx, cleanup
}

// InitCmdWithSeparator initializes command with -- separator for sub-command args
// Returns CmdContext, cleanup func, and sub-command args after --
func InitCmdWithSeparator(name string, args []string) (*CmdContext, func(), []string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	cfg := utils.NewConfig()
	applyFlags := utils.GetFlags(fs, cfg)

	dashIdx := -1
	for i, arg := range args {
		if arg == utils.CMDSeparator {
			dashIdx = i
			break
		}
	}

	var flagArgs, cmdArgs []string
	if dashIdx >= 0 {
		flagArgs = args[:dashIdx]
		cmdArgs = args[dashIdx+1:]
	} else {
		flagArgs = args
	}

	fs.Parse(flagArgs)
	applyFlags()

	manager := collecting.NewManager(cfg)

	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	ctx := &CmdContext{
		Manager: manager,
		Config:  cfg,
	}

	cleanup := func() {
		manager.Close()
	}

	return ctx, cleanup, cmdArgs
}

// CollectSnapshot collects a single dynamic snapshot with appropriate flattening.
// flatten=true: expand all arrays (procs, gpu procs, gpu metrics) into top-level fields
// flatten=false: keep procs as JSON strings but still expand nvidia GPU metrics
func CollectSnapshot(manager *collecting.Manager, flatten bool) exporting.Record {
	dynamic := &collecting.DynamicMetrics{}
	record := manager.CollectDynamic(dynamic)
	if flatten {
		record = exporting.FlattenRecord(record)
	} else {
		// Even with --no-flatten, always flatten nvidia GPU metrics
		// Only processes become JSON strings
		record = exporting.FlattenRecordNoProcesses(record)
	}
	return record
}

// ComputeDelta takes initial and final snapshots and computes the delta
func ComputeDelta(initial, final exporting.Record, durationMs int64) exporting.Record {
	return exporting.DeltaRecord(initial, final, durationMs)
}

// DeltaCapture performs a complete delta capture: initial snapshot -> wait/context -> final snapshot -> delta
type DeltaCapture struct {
	InitialRecord exporting.Record
	FinalRecord   exporting.Record
	DeltaRecord   exporting.Record
	Duration      time.Duration
}

// RunDelta executes a delta capture waiting on context cancellation.
// flatten=true: expand all arrays into top-level prefixed fields
// flatten=false: keep processes as JSON, but still expand nvidia GPU metrics
func RunDelta(ctx context.Context, manager *collecting.Manager, flatten bool) *DeltaCapture {
	log.Println("Capturing initial snapshot...")
	initial := CollectSnapshot(manager, flatten)
	startTime := time.Now()

	<-ctx.Done()

	log.Println("Capturing final snapshot...")
	final := CollectSnapshot(manager, flatten)
	elapsed := time.Since(startTime)

	delta := ComputeDelta(initial, final, elapsed.Milliseconds())

	return &DeltaCapture{
		InitialRecord: initial,
		FinalRecord:   final,
		DeltaRecord:   delta,
		Duration:      elapsed,
	}
}

// RunDeltaWithDuration executes a delta capture for a specific duration.
// flatten=true: expand all arrays into top-level prefixed fields
// flatten=false: keep processes as JSON, but still expand nvidia GPU metrics
func RunDeltaWithDuration(manager *collecting.Manager, duration time.Duration, flatten bool) *DeltaCapture {
	log.Println("Capturing initial snapshot...")
	initial := CollectSnapshot(manager, flatten)
	startTime := time.Now()

	if duration > 0 {
		log.Printf("Waiting %v for final snapshot...", duration)
		time.Sleep(duration)
	}

	log.Println("Capturing final snapshot...")
	final := CollectSnapshot(manager, flatten)
	elapsed := time.Since(startTime)

	delta := ComputeDelta(initial, final, elapsed.Milliseconds())

	return &DeltaCapture{
		InitialRecord: initial,
		FinalRecord:   final,
		DeltaRecord:   delta,
		Duration:      elapsed,
	}
}

// StreamCollector handles streaming collection with periodic writes
type StreamCollector struct {
	manager  *collecting.Manager
	writer   exporting.Writer
	flatten  bool
	interval time.Duration
	count    int
}

// NewStreamCollector creates a new streaming collector
func NewStreamCollector(manager *collecting.Manager, format, outputFile string, flatten bool, intervalMs int) (*StreamCollector, error) {
	f, ok := exporting.Get(format)
	if !ok {
		return nil, nil
	}

	writer := f.Writer()
	if err := writer.Init(outputFile); err != nil {
		return nil, err
	}

	return &StreamCollector{
		manager:  manager,
		writer:   writer,
		flatten:  flatten,
		interval: time.Duration(intervalMs) * time.Millisecond,
	}, nil
}

// Run starts the streaming collection until context is cancelled
func (sc *StreamCollector) Run(ctx context.Context) int {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sc.writer.Flush()
			return sc.count

		case <-ticker.C:
			record := CollectSnapshot(sc.manager, sc.flatten)
			sc.writer.Write(record)
			sc.count++

			if sc.count%50 == 0 {
				sc.writer.Flush()
			}
		}
	}
}

// Close closes the underlying writer
func (sc *StreamCollector) Close() error {
	return sc.writer.Close()
}

// Count returns the number of records collected
func (sc *StreamCollector) Count() int {
	return sc.count
}

// BatchCollector handles batch collection (in-memory until write)
type BatchCollector struct {
	manager  *collecting.Manager
	records  []exporting.Record
	flatten  bool
	interval time.Duration
}

// NewBatchCollector creates a new batch collector
func NewBatchCollector(manager *collecting.Manager, flatten bool, intervalMs int) *BatchCollector {
	return &BatchCollector{
		manager:  manager,
		flatten:  flatten,
		interval: time.Duration(intervalMs) * time.Millisecond,
	}
}

// Run collects records until context is cancelled
func (bc *BatchCollector) Run(ctx context.Context) {
	ticker := time.NewTicker(bc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			dynamic := &collecting.DynamicMetrics{}
			record := bc.manager.CollectDynamic(dynamic)
			bc.records = append(bc.records, record)
		}
	}
}

// Records returns all collected records
func (bc *BatchCollector) Records() []exporting.Record {
	return bc.records
}

// Count returns the number of records collected
func (bc *BatchCollector) Count() int {
	return len(bc.records)
}

// SaveBatch saves batch records to file with appropriate flattening.
// flatten=true: expand all arrays into top-level prefixed fields
// flatten=false: keep processes as JSON, but still expand nvidia GPU metrics
func SaveBatch(outputFile string, records []exporting.Record, flatten bool) error {
	for i := range records {
		if flatten {
			records[i] = exporting.FlattenRecord(records[i])
		} else {
			records[i] = exporting.FlattenRecordNoProcesses(records[i])
		}
	}
	return exporting.SaveRecords(outputFile, records)
}

// SaveDelta saves a delta record to file
func SaveDelta(outputFile string, delta exporting.Record) error {
	return exporting.SaveRecords(outputFile, []exporting.Record{delta})
}

// GenerateOutputFilename generates a default output filename
func GenerateOutputFilename(cmdName, mode, format string) string {
	timestamp := time.Now().Format("20060102_150405")
	ext := exporting.GetExtension(format)
	if cmdName != "" {
		return mode + "_" + cmdName + "_" + timestamp + ext
	}
	return mode + "_" + timestamp + ext
}

// ShouldFlatten returns true if flattening should be applied (inverts DisableFlatten)
func ShouldFlatten(cfg *utils.Config) bool {
	return !cfg.DisableFlatten
}
