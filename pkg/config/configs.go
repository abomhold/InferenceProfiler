// Package config provides configuration management for the profiler.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"InferenceProfiler/pkg/formatting"
)

// Record is a map of string keys to arbitrary values.
type Record = map[string]interface{}

// Config holds all profiler configuration options.
type Config struct {
	// Collection settings
	Interval            time.Duration
	Duration            time.Duration
	EnableVM            bool
	EnableContainer     bool
	EnableProcess       bool
	EnableNvidia        bool
	EnableVLLM          bool
	CollectGPUProcesses bool

	// Output settings
	OutputDir    string
	OutputFormat string
	OutputName   string

	// Graph settings
	GenerateGraphs bool
	GraphFormat    string
	GraphOutput    string

	// System identification
	UUID     string
	VMID     string
	Hostname string
}

// Default configuration values.
const (
	DefaultInterval    = 100 * time.Millisecond
	DefaultOutputDir   = "."
	DefaultFormat      = "parquet"
	DefaultGraphFormat = "html"
)

// New creates a Config with default values.
func New() *Config {
	hostname, _ := os.Hostname()

	return &Config{
		Interval:            DefaultInterval,
		Duration:            0,
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       false,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
		OutputDir:           DefaultOutputDir,
		OutputFormat:        DefaultFormat,
		GenerateGraphs:      false,
		GraphFormat:         DefaultGraphFormat,
		Hostname:            hostname,
		UUID:                fmt.Sprintf("%d", time.Now().UnixNano()),
	}
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Interval < time.Millisecond {
		return fmt.Errorf("interval must be at least 1ms, got %v", c.Interval)
	}

	if c.Duration < 0 {
		return fmt.Errorf("duration cannot be negative, got %v", c.Duration)
	}

	if !isValidOutputFormat(c.OutputFormat) {
		return fmt.Errorf("invalid output format: %s (valid: parquet, jsonl, csv, tsv)", c.OutputFormat)
	}

	if c.GenerateGraphs && !isValidGraphFormat(c.GraphFormat) {
		return fmt.Errorf("invalid graph format: %s (valid: html, png, svg)", c.GraphFormat)
	}

	if c.OutputDir != "" {
		if info, err := os.Stat(c.OutputDir); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("cannot access output directory: %w", err)
			}
		} else if !info.IsDir() {
			return fmt.Errorf("output path is not a directory: %s", c.OutputDir)
		}
	}

	return nil
}

// ValidOutputFormats returns the list of supported output formats.
func ValidOutputFormats() []string {
	return []string{"parquet", "jsonl", "csv", "tsv"}
}

// ValidGraphFormats returns the list of supported graph formats.
func ValidGraphFormats() []string {
	return []string{"html", "png", "svg"}
}

func isValidOutputFormat(format string) bool {
	for _, f := range ValidOutputFormats() {
		if f == format {
			return true
		}
	}
	return false
}

func isValidGraphFormat(format string) bool {
	for _, f := range ValidGraphFormats() {
		if f == format {
			return true
		}
	}
	return false
}

// ApplyDefaults fills in any missing values with defaults.
func (c *Config) ApplyDefaults() {
	if c.Interval == 0 {
		c.Interval = DefaultInterval
	}
	if c.OutputDir == "" {
		c.OutputDir = DefaultOutputDir
	}
	if c.OutputFormat == "" {
		c.OutputFormat = DefaultFormat
	}
	if c.GraphFormat == "" {
		c.GraphFormat = DefaultGraphFormat
	}
	if c.Hostname == "" {
		c.Hostname, _ = os.Hostname()
	}
	if c.UUID == "" {
		c.UUID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
}

// BuildBaseStatic creates the base static metrics as a Record.
func (c *Config) BuildBaseStatic() Record {
	return Record{
		"uuid":      c.UUID,
		"vId":       c.VMID,
		"vHostname": c.Hostname,
		"vBootTime": GetBootTime(),
	}
}

// GenerateOutputPath creates an auto-generated output path.
func (c *Config) GenerateOutputPath(prefix string) string {
	timestamp := time.Now().Format("20060102-150405")
	ext := formatting.GetExtension(c.OutputFormat)
	return filepath.Join(c.OutputDir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

// GenerateGraphPath creates an auto-generated graph output path.
func (c *Config) GenerateGraphPath(inputFile string) string {
	if c.GraphOutput != "" {
		return c.GraphOutput
	}

	dir := filepath.Dir(inputFile)
	base := filepath.Base(inputFile)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	graphExt := formatting.GetGraphExtension(c.GraphFormat)
	return filepath.Join(dir, name+"_graphs"+graphExt)
}

// GetBootTime reads the system boot time from /proc/stat.
func GetBootTime() int64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0
	}

	for _, line := range formatting.SplitLines(string(data)) {
		if len(line) > 6 && line[:6] == "btime " {
			var bootTime int64
			fmt.Sscanf(line[6:], "%d", &bootTime)
			return bootTime * 1_000_000_000 // Convert to nanoseconds
		}
	}
	return 0
}

// MergeRecords combines multiple Record maps into one.
func MergeRecords(maps ...Record) Record {
	result := make(Record)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
