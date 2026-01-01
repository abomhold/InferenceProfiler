package main

import (
	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/output"

	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
)

func main() {
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")

	// -------------------------------------------------------------------------
	// 1. Configuration & Flags
	// -------------------------------------------------------------------------
	outputDir := flag.String("o", "./profiler-output", "Output directory for logs")
	intervalArg := flag.Int("t", 1, "Sampling interval in milliseconds")
	format := flag.String("f", "jsonl", "Export format: jsonl, parquet, csv, tsv")
	noFlatten := flag.Bool("no-flatten", false, "Disable flattening nested data (GPUs, processes) to columns")
	noCleanup := flag.Bool("no-cleanup", false, "Disable deleting intermediary snapshot files after final export")
	stream := flag.Bool("stream", false, "Stream mode: write directly to output file")

	// Collector toggles
	noCPU := flag.Bool("no-cpu", false, "Disable CPU metrics collection")
	noMemory := flag.Bool("no-memory", false, "Disable memory metrics collection")
	noDisk := flag.Bool("no-disk", false, "Disable disk metrics collection")
	noNetwork := flag.Bool("no-network", false, "Disable network metrics collection")
	noContainer := flag.Bool("no-container", false, "Disable container/cgroup metrics collection")
	noProcs := flag.Bool("no-procs", false, "Disable per-process metrics collection")
	noNvidia := flag.Bool("no-nvidia", false, "Disable NVIDIA GPU metrics collection")
	noGPUProcs := flag.Bool("no-gpu-procs", false, "Disable GPU process enumeration (still collect GPU metrics)")
	noVLLM := flag.Bool("no-vllm", false, "Disable vLLM metrics collection")

	flag.Parse()
	args := flag.Args()

	// Validate format
	validFormats := map[string]bool{"jsonl": true, "parquet": true, "csv": true, "tsv": true}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Invalid format: %s. Use: jsonl, parquet, csv, tsv\n", *format)
		os.Exit(1)
	}

	sessionUUID := uuid.New()

	log.Printf("Session UUID: %s", sessionUUID)
	log.Printf("Output Dir:   %s", *outputDir)
	log.Printf("Interval:     %dms", *intervalArg)
	log.Printf("Format:       %s", *format)
	log.Printf("Mode:         %s", map[bool]string{true: "streaming", false: "batch"}[*stream])
	log.Printf("Flatten:      %v", !*noFlatten)
	if !*stream {
		log.Printf("Cleanup:      %v", !*noCleanup)
	}

	// Build collector config
	cfg := collectors.CollectorConfig{
		CPU:         !*noCPU,
		Memory:      !*noMemory,
		Disk:        !*noDisk,
		Network:     !*noNetwork,
		Container:   !*noContainer,
		Processes:   !*noProcs,
		Nvidia:      !*noNvidia,
		NvidiaProcs: !*noGPUProcs,
		VLLM:        !*noVLLM,
	}

	collector := collectors.NewCollectorManager(cfg)
	exp, err := output.NewExporter(*outputDir, sessionUUID, !*noFlatten || *stream, *stream, *format)
	if err != nil {
		collector.Close()
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// Capture and save static info
	log.Println("Capturing static hardware info...")
	staticData := collector.CollectStaticMetrics(sessionUUID)
	if err := exp.SaveStatic(staticData); err != nil {
		log.Printf("Warning: Failed to save static info: %v", err)
	}

	// Create a cancellable context for the whole app
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var proc *exec.Cmd
	processDone := make(chan error, 1)

	// Start subprocess if command provided
	if len(args) > 0 {
		log.Printf("Starting subprocess: %s", args)
		proc = exec.Command(args[0], args[1:]...)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr

		if err := proc.Start(); err != nil {
			log.Fatalf("Failed to start command: %v", err)
		}

		// Monitor subprocess in background
		go func() {
			err := proc.Wait()
			processDone <- err
			if err != nil {
				log.Printf("Subprocess finished with error: %v", err)
			} else {
				log.Printf("Subprocess finished successfully.")
			}
			// Trigger main shutdown when process exits
			cancel()
		}()
	}

	var wg sync.WaitGroup
	var collectionWg sync.WaitGroup // Track in-flight collection goroutines
	wg.Add(1)

	go func() {
		defer wg.Done()

		intervalDuration := time.Duration(*intervalArg) * time.Millisecond
		ticker := time.NewTicker(intervalDuration)
		defer ticker.Stop()

		var busy int32 = 0

		if *stream {
			log.Printf("Profiling started (streaming to %s). Press Ctrl+C to stop.", *format)
		} else {
			log.Println("Profiling started. Press Ctrl+C to stop.")
		}

		for {
			select {
			case <-ctx.Done():
				return // Stop loop

			case <-ticker.C:
				// NON-BLOCKING CHECK: Skip if previous collection is still running
				if !atomic.CompareAndSwapInt32(&busy, 0, 1) {
					continue
				}

				// Run collection
				collectionWg.Add(1)
				go func() {
					defer collectionWg.Done()
					defer atomic.StoreInt32(&busy, 0)

					// Double check context before doing work (optional optimization)
					if ctx.Err() != nil {
						return
					}

					metrics := collector.CollectDynamicMetrics()
					if err := exp.SaveSnapshot(metrics); err != nil {
						log.Printf("Error saving snapshot: %v", err)
					}
				}()
			}
		}
	}()

	<-ctx.Done() // Waits for Signal OR Subprocess Exit
	log.Println("Shutting down...")

	// Kill subprocess if it's still alive
	if proc != nil && proc.Process != nil {
		// Check if it's already dead
		select {
		case <-processDone:
			// Already done
		default:
			log.Println("Terminating subprocess...")
			proc.Process.Signal(syscall.SIGTERM)

			// Quick wait for SIGTERM to work
			killTimer := time.NewTimer(2 * time.Second)
			select {
			case <-processDone:
				// Exited gracefully
			case <-killTimer.C:
				log.Println("Subprocess did not exit, sending SIGKILL...")
				proc.Process.Kill()
			}
		}
	}

	// Wait for profiler loop to finish with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		collectionWg.Wait()
		collector.Close()

		if *stream {
			log.Printf("Finalizing %s stream...", *format)
			if err := exp.CloseStream(); err != nil {
				log.Printf("Error closing stream: %v", err)
			}
		} else {
			log.Printf("Converting session data to %s...", *format)
			exp.CloseStream()
			err := exp.ProcessSession()
			if err != nil {
				log.Printf("Error processing session: %v", err)
			} else {
				if !*noCleanup {
					log.Println("Cleaning up intermediary files...")
					exp.Cleanup()
				}
			}
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("Done.")
	case <-shutdownCtx.Done():
		log.Println("!! Cleanup timed out (likely blocked in CGO). Forcing exit. !!")
		os.Exit(1)
	}
}
