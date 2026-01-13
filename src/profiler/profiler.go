// Package profiler provides the main orchestration for system profiling.
// It coordinates configuration, collection, export, and signal handling.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/config"
	"InferenceProfiler/src/output"
)

// Profiler orchestrates system metric collection, export, and lifecycle management.
// It supports two modes of operation:
//   - Standalone mode: Run for a specified duration or until interrupted
//   - Subprocess mode: Profile while a subprocess runs, stopping when it exits
type Profiler struct {
	// Configuration
	cfg *config.Config

	// Components
	manager  *collectors.CollectorManager
	exporter *output.Exporter

	// State
	staticMetrics *collectors.StaticMetrics
	sampleCount   int
	startTime     time.Time
	stopRequested bool

	// Synchronization
	mu       sync.Mutex
	stopOnce sync.Once
}

// New creates a new Profiler with the given configuration.
// Returns an error if initialization fails.
func New(cfg *config.Config) (*Profiler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	p := &Profiler{
		cfg: cfg,
	}

	// Initialize collector manager
	// FIX: cfg.Collectors is already config.CollectorConfig, passed directly
	p.manager = collectors.NewCollectorManager(cfg.Collectors)

	// Initialize exporter
	opts := []output.ExporterOption{
		output.WithFormat(cfg.Format),
		output.WithFlatten(cfg.Flatten),
	}
	if cfg.Stream {
		opts = append(opts, output.WithStreaming())
	}

	var err error
	p.exporter, err = output.NewExporter(cfg.OutputDir, cfg.SessionUUID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	return p, nil
}

// Run starts profiling and blocks until completion or interruption.
// In standalone mode, runs indefinitely (or for duration if set).
// Use context cancellation to stop profiling gracefully.
func (p *Profiler) Run(ctx context.Context) error {
	return p.run(ctx, nil)
}

// RunWithSubprocess profiles while a subprocess runs.
// Profiling stops automatically when the subprocess exits.
// Returns the subprocess exit error (if any) along with any profiling errors.
func (p *Profiler) RunWithSubprocess(ctx context.Context, cmdArgs []string) error {
	if len(cmdArgs) == 0 {
		return fmt.Errorf("subprocess command cannot be empty")
	}

	// Create subprocess
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return p.run(ctx, cmd)
}

// run is the internal run implementation supporting both modes
func (p *Profiler) run(ctx context.Context, subprocess *exec.Cmd) error {
	p.startTime = time.Now()

	// Setup signal handling
	ctx, cancel := p.setupSignalHandling(ctx)
	defer cancel()

	// Collect static metrics first
	// FIX: Passed p.cfg.SessionUUID argument
	p.staticMetrics = p.manager.CollectStaticMetrics(p.cfg.SessionUUID)
	if err := p.exporter.WriteStaticMetrics(p.staticMetrics); err != nil {
		log.Printf("Warning: failed to write static metrics: %v", err)
	}

	// Print startup info if not in graph-only mode
	if !p.cfg.IsGraphOnlyMode() {
		p.printStartupInfo()
	}

	// Handle subprocess mode
	var subprocessDone <-chan error
	if subprocess != nil {
		// Start subprocess
		if err := subprocess.Start(); err != nil {
			return fmt.Errorf("failed to start subprocess: %w", err)
		}

		// Monitor subprocess in goroutine
		subprocessChan := make(chan error, 1)
		go func() {
			subprocessChan <- subprocess.Wait()
		}()
		subprocessDone = subprocessChan
	}

	// Run collection loop
	err := p.collectionLoop(ctx, subprocessDone)

	// Finalize export
	if finalizeErr := p.exporter.Finalize(); finalizeErr != nil {
		log.Printf("Warning: failed to finalize export: %v", finalizeErr)
	}

	// Generate graphs if requested
	if p.cfg.GenerateGraphs {
		if graphErr := p.generateGraphs(); graphErr != nil {
			log.Printf("Warning: failed to generate graphs: %v", graphErr)
		}
	}

	// Cleanup
	if cleanupErr := p.exporter.Cleanup(); cleanupErr != nil {
		log.Printf("Warning: failed to cleanup: %v", cleanupErr)
	}

	// Print summary
	p.printSummary()

	return err
}

// collectionLoop runs the main metrics collection loop
func (p *Profiler) collectionLoop(ctx context.Context, subprocessDone <-chan error) error {
	ticker := time.NewTicker(time.Duration(p.cfg.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	var subprocessErr error

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-subprocessDone:
			// Subprocess finished - collect final sample and stop
			subprocessErr = err
			p.collectSample()
			return subprocessErr

		case <-ticker.C:
			p.collectSample()
		}
	}
}

// collectSample collects and exports a single metrics snapshot
func (p *Profiler) collectSample() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopRequested {
		return
	}

	metrics := p.manager.CollectDynamicMetrics()
	if err := p.exporter.WriteDynamicMetrics(metrics); err != nil {
		log.Printf("Warning: failed to write metrics: %v", err)
		return
	}

	p.sampleCount++

	// Progress indicator (every 100 samples)
	if p.sampleCount%100 == 0 {
		log.Printf("Collected %d samples", p.sampleCount)
	}
}

// setupSignalHandling creates a context that cancels on SIGINT/SIGTERM
func (p *Profiler) setupSignalHandling(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			log.Printf("Received signal %v, stopping...", sig)
			p.mu.Lock()
			p.stopRequested = true
			p.mu.Unlock()
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigChan)
	}()

	return ctx, cancel
}

// generateGraphs creates HTML visualization from collected data
func (p *Profiler) generateGraphs() error {
	outputPath := p.exporter.OutputPath()
	graphPath := filepath.Join(p.cfg.OutputDir, fmt.Sprintf("report_%s.html", p.exporter.SessionUUID()))

	// Load records from output file
	loader, err := output.NewRecordLoader(outputPath)
	if err != nil {
		return err
	}

	records, err := loader.Load(outputPath)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return fmt.Errorf("no records to generate graphs from")
	}

	return output.GenerateHTML(graphPath, p.exporter.SessionUUID().String(), p.staticMetrics, records)
}

// printStartupInfo logs profiler configuration at startup
func (p *Profiler) printStartupInfo() {
	log.Printf("=== System Profiler Started ===")
	log.Printf("Session UUID: %s", p.exporter.SessionUUID())
	log.Printf("Output directory: %s", p.cfg.OutputDir)
	log.Printf("Format: %s", p.cfg.Format)
	log.Printf("Interval: %dms", p.cfg.IntervalMs)
	log.Printf("Mode: %s", p.cfg.ModeName())
	log.Printf("Active collectors: %d (%s)", p.manager.CollectorCount(),
		formatCollectorList(p.manager.CollectorNames()))

	if p.staticMetrics != nil {
		log.Printf("Host: %s (%s)", p.staticMetrics.Hostname, p.staticMetrics.Hostname)
		log.Printf("CPUs: %d x %s", p.staticMetrics.NumProcessors, p.staticMetrics.CPUType)
	}
}

// printSummary logs profiling results at completion
func (p *Profiler) printSummary() {
	duration := time.Since(p.startTime)

	log.Printf("=== Profiling Complete ===")
	log.Printf("Duration: %s", duration.Round(time.Millisecond))
	log.Printf("Samples collected: %d", p.sampleCount)
	log.Printf("Output file: %s", p.exporter.OutputPath())

	if p.cfg.GenerateGraphs {
		log.Printf("Report: %s", filepath.Join(p.cfg.OutputDir,
			fmt.Sprintf("report_%s.html", p.exporter.SessionUUID())))
	}
}

// Close releases resources held by the profiler
func (p *Profiler) Close() error {
	if p.manager != nil {
		p.manager.Close()
	}
	return nil
}

// Statistics returns current profiling statistics
type Statistics struct {
	SessionUUID   string
	SampleCount   int
	Duration      time.Duration
	OutputPath    string
	StaticMetrics *collectors.StaticMetrics
}

// Stats returns current profiling statistics
func (p *Profiler) Stats() Statistics {
	p.mu.Lock()
	defer p.mu.Unlock()

	return Statistics{
		SessionUUID:   p.exporter.SessionUUID().String(),
		SampleCount:   p.sampleCount,
		Duration:      time.Since(p.startTime),
		OutputPath:    p.exporter.OutputPath(),
		StaticMetrics: p.staticMetrics,
	}
}

// formatCollectorList formats collector names for display
func formatCollectorList(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	if len(names) <= 5 {
		return fmt.Sprintf("%v", names)
	}
	return fmt.Sprintf("%v... and %d more", names[:5], len(names)-5)
}
