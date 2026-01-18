package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Profiler(args []string) {
	ctx, cleanup := InitCmd("profiler", args)
	defer cleanup()

	cfg := ctx.Config

	// Default to stream mode if neither batch nor stream specified
	if !cfg.Batch && !cfg.Stream {
		cfg.Stream = true
	}

	// Generate output filename if not specified
	if cfg.OutputFile == "" {
		mode := "profiler"
		if cfg.Delta {
			mode = "delta"
		}
		cfg.OutputFile = GenerateOutputFilename("", mode, cfg.Format)
	}

	// Create base context
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// If duration is set, create a timeout context
	if cfg.Duration > 0 {
		var timeoutCancel context.CancelFunc
		runCtx, timeoutCancel = context.WithTimeout(runCtx, time.Duration(cfg.Duration)*time.Millisecond)
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
		profilerDeltaMode(runCtx, ctx)
		return
	}

	// Regular continuous profiling (with optional duration limit)
	if cfg.Batch {
		profilerBatchMode(runCtx, ctx)
	} else {
		profilerStreamMode(runCtx, ctx)
	}
}

func profilerDeltaMode(runCtx context.Context, ctx *CmdContext) {
	cfg := ctx.Config
	flatten := ShouldFlatten(cfg)

	log.Printf("Delta profiler started")
	log.Printf("  Output: %s", cfg.OutputFile)
	log.Printf("  Format: %s", cfg.Format)
	if cfg.Duration > 0 {
		log.Printf("  Duration: %dms", cfg.Duration)
	} else {
		log.Printf("  Duration: until Ctrl+C")
	}

	// Use common RunDelta which waits on context
	capture := RunDelta(runCtx, ctx.Manager, flatten)

	// Write output
	log.Printf("Writing delta record to %s...", cfg.OutputFile)
	if err := SaveDelta(cfg.OutputFile, capture.DeltaRecord); err != nil {
		log.Fatalf("Failed to write delta record: %v", err)
	}

	log.Printf("Delta profiling completed in %v", capture.Duration)

	if cfg.Graphs {
		generateGraphs(cfg.OutputFile, cfg.GraphDir)
	}
}

func profilerStreamMode(runCtx context.Context, ctx *CmdContext) {
	cfg := ctx.Config
	flatten := ShouldFlatten(cfg)

	sc, err := NewStreamCollector(ctx.Manager, cfg.Format, cfg.OutputFile, flatten, cfg.Interval)
	if err != nil {
		log.Fatalf("Failed to initialize writer: %v", err)
	}
	if sc == nil {
		log.Fatalf("Unsupported format: %s", cfg.Format)
	}
	defer sc.Close()

	log.Printf("Profiler started (stream mode)")
	log.Printf("  Output: %s", cfg.OutputFile)
	log.Printf("  Format: %s", cfg.Format)
	log.Printf("  Interval: %dms", cfg.Interval)
	if cfg.Duration > 0 {
		log.Printf("  Duration: %dms (auto-stop)", cfg.Duration)
	} else {
		log.Printf("  Duration: until Ctrl+C")
	}

	start := time.Now()
	count := sc.Run(runCtx)
	elapsed := time.Since(start)

	log.Printf("Collected %d records in %v (%.2f records/sec)",
		count, elapsed, float64(count)/elapsed.Seconds())

	if cfg.Graphs {
		generateGraphs(cfg.OutputFile, cfg.GraphDir)
	}
}

func profilerBatchMode(runCtx context.Context, ctx *CmdContext) {
	cfg := ctx.Config
	flatten := ShouldFlatten(cfg)

	log.Printf("Profiler started (batch mode)")
	log.Printf("  Interval: %dms", cfg.Interval)
	if cfg.Duration > 0 {
		log.Printf("  Duration: %dms (auto-stop)", cfg.Duration)
	} else {
		log.Printf("  Duration: until Ctrl+C")
	}
	log.Println("  Collecting in memory, will write at end...")

	bc := NewBatchCollector(ctx.Manager, false, cfg.Interval) // Don't flatten during collection
	start := time.Now()
	bc.Run(runCtx)
	elapsed := time.Since(start)

	records := bc.Records()
	log.Printf("Collected %d records in %v", len(records), elapsed)

	log.Printf("Writing to %s...", cfg.OutputFile)
	if err := SaveBatch(cfg.OutputFile, records, flatten); err != nil {
		log.Fatalf("Failed to write records: %v", err)
	}
	log.Printf("Successfully wrote %d records", len(records))

	if cfg.Graphs {
		generateGraphs(cfg.OutputFile, cfg.GraphDir)
	}
}
