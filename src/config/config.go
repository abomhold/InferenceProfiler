package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

// Config holds all profiler configuration
type Config struct {
	// Session identifier
	SessionUUID uuid.UUID

	// Output settings
	OutputDir string
	Format    string
	Flatten   bool
	Stream    bool
	Cleanup   bool

	// Sampling
	IntervalMs int

	// Graph generation
	GenerateGraphs bool
	InputFile      string // Path to existing file for graph-only mode

	// Collector toggles
	Collectors CollectorConfig

	// Subprocess command (if any)
	Subprocess []string
}

// CollectorConfig controls which collectors are enabled
type CollectorConfig struct {
	CPU         bool
	Memory      bool
	Disk        bool
	Network     bool
	Container   bool
	Processes   bool
	Nvidia      bool
	NvidiaProcs bool
	VLLM        bool
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		SessionUUID:    uuid.New(),
		OutputDir:      "./profiler-output",
		Format:         "jsonl",
		Flatten:        true,
		Stream:         false,
		Cleanup:        true,
		IntervalMs:     100,
		GenerateGraphs: true,
		InputFile:      "",
		Collectors: CollectorConfig{
			CPU:         true,
			Memory:      true,
			Disk:        true,
			Network:     true,
			Container:   true,
			Processes:   true,
			Nvidia:      true,
			NvidiaProcs: true,
			VLLM:        true,
		},
	}
}

// ParseFlags parses command-line flags into a Config struct
func ParseFlags() (*Config, error) {
	cfg := DefaultConfig()

	// Output options
	flag.StringVar(&cfg.OutputDir, "o", cfg.OutputDir, "Output directory for logs and exported data")
	flag.IntVar(&cfg.IntervalMs, "t", cfg.IntervalMs, "Sampling interval in milliseconds")
	flag.StringVar(&cfg.Format, "f", cfg.Format, "Export format: jsonl, parquet, csv, tsv")
	flag.BoolVar(&cfg.Stream, "stream", cfg.Stream, "Stream mode: write directly to output file")

	// These are inverted flags (no-X disables X)
	noFlatten := flag.Bool("no-flatten", false, "Disable flattening nested data (GPUs, processes) to columns")
	noCleanup := flag.Bool("no-cleanup", false, "Keep intermediary JSON snapshot files after final export")

	// Graph options
	flag.BoolVar(&cfg.GenerateGraphs, "graphs", cfg.GenerateGraphs, "Generate HTML graphs after profiling")
	flag.StringVar(&cfg.InputFile, "i", "", "Input file for graph-only mode (skips profiling)")
	graphOnlyFlag := flag.Bool("g", false, "Graph-only mode: generate graphs from -i file without profiling")

	// Collector toggles (inverted)
	noCPU := flag.Bool("no-cpu", false, "Disable CPU metrics collection")
	noMemory := flag.Bool("no-memory", false, "Disable memory metrics collection")
	noDisk := flag.Bool("no-disk", false, "Disable disk I/O metrics collection")
	noNetwork := flag.Bool("no-network", false, "Disable network metrics collection")
	noContainer := flag.Bool("no-container", false, "Disable container/cgroup metrics collection")
	noProcs := flag.Bool("no-procs", false, "Disable per-process metrics collection")
	noNvidia := flag.Bool("no-nvidia", false, "Disable all NVIDIA GPU metrics collection")
	noGPUProcs := flag.Bool("no-gpu-procs", false, "Disable GPU process enumeration (still collects GPU metrics)")
	noVLLM := flag.Bool("no-vllm", false, "Disable vLLM metrics collection")

	flag.Parse()

	// Apply inverted flags
	cfg.Flatten = !*noFlatten
	cfg.Cleanup = !*noCleanup
	cfg.Collectors.CPU = !*noCPU
	cfg.Collectors.Memory = !*noMemory
	cfg.Collectors.Disk = !*noDisk
	cfg.Collectors.Network = !*noNetwork
	cfg.Collectors.Container = !*noContainer
	cfg.Collectors.Processes = !*noProcs
	cfg.Collectors.Nvidia = !*noNvidia
	cfg.Collectors.NvidiaProcs = !*noGPUProcs
	cfg.Collectors.VLLM = !*noVLLM

	// Graph-only mode requires -g and -i
	if *graphOnlyFlag && cfg.InputFile == "" {
		return nil, errors.New("-g (graph-only) requires -i <input file>")
	}

	// Remaining args are the subprocess command
	cfg.Subprocess = flag.Args()

	// Streaming always uses flattened format
	if cfg.Stream {
		cfg.Flatten = true
	}

	return cfg, cfg.Validate()
}

// Validate checks configuration for errors
func (c *Config) Validate() error {
	var errs []error

	// Validate format
	validFormats := map[string]bool{
		"jsonl":   true,
		"parquet": true,
		"csv":     true,
		"tsv":     true,
	}
	if !validFormats[c.Format] {
		errs = append(errs, fmt.Errorf("invalid format %q: must be one of jsonl, parquet, csv, tsv", c.Format))
	}

	// Validate interval
	if c.IntervalMs < 1 {
		errs = append(errs, errors.New("interval must be >= 1ms"))
	}

	// Validate output directory is writable (if not graph-only mode)
	if !c.IsGraphOnlyMode() {
		if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
			errs = append(errs, fmt.Errorf("cannot create output directory %q: %w", c.OutputDir, err))
		}
	}

	// Validate input file exists for graph-only mode
	if c.InputFile != "" {
		if _, err := os.Stat(c.InputFile); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("input file does not exist: %s", c.InputFile))
		}
	}

	return errors.Join(errs...)
}

// IsGraphOnlyMode returns true if we should only generate graphs
func (c *Config) IsGraphOnlyMode() bool {
	return c.InputFile != ""
}

// HasSubprocess returns true if a subprocess command was provided
func (c *Config) HasSubprocess() bool {
	return len(c.Subprocess) > 0
}

// ModeName returns "streaming" or "batch" for display
func (c *Config) ModeName() string {
	if c.Stream {
		return "streaming"
	}
	return "batch"
}

// String returns a human-readable configuration summary
func (c *Config) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Output Dir:   %s\n", c.OutputDir))
	b.WriteString(fmt.Sprintf("Interval:     %dms\n", c.IntervalMs))
	b.WriteString(fmt.Sprintf("Format:       %s\n", c.Format))
	b.WriteString(fmt.Sprintf("Mode:         %s\n", c.ModeName()))
	b.WriteString(fmt.Sprintf("Flatten:      %v\n", c.Flatten))
	if !c.Stream {
		b.WriteString(fmt.Sprintf("Cleanup:      %v\n", c.Cleanup))
	}
	if c.GenerateGraphs {
		b.WriteString("Graphs:       enabled\n")
	}
	return b.String()
}
