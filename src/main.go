package main

import (
	"InferenceProfiler/src/profiler"
	"context"
	"fmt"
	"log"
	"os"

	"InferenceProfiler/src/config"
	"InferenceProfiler/src/output"
)

func main() {
	// Parse and validate configuration
	cfg, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Handle graph-only mode
	if cfg.IsGraphOnlyMode() {
		if err := generateGraphsOnly(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create and run profiler
	p, err := profiler.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize profiler: %v\n", err)
		os.Exit(1)
	}
	defer p.Close()

	// Run in appropriate mode
	ctx := context.Background()
	if cfg.HasSubprocess() {
		err = p.RunWithSubprocess(ctx, cfg.Subprocess)
	} else {
		err = p.Run(ctx)
	}

	if err != nil && err != context.Canceled {
		log.Printf("Profiler error: %v", err)
	}
}

// generateGraphsOnly handles the graph-only mode (-g flag with -i input file)
func generateGraphsOnly(cfg *config.Config) error {
	if cfg.InputFile == "" {
		return fmt.Errorf("graph-only mode requires -i <input file>")
	}

	err := output.GenerateGraphsFromOutputFile(cfg.InputFile)
	if err != nil {
		return fmt.Errorf("failed to generate graphs: %w", err)
	}

	//log.Printf("Generated report: %s", outputPath)
	return nil
}
