package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Snapshot(args []string) {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
	cfg := utils.NewConfig()
	applyFlags := utils.GetFlags(fs, cfg)

	// Additional delta-specific flags
	var deltaDuration time.Duration
	var graphOutput string
	var showErrors bool

	parseArgs := args
	if len(args) > 0 && args[0] == "ss" {
		parseArgs = args[1:]
	}
	fs.Parse(parseArgs)
	applyFlags()

	if cfg.Delta {
		runDeltaSnapshot(cfg, deltaDuration, graphOutput, showErrors)
	} else {
		runSingleSnapshot(cfg)
	}
}

func runSingleSnapshot(cfg *utils.Config) {
	manager := collecting.NewManager(cfg)
	defer manager.Close()

	record := make(map[string]interface{})

	if cfg.CollectStatic {
		static := &collecting.StaticMetrics{
			UUID:     cfg.UUID,
			Hostname: cfg.Hostname,
		}
		manager.CollectStatic(static)
		if s := manager.GetStaticRecord(); s != nil {
			record["static"] = s
		}
	}

	if cfg.CollectDynamic {
		dRecord := CollectSnapshot(manager, ShouldExpandAll(cfg))

		for k, v := range dRecord {
			record[k] = v
		}
	}

	output, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal output: %v", err)
	}

	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, output, 0644); err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		log.Printf("Wrote snapshot to %s", cfg.OutputFile)
	} else {
		fmt.Println(string(output))
	}
}

func runDeltaSnapshot(cfg *utils.Config, duration time.Duration, graphOutput string, showErrors bool) {
	manager := collecting.NewManager(cfg)
	defer manager.Close()

	log.Printf("Taking initial snapshot...")

	// Collect initial snapshot (always flatten for delta)
	initialRecord := CollectSnapshot(manager, true)
	initialTime := time.Now()

	log.Printf("Waiting %v before taking final snapshot...", duration)
	time.Sleep(duration)

	log.Printf("Taking final snapshot...")

	// Collect final snapshot (always flatten for delta)
	finalRecord := CollectSnapshot(manager, true)
	finalTime := time.Now()

	// Compute delta with error tracking
	durationMs := finalTime.Sub(initialTime).Milliseconds()
	deltaResult := exporting.DeltaRecordWithErrors(initialRecord, finalRecord, durationMs)

	// Add static data if requested
	if cfg.CollectStatic {
		static := &collecting.StaticMetrics{
			UUID:     cfg.UUID,
			Hostname: cfg.Hostname,
		}
		manager.CollectStatic(static)
		if s := manager.GetStaticRecord(); s != nil {
			// Copy static fields to delta record
			for k, v := range s {
				deltaResult.Record[k] = v
			}
		}
	}

	// Build output object
	outputRecord := deltaResult.Record
	if showErrors && len(deltaResult.Errors) > 0 {
		// Add errors to a metadata field
		errorsJson := make([]map[string]interface{}, len(deltaResult.Errors))
		for i, e := range deltaResult.Errors {
			errorsJson[i] = map[string]interface{}{
				"field":   e.Field,
				"reason":  e.Reason,
				"initial": fmt.Sprintf("%v", e.Initial),
				"final":   fmt.Sprintf("%v", e.Final),
			}
		}
		outputRecord["_delta_errors"] = errorsJson
		outputRecord["_delta_error_count"] = len(deltaResult.Errors)
	}

	// Marshal output
	output, err := json.MarshalIndent(outputRecord, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal output: %v", err)
	}

	// Write to file or stdout
	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, output, 0644); err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		log.Printf("Wrote delta snapshot to %s", cfg.OutputFile)
	} else {
		fmt.Println(string(output))
	}

	// Generate graph if requested
	if graphOutput != "" {
		log.Printf("Generating graph visualization...")

		if err := exporting.GenerateDeltaGraph(deltaResult, graphOutput); err != nil {
			log.Printf("Warning: failed to generate graph: %v", err)
		} else {
			log.Printf("Generated graph in %s", graphOutput)
		}

		// Also write the original and flattened records for debugging
		origPath := filepath.Join(graphOutput, "original_initial.json")
		if data, err := json.MarshalIndent(initialRecord, "", "  "); err == nil {
			os.WriteFile(origPath, data, 0644)
		}

		finalPath := filepath.Join(graphOutput, "original_final.json")
		if data, err := json.MarshalIndent(finalRecord, "", "  "); err == nil {
			os.WriteFile(finalPath, data, 0644)
		}
	}

	// Report summary
	log.Printf("Delta computed over %v, %d fields processed, %d errors",
		duration, len(deltaResult.Record), len(deltaResult.Errors))
}
