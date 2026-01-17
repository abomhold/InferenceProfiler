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

	if len(cmdArgs) == 0 {
		log.Fatalf("No command specified. Usage: infpro profile [flags] %s <command> [args]", utils.CMDSeparator)
	}

	if !cfg.Batch && !cfg.Stream {
		cfg.Stream = true
	}

	if cfg.OutputFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		cmdName := filepath.Base(cmdArgs[0])
		ext := exporting.GetExtension(cfg.Format)
		cfg.OutputFile = fmt.Sprintf("profile_%s_%s%s", cmdName, timestamp, ext)
	}

	manager := collecting.NewManager(cfg)
	defer manager.Close()

	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	targetCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr
	targetCmd.Stdin = os.Stdin

	log.Printf("Profiling command: %v", cmdArgs)
	log.Printf("Output: %s", cfg.OutputFile)

	if err := targetCmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	start := time.Now()

	go func() {
		defer close(done)
		if cfg.Batch {
			profileBatchMode(ctx, manager, cfg)
		} else {
			profileStreamMode(ctx, manager, cfg)
		}
	}()

	cmdErr := targetCmd.Wait()
	elapsed := time.Since(start)

	cancel()
	<-done

	log.Printf("Command completed in %v", elapsed)

	if cmdErr != nil {
		log.Printf("Command exited with error: %v", cmdErr)
	}

	if cfg.Graphs {
		generateGraphs(cfg.OutputFile, cfg.GraphDir)
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

			// Check DisableFlatten (inverted)
			if !cfg.DisableFlatten {
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
			// Check DisableFlatten (inverted)
			writeRecordsBatch(cfg.OutputFile, cfg.Format, records, !cfg.DisableFlatten)
			return

		case <-ticker.C:
			dynamic := &collecting.DynamicMetrics{}
			record := manager.CollectDynamic(dynamic)
			records = append(records, record)
		}
	}
}
