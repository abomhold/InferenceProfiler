package cmd

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Profile(args []string) {
	ctx, cleanup, cmdArgs := InitCmdWithSeparator("profile", args)
	defer cleanup()

	cfg := ctx.Config

	if len(cmdArgs) == 0 {
		log.Fatalf("No command specified. Usage: infpro profile [flags] -- <command> [args]")
	}

	// Default to stream mode if neither batch nor stream specified
	if !cfg.Batch && !cfg.Stream {
		cfg.Stream = true
	}

	// Generate output filename if not specified
	if cfg.OutputFile == "" {
		cmdName := filepath.Base(cmdArgs[0])
		mode := "profile"
		if cfg.Delta {
			mode = "delta"
		}
		cfg.OutputFile = GenerateOutputFilename(cmdName, mode, cfg.Format)
	}

	targetCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr
	targetCmd.Stdin = os.Stdin

	log.Printf("Profiling command: %v", cmdArgs)
	log.Printf("Output: %s", cfg.OutputFile)

	// Delta mode: take initial snapshot, run command, take final snapshot
	if cfg.Delta {
		log.Printf("Mode: delta (initial + final snapshot)")
		profileDeltaMode(ctx, targetCmd)
		return
	}

	// Regular continuous profiling during command execution
	if err := targetCmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	start := time.Now()

	go func() {
		defer close(done)
		if cfg.Batch {
			profileBatchMode(runCtx, ctx)
		} else {
			profileStreamMode(runCtx, ctx)
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

func profileDeltaMode(ctx *CmdContext, targetCmd *exec.Cmd) {
	cfg := ctx.Config
	expandAll := ShouldExpandAll(cfg)

	// Take initial snapshot before command starts
	log.Println("Capturing initial snapshot...")
	initial := CollectSnapshot(ctx.Manager, expandAll)
	startTime := time.Now()

	// Start command
	if err := targetCmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// Wait for command to complete
	cmdErr := targetCmd.Wait()
	elapsed := time.Since(startTime)

	// Take final snapshot after command completes
	log.Println("Capturing final snapshot...")
	final := CollectSnapshot(ctx.Manager, expandAll)

	// Calculate and save delta
	delta := ComputeDelta(initial, final, elapsed.Milliseconds())

	log.Printf("Writing delta record to %s...", cfg.OutputFile)
	if err := SaveDelta(cfg.OutputFile, delta); err != nil {
		log.Fatalf("Failed to write delta record: %v", err)
	}

	log.Printf("Command completed in %v", elapsed)

	if cmdErr != nil {
		log.Printf("Command exited with error: %v", cmdErr)
	}

	if cfg.Graphs {
		generateGraphs(cfg.OutputFile, cfg.GraphDir)
	}
}

func profileStreamMode(runCtx context.Context, ctx *CmdContext) {
	cfg := ctx.Config
	expandAll := ShouldExpandAll(cfg)

	sc, err := NewStreamCollector(ctx.Manager, cfg.Format, cfg.OutputFile, expandAll, cfg.Interval)
	if err != nil {
		log.Printf("Stream collector init error: %v", err)
		return
	}
	if sc == nil {
		log.Printf("Unsupported format: %s", cfg.Format)
		return
	}
	defer sc.Close()

	count := sc.Run(runCtx)
	log.Printf("Collected %d records", count)
}

func profileBatchMode(runCtx context.Context, ctx *CmdContext) {
	cfg := ctx.Config
	expandAll := ShouldExpandAll(cfg)

	bc := NewBatchCollector(ctx.Manager, false, cfg.Interval) // Don't expand during collection
	bc.Run(runCtx)

	records := bc.Records()
	log.Printf("Writing %d records...", len(records))

	if err := SaveBatch(cfg.OutputFile, records, expandAll); err != nil {
		log.Printf("Failed to write records: %v", err)
	}
}
