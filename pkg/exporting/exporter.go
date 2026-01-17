// Package exporting provides the service for persisting metrics data.
package exporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"InferenceProfiler/pkg/formatting"
)

// Exporter handles writing metrics to various output formats.
type Exporter struct {
	path   string
	format string
	writer formatting.Writer
}

// NewExporter creates a new exporter for the given path and format.
func NewExporter(path, format string) (*Exporter, error) {
	// Ensure output directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Get format handler
	f, ok := formatting.Get(format)
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// Create writer
	writer := f.Writer()
	if err := writer.Init(path, nil); err != nil {
		return nil, fmt.Errorf("failed to initialize writer: %w", err)
	}

	return &Exporter{
		path:   path,
		format: format,
		writer: writer,
	}, nil
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

// Write writes a single record, flattening any deferred slice data.
func (e *Exporter) Write(record formatting.Record) error {
	return e.writer.Write(formatting.FlattenRecord(record))
}

// WriteBatch writes multiple records, flattening each.
func (e *Exporter) WriteBatch(records []formatting.Record) error {
	for i, r := range records {
		if err := e.writer.Write(formatting.FlattenRecord(r)); err != nil {
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
func (e *Exporter) WriteStatic(record formatting.Record) error {
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
