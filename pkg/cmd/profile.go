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
	"os/exec"
	"path/filepath"
	"time"
)

func Profile(args []string) {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	flags := &CommonFlags{}

	addCollectionFlags(fs, flags)
	addOutputFlags(fs, flags, "parquet")
	addModeFlags(fs, flags)
	addProfilingFlags(fs, flags)
	addGraphFlags(fs, flags)

	// Find the -- separator
	dashIdx := -1
	for i, arg := range args {
		if arg == "--" {
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

	if len(cmdArgs) == 0 {
		log.Fatal("No command specified. Usage: infpro profile [flags] -- <command> [args]")
	}

	// Default to stream mode
	if !flags.Batch && !flags.Stream {
		flags.Stream = true
	}

	cfg := flags.ToConfig()

	// Generate output filename if not provided
	if cfg.OutputFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		cmdName := filepath.Base(cmdArgs[0])
		ext := exporting.GetExtension(cfg.Format)
		cfg.OutputFile = fmt.Sprintf("profile_%s_%s%s", cmdName, timestamp, ext)
	}

	// Initialize collector manager
	manager := collecting.NewManager(cfg)
	defer manager.Close()

	// Collect static metrics
	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	// Start target command
	targetCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr
	targetCmd.Stdin = os.Stdin

	log.Printf("Profiling command: %v", cmdArgs)
	log.Printf("Output: %s", cfg.OutputFile)

	if err := targetCmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// Profile in background
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	start := time.Now()

	go func() {
		defer close(done)
		if flags.Batch {
			profileBatchMode(ctx, manager, cfg)
		} else {
			profileStreamMode(ctx, manager, cfg)
		}
	}()

	// Wait for command to complete
	cmdErr := targetCmd.Wait()
	elapsed := time.Since(start)

	cancel()
	<-done

	log.Printf("Command completed in %v", elapsed)

	if cmdErr != nil {
		log.Printf("Command exited with error: %v", cmdErr)
	}

	if flags.Graphs {
		generateGraphs(cfg.OutputFile, flags.GraphDir)
	}
}

func profileStreamMode(ctx context.Context, manager *collecting.Manager, cfg *utils.Config) {
	f, _ := exporting.Get(cfg.Format)
	writer := f.Writer()
	if err := writer.Init(cfg.OutputFile, nil); err != nil {
		log.Printf("Writer init error: %v", err)
		return
	}
	defer writer.Close()

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Millisecond)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			writer.Flush()
			log.Printf("Collected %d records", count)
			return

		case <-ticker.C:
			dynamic := &collecting.DynamicMetrics{}
			record := manager.CollectDynamic(dynamic)

			if cfg.Flatten {
				record = exporting.FlattenRecord(record)
			}

			writer.Write(record)
			count++

			if count%50 == 0 {
				writer.Flush()
			}
		}
	}
}

func profileBatchMode(ctx context.Context, manager *collecting.Manager, cfg *utils.Config) {
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Millisecond)
	defer ticker.Stop()

	var records []exporting.Record

	for {
		select {
		case <-ctx.Done():
			log.Printf("Writing %d records...", len(records))
			writeRecordsBatch(cfg.OutputFile, cfg.Format, records, cfg.Flatten)
			return

		case <-ticker.C:
			dynamic := &collecting.DynamicMetrics{}
			record := manager.CollectDynamic(dynamic)
			records = append(records, record)
		}
	}
}
