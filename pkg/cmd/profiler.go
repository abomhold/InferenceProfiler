package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Profiler(args []string) {
	fs := flag.NewFlagSet("profiler", flag.ExitOnError)
	cfg := utils.NewConfig()
	applyFlags := utils.GetFlags(fs, cfg)
	fs.Parse(args)
	applyFlags()

	if !cfg.Batch && !cfg.Stream {
		cfg.Stream = true
	}

	if cfg.OutputFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		ext := exporting.GetExtension(cfg.Format)
		mode := "profiler"
		if cfg.Delta {
			mode = "delta"
		}
		cfg.OutputFile = fmt.Sprintf("%s_%s%s", mode, timestamp, ext)
	}

	manager := collecting.NewManager(cfg)
	defer manager.Close()

	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	// Create base context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// If duration is set, create a timeout context
	if cfg.Duration > 0 {
		var timeoutCancel context.CancelFunc
		ctx, timeoutCancel = context.WithTimeout(ctx, time.Duration(cfg.Duration)*time.Millisecond)
		defer timeoutCancel()
	}

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Delta mode: take initial snapshot, wait, take final snapshot, compute diff
	if cfg.Delta {
		runDeltaMode(ctx, manager, cfg)
		return
	}

	// Regular continuous profiling (with optional duration limit)
	if cfg.Batch {
		runBatchMode(ctx, manager, cfg)
	} else {
		runStreamMode(ctx, manager, cfg)
	}
}

func runDeltaMode(ctx context.Context, manager *collecting.Manager, cfg *utils.Config) {
	log.Printf("Delta profiler started")
	log.Printf("  Output: %s", cfg.OutputFile)
	log.Printf("  Format: %s", cfg.Format)
	if cfg.Duration > 0 {
		log.Printf("  Duration: %dms", cfg.Duration)
	} else {
		log.Printf("  Duration: until Ctrl+C")
	}

	// Take initial snapshot
	log.Println("Capturing initial snapshot...")
	initialDynamic := &collecting.DynamicMetrics{}
	initialRecord := manager.CollectDynamic(initialDynamic)
	if !cfg.DisableFlatten {
		initialRecord = exporting.FlattenRecord(initialRecord)
	}
	startTime := time.Now()

	// Wait for context (either timeout from duration or signal)
	if cfg.Duration > 0 {
		log.Printf("Waiting %v for final snapshot...", time.Duration(cfg.Duration)*time.Millisecond)
	} else {
		log.Println("Waiting for Ctrl+C to capture final snapshot...")
	}
	<-ctx.Done()

	// Take final snapshot
	log.Println("Capturing final snapshot...")
	finalDynamic := &collecting.DynamicMetrics{}
	finalRecord := manager.CollectDynamic(finalDynamic)
	if !cfg.DisableFlatten {
		finalRecord = exporting.FlattenRecord(finalRecord)
	}
	elapsed := time.Since(startTime)

	// Calculate delta
	deltaRecord := exporting.DeltaRecord(initialRecord, finalRecord, elapsed.Milliseconds())

	// Write output
	log.Printf("Writing delta record to %s...", cfg.OutputFile)
	if err := exporting.SaveRecords(cfg.OutputFile, []exporting.Record{deltaRecord}, nil); err != nil {
		log.Fatalf("Failed to write delta record: %v", err)
	}

	log.Printf("Delta profiling completed in %v", elapsed)

	if cfg.Graphs {
		generateGraphs(cfg.OutputFile, cfg.GraphDir)
	}
}

func runStreamMode(ctx context.Context, manager *collecting.Manager, cfg *utils.Config) {
	f, ok := exporting.Get(cfg.Format)
	if !ok {
		log.Fatalf("Unsupported format: %s", cfg.Format)
	}

	writer := f.Writer()
	if err := writer.Init(cfg.OutputFile, nil); err != nil {
		log.Fatalf("Failed to initialize writer: %v", err)
	}
	defer writer.Close()

	log.Printf("Profiler started (stream mode)")
	log.Printf("  Output: %s", cfg.OutputFile)
	log.Printf("  Format: %s", cfg.Format)
	log.Printf("  Interval: %dms", cfg.Interval)
	if cfg.Duration > 0 {
		log.Printf("  Duration: %dms (auto-stop)", cfg.Duration)
	} else {
		log.Printf("  Duration: until Ctrl+C")
	}

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Millisecond)
	defer ticker.Stop()

	count := 0
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			writer.Flush()
			elapsed := time.Since(start)
			log.Printf("Collected %d records in %v (%.2f records/sec)",
				count, elapsed, float64(count)/elapsed.Seconds())

			if cfg.Graphs {
				generateGraphs(cfg.OutputFile, cfg.GraphDir)
			}
			return

		case <-ticker.C:
			dynamic := &collecting.DynamicMetrics{}
			record := manager.CollectDynamic(dynamic)

			// Check DisableFlatten (inverted)
			if !cfg.DisableFlatten {
				record = exporting.FlattenRecord(record)
			}

			if err := writer.Write(record); err != nil {
				log.Printf("Write error: %v", err)
				continue
			}

			count++
			if count%100 == 0 {
				writer.Flush()
				log.Printf("Progress: %d records", count)
			}
		}
	}
}

func runBatchMode(ctx context.Context, manager *collecting.Manager, cfg *utils.Config) {
	log.Printf("Profiler started (batch mode)")
	log.Printf("  Interval: %dms", cfg.Interval)
	if cfg.Duration > 0 {
		log.Printf("  Duration: %dms (auto-stop)", cfg.Duration)
	} else {
		log.Printf("  Duration: until Ctrl+C")
	}
	log.Println("  Collecting in memory, will write at end...")

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Millisecond)
	defer ticker.Stop()

	var records []exporting.Record
	start := time.Now()

	tmpFile := ""
	if cfg.Cleanup {
		tmpFile = cfg.OutputFile + ".tmp.jsonl"
		defer func() {
			if tmpFile != "" {
				os.Remove(tmpFile)
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			elapsed := time.Since(start)
			log.Printf("Collected %d records in %v", len(records), elapsed)

			log.Printf("Writing to %s...", cfg.OutputFile)
			// Check DisableFlatten (inverted)
			if err := writeRecordsBatch(cfg.OutputFile, cfg.Format, records, !cfg.DisableFlatten); err != nil {
				log.Fatalf("Failed to write records: %v", err)
			}
			log.Printf("Successfully wrote %d records", len(records))

			if cfg.Graphs {
				generateGraphs(cfg.OutputFile, cfg.GraphDir)
			}
			return

		case <-ticker.C:
			dynamic := &collecting.DynamicMetrics{}
			record := manager.CollectDynamic(dynamic)
			records = append(records, record)

			if len(records)%100 == 0 {
				log.Printf("Progress: %d records (in memory)", len(records))
			}
		}
	}
}

func writeRecordsBatch(path, format string, records []exporting.Record, flatten bool) error {
	if flatten {
		for i := range records {
			records[i] = exporting.FlattenRecord(records[i])
		}
	}

	return exporting.SaveRecords(path, records, nil)
}
