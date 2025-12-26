package main

import (
	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/utils"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

func main() {
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")

	// Arguments
	outputDir := flag.String("o", "./profiler-output", "Output directory for logs")
	interval := flag.Int("t", 1000, "Sampling interval in milliseconds")
	format := flag.String("f", "parquet", "Final export format (parquet, csv, tsv)")
	collectProcs := flag.Bool("p", false, "Collect per-process metrics")
	flag.Parse()
	args := flag.Args()

	sessionUUID := uuid.New()

	log.Printf("Session UUID: %s", sessionUUID)
	log.Printf("Output Dir:   %s", *outputDir)
	log.Printf("Interval:     %dms", *interval)
	log.Printf("Format:       %s", *format)

	// Initialize
	collector := collectors.NewCollectorManager(*collectProcs)
	defer collector.Close()
	exp, err := utils.NewExporter(*outputDir, sessionUUID)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// Capture and save static info
	log.Println("Capturing static hardware info...")
	staticData := collector.GetStaticInfo(sessionUUID)
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
	if err := exp.ProcessSession(*format); err != nil {
		log.Printf("Error processing session: %v", err)
	}

	log.Println("Done.")
}
