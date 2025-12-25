// Package InferenceProfiler
// profiler collects system resource metrics and exports to JSON/CSV/Parquet.
//
// Usage:
//
//	inference-profiler [flags] [-- command args...]
//
// Flags:
//
//	-o, --output DIR     Output directory (default: ./profiler-output)
//	-t, --interval MS    Sampling interval in milliseconds (default: 1000)
//	-f, --format FMT     Export format: csv, tsv, parquet (default: parquet)
//	-p, --processes      Enable per-process metrics (expensive)
//	-h, --help           Show help
//
// If a command is provided after --, the profiler runs until that command exits.
// Otherwise, it runs until interrupted (Ctrl+C).
package InferenceProfiler

import (
	"InferenceProfiler/collectors"
	"InferenceProfiler/utils"

	//"crypto/rand"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	//"fmt"
	"github.com/google/uuid"
)

func main() {
	// Flags
	outputDir := flag.String("o", "./profiler-output", "Output directory")
	interval := flag.Int("t", 1000, "Sampling interval (ms)")
	format := flag.String("f", "parquet", "Export format: csv, tsv, parquet")
	processes := flag.Bool("p", false, "Collect per-process metrics")
	flag.Parse()

	// UUID for session
	sessionUUID, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Session UUID: %s", sessionUUID)
	log.Printf("Output Dir:   %s", *outputDir)
	log.Printf("Interval:     %dms", *interval)
	log.Printf("Format:       %s", *format)

	// Initialize
	mgr := collectors.NewManager(*processes)
	defer mgr.Close()

	exp := utils.New(*outputDir, sessionUUID)

	// Save static info
	log.Println("Capturing static hardware info...")
	static := mgr.GetStatic(sessionUUID)
	if err := exp.SaveStatic(static); err != nil {
		log.Printf("Warning: failed to save static info: %v", err)
	}

	// Optional subprocess
	var proc *exec.Cmd
	args := flag.Args()
	if len(args) > 0 {
		log.Printf("Starting subprocess: %v", args)
		proc = exec.Command(args[0], args[1:]...)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Start(); err != nil {
			log.Fatalf("Failed to start command: %v", err)
		}
	}

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Profiling loop
	log.Println("Profiling started. Press Ctrl+C to stop.")
	ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
	defer ticker.Stop()

	running := true
	for running {
		select {
		case <-ticker.C:
			snapshot := mgr.Collect()
			if err := exp.SaveSnapshot(snapshot); err != nil {
				log.Printf("Warning: failed to save snapshot: %v", err)
			}

			// Check subprocess
			if proc != nil {
				if proc.ProcessState != nil && proc.ProcessState.Exited() {
					log.Printf("Subprocess finished with exit code %d", proc.ProcessState.ExitCode())
					running = false
				}
			}

		case sig := <-sigCh:
			log.Printf("Signal %v received. Stopping...", sig)
			running = false
		}
	}

	// Cleanup
	log.Println("Shutting down...")

	if proc != nil && proc.Process != nil && proc.ProcessState == nil {
		proc.Process.Signal(syscall.SIGTERM)
		done := make(chan error, 1)
		go func() { done <- proc.Wait() }()

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
