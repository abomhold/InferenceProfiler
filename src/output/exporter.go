package output

import (
	"InferenceProfiler/src/collectors"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/parquet-go/parquet-go"
)

// Exporter handles saving metrics to files
type Exporter struct {
	outputDir     string
	sessionUUID   uuid.UUID
	flatten       bool
	snapshotFiles []string
	streamParquet bool
	parquetWriter *parquet.Writer
	parquetFile   *os.File
	parquetSchema *parquet.Schema
	snapshotCount int64
}

// NewExporter creates a new exporter
func NewExporter(outputDir string, sessionUUID uuid.UUID, flatten bool, streamParquet bool) (*Exporter, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create aggregate directory: %w", err)
	}
	return &Exporter{
		outputDir:     outputDir,
		sessionUUID:   sessionUUID,
		flatten:       flatten,
		streamParquet: streamParquet,
	}, nil
}

// SaveStatic saves static system information as JSON
func (e *Exporter) SaveStatic(data *collectors.StaticMetrics) error {
	path := filepath.Join(e.outputDir, fmt.Sprintf("%s.json", e.sessionUUID))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create static file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode static data: %w", err)
	}
	log.Printf("Saved static info: %s", path)
	return nil
}

// SaveSnapshot saves a metrics snapshot
// In streaming mode, writes directly to parquet file
// In normal mode, saves as JSON
func (e *Exporter) SaveSnapshot(metrics *collectors.DynamicMetrics) error {
	if e.streamParquet {
		return e.writeToParquetStream(metrics)
	}

	// Normal mode: save as JSON
	filename := fmt.Sprintf("%s-%d.json", e.sessionUUID, metrics.Timestamp)
	path := filepath.Join(e.outputDir, filename)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	// Use flatten mode or JSON mode based on config
	var data interface{}
	if e.flatten {
		data = FlattenMetrics(metrics)
	} else {
		data = ToJSONMode(metrics)
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("failed to encode metrics: %w", err)
	}
	e.snapshotFiles = append(e.snapshotFiles, path)
	return nil
}

// writeToParquetStream writes a single snapshot to the parquet stream
func (e *Exporter) writeToParquetStream(metrics *collectors.DynamicMetrics) error {
	// Convert to flat map (streaming parquet always uses flattened format)
	flatData := FlattenMetrics(metrics)

	// Initialize parquet writer on first snapshot
	if e.parquetWriter == nil {
		if err := e.initParquetStream(flatData); err != nil {
			return fmt.Errorf("failed to initialize parquet stream: %w", err)
		}
	}

	// Write the map directly - the schema will deconstruct it
	if err := e.parquetWriter.Write(flatData); err != nil {
		return fmt.Errorf("failed to write parquet row: %w", err)
	}

	e.snapshotCount++
	return nil
}

// initParquetStream initializes the parquet writer with schema from first snapshot
func (e *Exporter) initParquetStream(firstSnapshot map[string]interface{}) error {
	// Build schema from first snapshot
	nodeMap := make(map[string]parquet.Node)
	var keys []string

	for k, v := range firstSnapshot {
		keys = append(keys, k)
		if v != nil {
			nodeMap[k] = inferParquetType(v)
		}
	}
	sort.Strings(keys)

	// Build parquet schema
	group := make(parquet.Group)
	for _, k := range keys {
		node, ok := nodeMap[k]
		if !ok {
			node = parquet.String()
		}
		group[k] = parquet.Optional(node)
	}

	e.parquetSchema = parquet.NewSchema("InferenceMetrics", group)

	// Create parquet file
	path := filepath.Join(e.outputDir, fmt.Sprintf("%s-stream.parquet", e.sessionUUID))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	e.parquetFile = file

	// Create writer using lower-level API
	e.parquetWriter = parquet.NewWriter(file, e.parquetSchema)

	log.Printf("Initialized parquet stream: %s", path)
	return nil
}

// CloseStream finalizes the parquet stream if open
func (e *Exporter) CloseStream() error {
	if e.parquetWriter != nil {
		if err := e.parquetWriter.Close(); err != nil {
			return fmt.Errorf("failed to close parquet writer: %w", err)
		}
		e.parquetWriter = nil
		log.Printf("Closed parquet stream (%d records written)", e.snapshotCount)
	}
	if e.parquetFile != nil {
		if err := e.parquetFile.Close(); err != nil {
			return fmt.Errorf("failed to close parquet file: %w", err)
		}
		e.parquetFile = nil
	}
	return nil
}

// ProcessSession aggregates snapshots into final format
// Not used in streaming mode
func (e *Exporter) ProcessSession(format string) error {
	if e.streamParquet {
		// In streaming mode, just close the stream
		return e.CloseStream()
	}

	if len(e.snapshotFiles) == 0 {
		log.Println("No snapshots captured to process")
		return nil
	}

	log.Printf("Aggregating %d snapshots...", len(e.snapshotFiles))
	sort.Strings(e.snapshotFiles)

	// Load all records
	var allRecords []map[string]interface{}
	for _, filePath := range e.snapshotFiles {
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Skipping file %s: %v", filePath, err)
			continue
		}

		decoder := json.NewDecoder(file)
		decoder.UseNumber() // Keeps numbers as json.Number

		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			file.Close()
			log.Printf("Skipping corrupt file %s: %v", filePath, err)
			continue
		}
		file.Close()

		convertJSONNumbers(data)

		allRecords = append(allRecords, data)
	}

	if len(allRecords) == 0 {
		return nil
	}

	basePath := filepath.Join(e.outputDir, e.sessionUUID.String())

	switch format {
	case "jsonl":
		return e.exportJSONL(basePath+".jsonl", allRecords)
	case "parquet":
		return e.exportParquet(basePath+".parquet", allRecords)
	case "csv":
		return e.exportDelimited(basePath+".csv", allRecords, ',')
	case "tsv":
		return e.exportDelimited(basePath+".tsv", allRecords, '\t')
	default:
		return fmt.Errorf("unsupported format: %s (use 'jsonl', 'parquet', 'csv', or 'tsv')", format)
	}
}

// exportJSONL writes records as JSON Lines
func (e *Exporter) exportJSONL(path string, records []map[string]interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create JSONL file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("failed to encode record: %w", err)
		}
	}

	log.Printf("Exported JSONL: %s (%d records)", path, len(records))
	return nil
}

// exportDelimited writes records as CSV or TSV
func (e *Exporter) exportDelimited(path string, records []map[string]interface{}, delimiter rune) error {
	if len(records) == 0 {
		return nil
	}

	// Build column list from all records (union of all keys)
	keySet := make(map[string]bool)
	for _, record := range records {
		for k := range record {
			keySet[k] = true
		}
	}

	// Sort keys for deterministic column order
	var keys []string
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = delimiter

	// Write header
	if err := writer.Write(keys); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, record := range records {
		row := make([]string, len(keys))
		for i, key := range keys {
			if val, exists := record[key]; exists && val != nil {
				row[i] = formatValue(val)
			}
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("csv writer error: %w", err)
	}

	formatName := "CSV"
	if delimiter == '\t' {
		formatName = "TSV"
	}
	log.Printf("Exported %s: %s (%d records, %d columns)", formatName, path, len(records), len(keys))
	return nil
}

// formatValue converts a value to string for CSV/TSV output
func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case float64:
		// Check if it's actually an integer
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		// For complex types (arrays, objects), serialize to JSON
		if data, err := json.Marshal(v); err == nil {
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	}
}

// exportParquet writes records to Parquet with dynamic schema
func (e *Exporter) exportParquet(path string, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	// Build schema from all records (union of all keys)
	nodeMap := make(map[string]parquet.Node)
	keySet := make(map[string]bool)

	for _, record := range records {
		for k, v := range record {
			keySet[k] = true
			if _, exists := nodeMap[k]; !exists && v != nil {
				nodeMap[k] = inferParquetType(v)
			}
		}
	}

	// Sort keys for deterministic column order
	var keys []string
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build parquet schema
	group := make(parquet.Group)
	for _, k := range keys {
		node, ok := nodeMap[k]
		if !ok {
			node = parquet.String()
		}
		group[k] = parquet.Optional(node)
	}

	schema := parquet.NewSchema("InferenceMetrics", group)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	defer file.Close()

	writer := parquet.NewWriter(file, schema)

	// Write maps directly - the schema will deconstruct them
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write parquet row: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close parquet writer: %w", err)
	}

	log.Printf("Exported Parquet: %s (%d records)", path, len(records))
	return nil
}

// inferParquetType determines the parquet type for a Go value
func inferParquetType(v interface{}) parquet.Node {
	switch val := v.(type) {
	case int, int32, int64:
		return parquet.Int(64)
	case float64:
		return parquet.Leaf(parquet.DoubleType)
	case json.Number:
		// FIX: Try to parse as Int64 first, fallback to Float64
		if _, err := val.Int64(); err == nil {
			return parquet.Int(64)
		}
		return parquet.Leaf(parquet.DoubleType)
	case bool:
		return parquet.Leaf(parquet.BooleanType)
	default:
		return parquet.String()
	}
}
func convertJSONNumbers(v interface{}) interface{} {
	switch val := v.(type) {
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return i
		}
		if f, err := val.Float64(); err == nil {
			return f
		}
		return val.String()
	case map[string]interface{}:
		for k, v := range val {
			val[k] = convertJSONNumbers(v)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = convertJSONNumbers(v)
		}
		return val
	}
	return v
}

// Cleanup removes intermediate snapshot files
func (e *Exporter) Cleanup() {
	for _, path := range e.snapshotFiles {
		os.Remove(path)
	}
}
