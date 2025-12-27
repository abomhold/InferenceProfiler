package utils

import (
	"InferenceProfiler/src/collectors"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	return &Exporter{
		outputDir:   outputDir,
		sessionUUID: sessionUUID,
	}, nil
}

// SaveStatic saves static system information
func (e *Exporter) SaveStatic(data collectors.StaticMetrics) error {
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
func (e *Exporter) SaveSnapshot(metrics collectors.DynamicMetrics) error {
	ts, ok := metrics["timestamp"]
	if !ok {
		return fmt.Errorf("metrics missing timestamp")
	}
	timestamp, ok := ts.Value.(int64)
	if !ok {
		return fmt.Errorf("invalid timestamp type")
	}

	filename := fmt.Sprintf("%s-%d.json", e.sessionUUID, timestamp)
	path := filepath.Join(e.outputDir, filename)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(metrics); err != nil {
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

	// Sort files for consistent ordering
	sort.Strings(e.snapshotFiles)

	// Load all records
	var allRecords []map[string]interface{}
	for _, filePath := range e.snapshotFiles {
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Skipping file %s: %v", filePath, err)
			continue
		}

		var data collectors.DynamicMetrics
		if err := json.NewDecoder(file).Decode(&data); err != nil {
			file.Close()
			log.Printf("Skipping corrupt file %s: %v", filePath, err)
			continue
		}
		file.Close()

		flat := flattenMetrics(data)
		allRecords = append(allRecords, flat)
	}

	if len(allRecords) == 0 {
		return nil
	}

	basePath := filepath.Join(e.outputDir, e.sessionUUID.String())

	switch format {
	case "csv":
		return e.exportCSV(basePath+".csv", allRecords, ",")
	case "tsv":
		return e.exportCSV(basePath+".tsv", allRecords, "\t")
	case "parquet":
		return e.exportParquet(basePath+".parquet", allRecords)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (e *Exporter) exportCSV(path string, records []map[string]interface{}, delimiter string) error {
	if len(records) == 0 {
		return nil
	}

	// Collect all unique keys
	keySet := make(map[string]bool)
	for _, record := range records {
		for k := range record {
			keySet[k] = true
		}
	}

	// Sort keys for consistent output
	var keys []string
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if delimiter == "\t" {
		writer.Comma = '\t'
	}

	// Write header
	if err := writer.Write(keys); err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		row := make([]string, len(keys))
		for i, key := range keys {
			if val, ok := record[key]; ok {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	log.Printf("Exported %s: %s", strings.ToUpper(filepath.Ext(path)[1:]), path)
	return writer.Error()
}

// exportParquet saves records to a Parquet file
func (e *Exporter) exportParquet(path string, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	// Infer schema nodes
	nodeMap := make(map[string]parquet.Node)
	allKeysSet := make(map[string]bool)

	for _, record := range records {
		for k, v := range record {
			allKeysSet[k] = true
			if _, exists := nodeMap[k]; !exists && v != nil {
				switch v.(type) {
				case int, int32, int64:
					nodeMap[k] = parquet.Int(64)
				case float32, float64:
					nodeMap[k] = parquet.Leaf(parquet.DoubleType)
				case bool:
					nodeMap[k] = parquet.Leaf(parquet.BooleanType)
				default:
					nodeMap[k] = parquet.String()
				}
			}
		}
	}

	// Sort keys for deterministic column order
	var keys []string
	for k := range allKeysSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build Parquet Group
	group := make(parquet.Group)
	for _, k := range keys {
		node, ok := nodeMap[k]
		if !ok {
			node = parquet.String()
		}
		group[k] = parquet.Optional(node)
	}

	schema := parquet.NewSchema("InferenceMetrics", group)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	defer f.Close()

	writer := parquet.NewGenericWriter[any](f, schema)

	// Convert maps to Parquet rows
	rows := make([]parquet.Row, len(records))

	for i, record := range records {
		rowValues := make([]parquet.Value, 0, len(keys))

		for j, key := range keys {
			val, exists := record[key]

			if !exists || val == nil {
				rowValues = append(rowValues, parquet.ValueOf(nil).Level(0, 0, j))
				continue
			}

			expectedNode := nodeMap[key]
			if expectedNode == nil {
				expectedNode = parquet.String()
			}

			var pVal parquet.Value

			switch expectedNode.Type().Kind() {
			case parquet.Int64:
				switch v := val.(type) {
				case int:
					pVal = parquet.ValueOf(int64(v))
				case int64:
					pVal = parquet.ValueOf(v)
				case float64:
					pVal = parquet.ValueOf(int64(v))
				default:
					pVal = parquet.ValueOf(int64(0))
				}
			case parquet.Double:
				switch v := val.(type) {
				case float64:
					pVal = parquet.ValueOf(v)
				case int64:
					pVal = parquet.ValueOf(float64(v))
				default:
					pVal = parquet.ValueOf(0.0)
				}
			case parquet.Boolean:
				if v, ok := val.(bool); ok {
					pVal = parquet.ValueOf(v)
				} else {
					pVal = parquet.ValueOf(false)
				}
			default:
				pVal = parquet.ValueOf(fmt.Sprintf("%v", val))
			}

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

	log.Printf("Exported PARQUET: %s", path)
	return nil
}

// flattenMetrics converts DynamicMetrics to a simple map for export
// Since DynamicMetrics is already flat, this just extracts values and timestamps
func flattenMetrics(data collectors.DynamicMetrics) map[string]interface{} {
	flat := make(map[string]interface{})

	for k, v := range data {
		flat[k] = v.Value
		if v.Time > 0 {
			flat["t"+k] = v.Time
		}
	}

	return flat
}

// Cleanup removes intermediate snapshot files
func (e *Exporter) Cleanup() {
	for _, path := range e.snapshotFiles {
		os.Remove(path)
	}
}
