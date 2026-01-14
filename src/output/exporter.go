package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"

	"InferenceProfiler/src/collectors"
)

// Exporter handles writing profiler metrics to various output formats.
// It supports both streaming mode (write directly to final format) and
// batch mode (collect snapshots, then convert).
type Exporter struct {
	outputDir     string
	sessionUUID   uuid.UUID
	format        string
	flatten       bool
	streaming     bool
	finalPath     string
	snapshotDir   string
	writer        FormatWriter
	schema        []string
	snapshotCount int
	mu            sync.Mutex
}

// ExporterOption configures an Exporter
type ExporterOption func(*Exporter)

// WithStreaming enables streaming mode (write directly to output)
func WithStreaming() ExporterOption {
	return func(e *Exporter) {
		e.streaming = true
		e.flatten = true // streaming always uses flattened format
	}
}

// WithFlatten enables/disables record flattening
func WithFlatten(enabled bool) ExporterOption {
	return func(e *Exporter) {
		e.flatten = enabled
	}
}

// WithFormat sets the output format
func WithFormat(format string) ExporterOption {
	return func(e *Exporter) {
		e.format = format
	}
}

// NewExporter creates a new Exporter with the given options.
// Default: batch mode, jsonl format, flattening enabled.
func NewExporter(outputDir string, sessionUUID uuid.UUID, opts ...ExporterOption) (*Exporter, error) {
	e := &Exporter{
		outputDir:   outputDir,
		sessionUUID: sessionUUID,
		format:      "jsonl",
		flatten:     true,
		streaming:   false,
	}

	for _, opt := range opts {
		opt(e)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Setup paths
	e.finalPath = filepath.Join(outputDir, fmt.Sprintf("%s%s", sessionUUID.String(), extensionForFormat(e.format)))
	e.snapshotDir = filepath.Join(outputDir, "snapshots")

	// Initialize based on mode
	if e.streaming {
		if err := e.initStreamingMode(); err != nil {
			return nil, err
		}
	} else {
		if err := os.MkdirAll(e.snapshotDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
		}
	}

	return e, nil
}

// initStreamingMode sets up the format writer for streaming
func (e *Exporter) initStreamingMode() error {
	writer, err := NewFormatWriter(e.format)
	if err != nil {
		return err
	}
	e.writer = writer
	// Schema will be initialized on first write
	return nil
}

// OutputPath returns the path to the final output file
func (e *Exporter) OutputPath() string {
	return e.finalPath
}

// SessionUUID returns the session identifier
func (e *Exporter) SessionUUID() uuid.UUID {
	return e.sessionUUID
}

// WriteStaticMetrics writes static metrics to a JSON file.
// Called once at the start of profiling.
func (e *Exporter) WriteStaticMetrics(metrics *collectors.StaticMetrics) error {
	path := filepath.Join(e.outputDir, fmt.Sprintf("%s.json", e.sessionUUID.String()))

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal static metrics: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// WriteDynamicMetrics writes a dynamic metrics snapshot.
// In streaming mode, writes directly to final output.
// In batch mode, writes to a snapshot file for later conversion.
func (e *Exporter) WriteDynamicMetrics(metrics *collectors.DynamicMetrics) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streaming {
		return e.writeStreamingRecord(metrics)
	}
	return e.writeSnapshotFile(metrics)
}

// writeStreamingRecord writes a record directly to the final output
func (e *Exporter) writeStreamingRecord(metrics *collectors.DynamicMetrics) error {
	records := e.metricsToRecords(metrics)

	for _, record := range records {
		// Initialize writer with schema on first record
		if e.schema == nil {
			e.schema = ExtractSchema([]map[string]interface{}{record})
			if err := e.writer.Init(e.finalPath, e.schema); err != nil {
				return err
			}
		}

		if err := e.writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}
	return nil
}

// writeSnapshotFile writes metrics to a numbered snapshot file
func (e *Exporter) writeSnapshotFile(metrics *collectors.DynamicMetrics) error {
	path := filepath.Join(e.snapshotDir, fmt.Sprintf("snapshot_%06d.json", e.snapshotCount))
	e.snapshotCount++

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Finalize completes the export process.
// In batch mode, converts snapshots to final format.
// In streaming mode, closes the writer.
func (e *Exporter) Finalize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streaming {
		if e.writer != nil {
			return e.writer.Close()
		}
		return nil
	}

	return e.convertSnapshots()
}

// convertSnapshots reads all snapshot files and writes to final format
func (e *Exporter) convertSnapshots() error {
	// Read all snapshots
	entries, err := os.ReadDir(e.snapshotDir)
	if err != nil {
		return fmt.Errorf("failed to read snapshot directory: %w", err)
	}

	if len(entries) == 0 {
		return nil // Nothing to convert
	}

	// Collect all records
	var allRecords []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(e.snapshotDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var metrics collectors.DynamicMetrics
		if err := json.Unmarshal(data, &metrics); err != nil {
			continue
		}

		records := e.metricsToRecords(&metrics)
		allRecords = append(allRecords, records...)
	}

	if len(allRecords) == 0 {
		return nil
	}

	// Extract schema from all records
	schema := ExtractSchema(allRecords)

	// Create writer and write all records
	writer, err := NewFormatWriter(e.format)
	if err != nil {
		return err
	}

	if err := writer.Init(e.finalPath, schema); err != nil {
		return err
	}

	for _, record := range allRecords {
		if err := writer.Write(record); err != nil {
			writer.Close()
			return err
		}
	}

	return writer.Close()
}

// Cleanup removes temporary snapshot files
func (e *Exporter) Cleanup() error {
	if e.snapshotDir != "" {
		return os.RemoveAll(e.snapshotDir)
	}
	return nil
}

// metricsToRecords converts DynamicMetrics to flat records
func (e *Exporter) metricsToRecords(m *collectors.DynamicMetrics) []map[string]interface{} {
	if !e.flatten {
		// Return single record with nested structure
		data, _ := json.Marshal(m)
		var record map[string]interface{}
		json.Unmarshal(data, &record)
		return []map[string]interface{}{record}
	}

	// Flatten: create separate records for each GPU/process combination
	return FlattenMetrics(m)
}

// extensionForFormat returns the file extension for a format
func extensionForFormat(format string) string {
	switch format {
	case "jsonl":
		return ".jsonl"
	case "parquet":
		return ".parquet"
	case "csv":
		return ".csv"
	case "tsv":
		return ".tsv"
	default:
		return ".jsonl"
	}
}
