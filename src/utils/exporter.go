package utils

import (
	"InferenceProfiler/src/collectors"
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
	snapshotFiles []string
}

// NewExporter creates a new exporter
func NewExporter(outputDir string, sessionUUID uuid.UUID) (*Exporter, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	return &Exporter{outputDir: outputDir, sessionUUID: sessionUUID}, nil
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

// SaveSnapshot saves a metrics snapshot as JSON
func (e *Exporter) SaveSnapshot(metrics *collectors.DynamicMetrics) error {
	filename := fmt.Sprintf("%s-%d.json", e.sessionUUID, metrics.Timestamp)
	path := filepath.Join(e.outputDir, filename)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	// Flatten before saving so snapshots have consistent format
	flat := FlattenMetrics(metrics)
	if err := json.NewEncoder(file).Encode(flat); err != nil {
		return fmt.Errorf("failed to encode metrics: %w", err)
	}
	e.snapshotFiles = append(e.snapshotFiles, path)
	return nil
}

// ProcessSession aggregates snapshots into final format
func (e *Exporter) ProcessSession(format string) error {
	if len(e.snapshotFiles) == 0 {
		log.Println("No snapshots captured to process")
		return nil
	}

	log.Printf("Aggregating %d snapshots...", len(e.snapshotFiles))
	sort.Strings(e.snapshotFiles)

	// Load all flattened records
	var allRecords []map[string]interface{}
	for _, filePath := range e.snapshotFiles {
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Skipping file %s: %v", filePath, err)
			continue
		}

		var data map[string]interface{}
		if err := json.NewDecoder(file).Decode(&data); err != nil {
			file.Close()
			log.Printf("Skipping corrupt file %s: %v", filePath, err)
			continue
		}
		file.Close()
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
	default:
		return fmt.Errorf("unsupported format: %s (use 'jsonl' or 'parquet')", format)
	}
}

// exportJSONL writes flattened records as JSON Lines
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

	writer := parquet.NewGenericWriter[any](file, schema)

	// Convert maps to parquet rows
	rows := make([]parquet.Row, len(records))
	for i, record := range records {
		rowValues := make([]parquet.Value, 0, len(keys))

		for j, key := range keys {
			val, exists := record[key]
			if !exists || val == nil {
				rowValues = append(rowValues, parquet.ValueOf(nil).Level(0, 0, j))
				continue
			}

			pVal := convertToParquetValue(val, nodeMap[key])
			rowValues = append(rowValues, pVal.Level(0, 1, j))
		}
		rows[i] = rowValues
	}

	if _, err := writer.WriteRows(rows); err != nil {
		return fmt.Errorf("failed to write parquet rows: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close parquet writer: %w", err)
	}

	log.Printf("Exported Parquet: %s (%d records)", path, len(records))
	return nil
}

// inferParquetType determines the parquet type for a Go value
func inferParquetType(v interface{}) parquet.Node {
	switch v.(type) {
	case int, int32, int64, float64:
		// JSON numbers are float64, but we want int64 for most metrics
		return parquet.Int(64)
	case bool:
		return parquet.Leaf(parquet.BooleanType)
	default:
		return parquet.String()
	}
}

// convertToParquetValue converts a Go value to a parquet.Value
func convertToParquetValue(val interface{}, node parquet.Node) parquet.Value {
	if node == nil {
		return parquet.ValueOf(fmt.Sprintf("%v", val))
	}

	switch node.Type().Kind() {
	case parquet.Int64:
		switch v := val.(type) {
		case float64:
			return parquet.ValueOf(int64(v))
		case int64:
			return parquet.ValueOf(v)
		case int:
			return parquet.ValueOf(int64(v))
		default:
			return parquet.ValueOf(int64(0))
		}
	case parquet.Double:
		switch v := val.(type) {
		case float64:
			return parquet.ValueOf(v)
		case int64:
			return parquet.ValueOf(float64(v))
		default:
			return parquet.ValueOf(0.0)
		}
	case parquet.Boolean:
		if v, ok := val.(bool); ok {
			return parquet.ValueOf(v)
		}
		return parquet.ValueOf(false)
	default:
		return parquet.ValueOf(fmt.Sprintf("%v", val))
	}
}

// Cleanup removes intermediate snapshot files
func (e *Exporter) Cleanup() {
	for _, path := range e.snapshotFiles {
		os.Remove(path)
	}
}
