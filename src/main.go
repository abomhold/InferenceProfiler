package main

import (
	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/utils"

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

	// Validate format
	validFormats := map[string]bool{"jsonl": true, "parquet": true, "csv": true, "tsv": true}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Invalid format: %s. Use: jsonl, parquet, csv, tsv\n", *format)
		os.Exit(1)
	}

	sessionUUID := uuid.New()

	log.Printf("Session UUID: %s", sessionUUID)
	log.Printf("Output Dir:   %s", *outputDir)
	log.Printf("Interval:     %dms", *interval)
	log.Printf("Format:       %s", *format)
	log.Printf("Flatten:      %v", !*noFlatten)
	log.Printf("Cleanup:      %v", !*noCleanup)

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
	if *noProcs {
		log.Printf("Processes:    enabled")
	}

	// Initialize
	collector := collectors.NewCollectorManager(cfg)
	defer collector.Close()
	exp, err := utils.NewExporter(*outputDir, sessionUUID, !*noFlatten)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// Capture and save static info
	log.Println("Capturing static hardware info...")
	staticData := collector.GetStaticMetrics(sessionUUID)
	if err := exp.SaveStatic(staticData); err != nil {
		log.Printf("Warning: Failed to save static info: %v", err)
	}

	// Start subprocess if command provided
	var proc *exec.Cmd
	if len(args) > 0 {
		log.Printf("Starting subprocess: %s", args)
		proc = exec.Command(args[0], args[1:]...)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Start(); err != nil {
			log.Fatalf("Failed to start command: %v", err)
		}
	}

	// Setup signal handling
	running := true
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Signal %v received. Stopping profiler...", sig)
		running = false
	}()

	// Profiling loop
	log.Println("Profiling started. Press Ctrl+C to stop.")
	intervalDuration := time.Duration(*interval) * time.Millisecond

	for running {
		loopStart := time.Now()

		// Collect and save metrics
		metrics := collector.CollectMetrics()
		if err := exp.SaveSnapshot(metrics); err != nil {
			log.Printf("Error saving snapshot: %v", err)
		}

		// Check subprocess status
		if proc != nil {
			if proc.ProcessState != nil && proc.ProcessState.Exited() {
				log.Printf("Subprocess finished with exit code %d", proc.ProcessState.ExitCode())
				running = false
				break
			}
			// Check if process has exited
			if err := proc.Process.Signal(syscall.Signal(0)); err != nil {
				proc.Wait()
				if proc.ProcessState != nil {
					log.Printf("Subprocess finished with exit code %d", proc.ProcessState.ExitCode())
				}
				running = false
				break
			}
		}

		// Sleep for remaining interval
		elapsed := time.Since(loopStart)
		if sleep := intervalDuration - elapsed; sleep > 0 {
			time.Sleep(sleep)
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
	log.Println("Done.")
}
