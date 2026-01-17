// Package exporting provides the service for persisting metrics data.
package exporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"InferenceProfiler/pkg/formatting"
	"InferenceProfiler/pkg/metrics"
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

// Write writes a single record.
func (e *Exporter) Write(record metrics.Record) error {
	return e.writer.Write(record)
}

// WriteBatch writes multiple records.
func (e *Exporter) WriteBatch(records []metrics.Record) error {
	for i, r := range records {
		if err := e.writer.Write(r); err != nil {
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
func (e *Exporter) WriteStatic(m *metrics.Static) error {
	dir := filepath.Dir(e.path)
	base := filepath.Base(e.path)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	staticPath := filepath.Join(dir, name+"_static.json")

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal static metrics: %w", err)
	}

	return os.WriteFile(staticPath, data, 0644)
}

// Flatten converts dynamic metrics to flat records for export.
func Flatten(m *metrics.Dynamic) []formatting.Record {
	// Convert to JSON and back for a generic map
	data, _ := json.Marshal(m)
	var record formatting.Record
	json.Unmarshal(data, &record)

	// Remove nested structures that should be flattened differently
	delete(record, "NvidiaGPUs")
	delete(record, "Processes")

	return []formatting.Record{record}
}
