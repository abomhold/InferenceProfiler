package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func Profiler(args []string) {
	fs := flag.NewFlagSet("profiler", flag.ExitOnError)
	flags := &CommonFlags{}

	addCollectionFlags(fs, flags)
	addOutputFlags(fs, flags, "parquet")
	addModeFlags(fs, flags)
	addProfilingFlags(fs, flags)
	addGraphFlags(fs, flags)

	fs.Parse(args)

	// Default to stream mode if neither specified
	if !flags.Batch && !flags.Stream {
		flags.Stream = true
	}

	cfg := flags.ToConfig()

	// Generate output filename if not provided
	if cfg.OutputFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		ext := exporting.GetExtension(cfg.Format)
		cfg.OutputFile = fmt.Sprintf("profiler_%s%s", timestamp, ext)
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

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	if flags.Batch {
		runBatchMode(ctx, manager, cfg, flags)
	} else {
		runStreamMode(ctx, manager, cfg, flags)
	}
}

func runStreamMode(ctx context.Context, manager *collecting.Manager, cfg *utils.Config, flags *CommonFlags) {
	// Create writer
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

			if flags.Graphs {
				generateGraphs(cfg.OutputFile, flags.GraphDir)
			}
			return

		case <-ticker.C:
			dynamic := &collecting.DynamicMetrics{}
			record := manager.CollectDynamic(dynamic)

			if cfg.Flatten {
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

func runBatchMode(ctx context.Context, manager *collecting.Manager, cfg *utils.Config, flags *CommonFlags) {
	log.Printf("Profiler started (batch mode)")
	log.Printf("  Interval: %dms", cfg.Interval)
	log.Println("  Collecting in memory, will write at end...")

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Millisecond)
	defer ticker.Stop()

	var records []exporting.Record
	start := time.Now()

	// Temporary JSON file for intermediate storage
	tmpFile := ""
	if flags.Cleanup {
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

			// Write all records
			log.Printf("Writing to %s...", cfg.OutputFile)
			if err := writeRecordsBatch(cfg.OutputFile, cfg.Format, records, cfg.Flatten); err != nil {
				log.Fatalf("Failed to write records: %v", err)
			}
			log.Printf("Successfully wrote %d records", len(records))

			if flags.Graphs {
				generateGraphs(cfg.OutputFile, flags.GraphDir)
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
