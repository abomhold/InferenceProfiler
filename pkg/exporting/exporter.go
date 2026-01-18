// Package exporting provides the service for persisting metrics data.
package exporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Exporter handles writing metrics to various output formats.
type Exporter struct {
	path        string
	format      string
	writer      Writer
	flattenMode FlattenMode
}

// ExporterOption configures an Exporter.
type ExporterOption func(*Exporter)

// WithFlattenMode sets the flattening mode for the exporter.
func WithFlattenMode(mode FlattenMode) ExporterOption {
	return func(e *Exporter) {
		e.flattenMode = mode
	}
}

// NewExporter creates a new exporter for the given path and format.
func NewExporter(path, format string, opts ...ExporterOption) (*Exporter, error) {
	// Ensure output directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Get format handler
	f, ok := Get(format)
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// Create writer
	writer := f.Writer()
	if err := writer.Init(path); err != nil {
		return nil, fmt.Errorf("failed to initialize writer: %w", err)
	}

	e := &Exporter{
		path:        path,
		format:      format,
		writer:      writer,
		flattenMode: FlattenDefault, // default: nvidia dynamic expanded, processes/static as JSON
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

// NewParquetExporter creates an exporter specifically for Parquet format.
func NewParquetExporter(path string) (*Exporter, error) {
	return NewExporter(path, "parquet")
}

// Path returns the output file path.
func (e *Exporter) Path() string {
	return e.path
}

// Format returns the output format.
func (e *Exporter) Format() string {
	return e.format
}

// Write writes a single record, flattening based on the configured mode.
func (e *Exporter) Write(record Record) error {
	return e.writer.Write(FlattenRecordWithMode(record, e.flattenMode))
}

// WriteBatch writes multiple records, flattening each based on the configured mode.
func (e *Exporter) WriteBatch(records []Record) error {
	for i, r := range records {
		if err := e.writer.Write(FlattenRecordWithMode(r, e.flattenMode)); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}
	return nil
}

// Flush ensures all buffered data is written.
func (e *Exporter) Flush() error {
	return e.writer.Flush()
}

// Close finalizes and closes the exporter.
func (e *Exporter) Close() error {
	return e.writer.Close()
}

// WriteStatic writes static metrics to a separate JSON file.
func (e *Exporter) WriteStatic(record Record) error {
	dir := filepath.Dir(e.path)
	base := filepath.Base(e.path)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	staticPath := filepath.Join(dir, name+"_static.json")

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal static metrics: %w", err)
	}

	return os.WriteFile(staticPath, data, 0644)
}
