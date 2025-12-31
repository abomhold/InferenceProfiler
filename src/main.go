package main

import (
	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/output"

	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

func main() {
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")

	// Output options
	outputDir := flag.String("o", "./profiler-output", "Output directory for logs")
	interval := flag.Int("t", 1000, "Sampling interval in milliseconds")
	format := flag.String("f", "jsonl", "Export format: jsonl, parquet, csv, tsv")
	noFlatten := flag.Bool("no-flatten", false, "Disable flattening nested data (GPUs, processes) to columns; false keeps JSON strings")
	noCleanup := flag.Bool("no-cleanup", false, "Disable deleting intermediary snapshot files after final export")
	streamParquet := flag.Bool("stream-parquet", false, "Stream metrics directly to parquet file (ignores -f, always uses flattened format)")

	// Collector toggles (all enabled by default)
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

	// Validate format (not used if streaming)
	if !*streamParquet {
		validFormats := map[string]bool{"jsonl": true, "parquet": true, "csv": true, "tsv": true}
		if !validFormats[*format] {
			fmt.Fprintf(os.Stderr, "Invalid format: %s. Use: jsonl, parquet, csv, tsv\n", *format)
			os.Exit(1)
		}
	}

	sessionUUID := uuid.New()

	log.Printf("Session UUID: %s", sessionUUID)
	log.Printf("Output Dir:   %s", *outputDir)
	log.Printf("Interval:     %dms", *interval)

	if *streamParquet {
		log.Printf("Mode:         streaming parquet (ignores -f flag)")
		log.Printf("Flatten:      true (streaming always flattens)")
	} else {
		log.Printf("Format:       %s", *format)
		log.Printf("Flatten:      %v", !*noFlatten)
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

	// Log disabled collectors
	var disabled []string
	if *noCPU {
		disabled = append(disabled, "cpu")
	}
	if *noMemory {
		disabled = append(disabled, "memory")
	}
	if *noDisk {
		disabled = append(disabled, "disk")
	}
	if *noNetwork {
		disabled = append(disabled, "network")
	}
	if *noContainer {
		disabled = append(disabled, "container")
	}
	if *noNvidia {
		disabled = append(disabled, "nvidia")
	}
	if *noGPUProcs && !*noNvidia {
		disabled = append(disabled, "gpu-procs")
	}
	if *noVLLM {
		disabled = append(disabled, "vllm")
	}
	if len(disabled) > 0 {
		log.Printf("Disabled:     %s", strings.Join(disabled, ", "))
	}
	if !*noProcs {
		log.Printf("Processes:    enabled")
	}

	// Initialize
	collector := collectors.NewCollectorManager(cfg)
	defer collector.Close()

	// For streaming, always use flatten mode
	exp, err := output.NewExporter(*outputDir, sessionUUID, !*noFlatten || *streamParquet, *streamParquet)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	defer exp.CloseStream() // Safe to call even if not streaming

	// Capture and save static info
	log.Println("Capturing static hardware info...")
	staticData := collector.GetStaticMetrics(sessionUUID)
	if err := exp.SaveStatic(staticData); err != nil {
		log.Printf("Warning: Failed to save static info: %v", err)
	}

	// Start subprocess if command provided
	var proc *exec.Cmd
	processDone := make(chan error, 1)

	if len(args) > 0 {
		log.Printf("Starting subprocess: %s", args)
		proc = exec.Command(args[0], args[1:]...)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Start(); err != nil {
			log.Fatalf("Failed to start command: %v", err)
		}

		go func() {
			processDone <- proc.Wait()
		}()
	}

	// Setup signal handling
	running := true
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Profiling loop
	if *streamParquet {
		log.Println("Profiling started (streaming to parquet). Press Ctrl+C to stop.")
	} else {
		log.Println("Profiling started. Press Ctrl+C to stop.")
	}
	intervalDuration := time.Duration(*interval) * time.Millisecond

	for running {
		loopStart := time.Now()

		// Check if subprocess finished
		if proc != nil {
			select {
			case err := <-processDone:
				if err != nil {
					log.Printf("Subprocess finished with error: %v", err)
				} else {
					log.Printf("Subprocess finished successfully.")
				}
				running = false
			default:
				// Process is still running, continue to collect metrics
			}
		}

		if !running {
			break
		}

		// Collect and save metrics
		metrics := collector.CollectMetrics()
		if err := exp.SaveSnapshot(metrics); err != nil {
			log.Printf("Error saving snapshot: %v", err)
		}

		// Sleep for remaining interval
		elapsed := time.Since(loopStart)
		if sleep := intervalDuration - elapsed; sleep > 0 {
			select {
			case <-time.After(sleep):
				// Continue
			case err := <-processDone:
				log.Printf("Subprocess finished during sleep.")
				if err != nil {
					log.Printf("Exit error: %v", err)
				}
				running = false
			case sig := <-sigChan:
				log.Printf("Signal %v received. Stopping...", sig)
				running = false
			}
		}
	}

	// Cleanup
	log.Println("Shutting down...")

	if proc != nil && proc.Process != nil {
		log.Println("Terminating subprocess...")
		proc.Process.Signal(syscall.SIGTERM)

		done := make(chan error, 1)
		go func() {
			done <- proc.Wait()
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			proc.Process.Kill()
		}
	}

	if *streamParquet {
		log.Println("Finalizing parquet stream...")
		if err := exp.CloseStream(); err != nil {
			log.Printf("Error closing parquet stream: %v", err)
		}
	} else {
		log.Printf("Converting session data to %s...", *format)
		err = exp.ProcessSession(*format)
		if err != nil {
			log.Printf("Error processing session: %v", err)
		} else {
			if !*noCleanup {
				log.Println("Cleaning up intermediary files...")
				exp.Cleanup()
			}
		}
	}

	log.Println("Done.")
}
