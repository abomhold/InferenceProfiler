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
	streaming     bool   // true = stream mode, false = batch mode
	format        string // export format: jsonl, parquet, csv, tsv
	snapshotFiles []string

	// Stream state
	streamFile    *os.File
	parquetWriter *parquet.Writer
	parquetSchema *parquet.Schema
	csvWriter     *csv.Writer
	jsonlEncoder  *json.Encoder
	schemaKeys    []string // for CSV/TSV column order
	snapshotCount int64
}

// NewExporter creates a new exporter
// If streaming is true, format must be one of: jsonl, parquet, csv, tsv
func NewExporter(outputDir string, sessionUUID uuid.UUID, flatten bool, streaming bool, format string) (*Exporter, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Validate format for streaming mode
	if streaming {
		validStreamFormats := map[string]bool{"jsonl": true, "parquet": true, "csv": true, "tsv": true}
		if !validStreamFormats[format] {
			return nil, fmt.Errorf("streaming mode only supports jsonl, parquet, csv, tsv (got: %s)", format)
		}
	}

	return &Exporter{
		outputDir:   outputDir,
		sessionUUID: sessionUUID,
		flatten:     flatten,
		streaming:   streaming,
		format:      format,
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
// In streaming mode, writes directly to the output file
// In batch mode, saves as intermediate JSON files
func (e *Exporter) SaveSnapshot(metrics *collectors.DynamicMetrics) error {
	if e.streaming {
		return e.writeToStream(metrics)
	}

	// Batch mode: save as intermediate JSON
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

// writeToStream writes a snapshot to the appropriate stream based on format
func (e *Exporter) writeToStream(metrics *collectors.DynamicMetrics) error {
	// Convert to flat map (streaming always uses flattened format for consistency)
	flatData := FlattenMetrics(metrics)

	// Initialize stream on first snapshot
	if e.streamFile == nil {
		if err := e.initStream(flatData); err != nil {
			return fmt.Errorf("failed to initialize stream: %w", err)
		}
	}

	// Write to appropriate stream
	switch e.format {
	case "jsonl":
		if err := e.jsonlEncoder.Encode(flatData); err != nil {
			return fmt.Errorf("failed to write jsonl record: %w", err)
		}
	case "parquet":
		if err := e.parquetWriter.Write(flatData); err != nil {
			return fmt.Errorf("failed to write parquet row: %w", err)
		}
	case "csv", "tsv":
		row := make([]string, len(e.schemaKeys))
		for i, key := range e.schemaKeys {
			if val, exists := flatData[key]; exists && val != nil {
				row[i] = formatValue(val)
			}
		}
		if err := e.csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write csv/tsv row: %w", err)
		}
	}

	e.snapshotCount++
	return nil
}

// initStream initializes the stream writer based on format
func (e *Exporter) initStream(firstSnapshot map[string]interface{}) error {
	// Extract and sort keys for consistent column order
	var keys []string
	for k := range firstSnapshot {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	e.schemaKeys = keys

	// Create output file
	ext := e.format
	if ext == "csv" || ext == "tsv" {
		ext = e.format
	}
	path := filepath.Join(e.outputDir, fmt.Sprintf("%s.%s", e.sessionUUID, ext))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create stream file: %w", err)
	}
	e.streamFile = file

	// Initialize format-specific writer
	switch e.format {
	case "jsonl":
		e.jsonlEncoder = json.NewEncoder(file)
		log.Printf("Initialized JSONL stream: %s", path)

	case "parquet":
		// Build parquet schema
		nodeMap := make(map[string]parquet.Node)
		for k, v := range firstSnapshot {
			if v != nil {
				nodeMap[k] = inferParquetType(v)
			}
		}

		group := make(parquet.Group)
		for _, k := range keys {
			node, ok := nodeMap[k]
			if !ok {
				node = parquet.String()
			}
			group[k] = parquet.Optional(node)
		}

		e.parquetSchema = parquet.NewSchema("InferenceMetrics", group)
		e.parquetWriter = parquet.NewWriter(file, e.parquetSchema)
		log.Printf("Initialized Parquet stream: %s", path)

	case "csv":
		e.csvWriter = csv.NewWriter(file)
		e.csvWriter.Comma = ','
		if err := e.csvWriter.Write(keys); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
		log.Printf("Initialized CSV stream: %s", path)

	case "tsv":
		e.csvWriter = csv.NewWriter(file)
		e.csvWriter.Comma = '\t'
		if err := e.csvWriter.Write(keys); err != nil {
			return fmt.Errorf("failed to write TSV header: %w", err)
		}
		log.Printf("Initialized TSV stream: %s", path)
	}

	return nil
}

// CloseStream finalizes the stream if open
func (e *Exporter) CloseStream() error {
	if !e.streaming || e.streamFile == nil {
		return nil
	}

	var err error

	// Flush format-specific writers
	switch e.format {
	case "parquet":
		if e.parquetWriter != nil {
			if err = e.parquetWriter.Close(); err != nil {
				return fmt.Errorf("failed to close parquet writer: %w", err)
			}
			e.parquetWriter = nil
		}
	case "csv", "tsv":
		if e.csvWriter != nil {
			e.csvWriter.Flush()
			if err = e.csvWriter.Error(); err != nil {
				return fmt.Errorf("failed to flush csv writer: %w", err)
			}
			e.csvWriter = nil
		}
	}

	// Close file
	if e.streamFile != nil {
		if err = e.streamFile.Close(); err != nil {
			return fmt.Errorf("failed to close stream file: %w", err)
		}
		e.streamFile = nil
	}

	log.Printf("Closed %s stream (%d records written)", e.format, e.snapshotCount)
	return nil
}

// ProcessSession aggregates snapshots into final format
// Only used in batch mode
func (e *Exporter) ProcessSession() error {
	if e.streaming {
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

	switch e.format {
	case "jsonl":
		return e.exportJSONL(basePath+".jsonl", allRecords)
	case "parquet":
		return e.exportParquet(basePath+".parquet", allRecords)
	case "csv":
		return e.exportDelimited(basePath+".csv", allRecords, ',')
	case "tsv":
		return e.exportDelimited(basePath+".tsv", allRecords, '\t')
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
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
	case int, int32, int64, uint, uint32, uint64: // Added uint types
		return parquet.Int(64)
	case float64:
		return parquet.Leaf(parquet.DoubleType)
	case json.Number:
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
