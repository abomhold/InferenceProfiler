// Package exporter handles metric output in JSON, CSV, and Parquet formats.
package utils

import (
	"InferenceProfiler/collectors"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
)

// Exporter manages metric output files.
type Exporter struct {
	outputDir     string
	sessionUUID   string
	snapshotFiles []string
}

// New creates a new Exporter.
func New(outputDir string, sessionUUID uuid.UUID) *Exporter {
	os.MkdirAll(outputDir, 0755)
	return &Exporter{
		outputDir:   outputDir,
		sessionUUID: sessionUUID,
	}
}

// SaveStatic writes static info to JSON.
func (e *Exporter) SaveStatic(info collectors.StaticInfo) error {
	path := filepath.Join(e.outputDir, fmt.Sprintf("static_%s.json", e.sessionUUID))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(info)
}

// SaveSnapshot writes a single metric snapshot to JSON.
func (e *Exporter) SaveSnapshot(s collectors.Snapshot) error {
	filename := fmt.Sprintf("%s-%d.json", e.sessionUUID, s.Timestamp)
	path := filepath.Join(e.outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(s); err != nil {
		return err
	}

	e.snapshotFiles = append(e.snapshotFiles, path)
	return nil
}

// ProcessSession aggregates all snapshots into final output format.
func (e *Exporter) ProcessSession(format string) error {
	if len(e.snapshotFiles) == 0 {
		return nil
	}

	// Sort files by name (timestamp)
	sort.Strings(e.snapshotFiles)

	// Load all snapshots
	var records []map[string]interface{}
	for _, path := range e.snapshotFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}

		flat := flatten(m, "")
		records = append(records, flat)
	}

	if len(records) == 0 {
		return nil
	}

	basePath := filepath.Join(e.outputDir, e.sessionUUID)

	switch format {
	case "csv":
		return e.writeCSV(basePath+".csv", records, ",")
	case "tsv":
		return e.writeCSV(basePath+".tsv", records, "\t")
	case "parquet":
		// Parquet requires additional dependencies (arrow)
		// Fall back to CSV with a note
		fmt.Println("Note: Parquet output requires Apache Arrow. Using CSV.")
		return e.writeCSV(basePath+".csv", records, ",")
	default:
		return e.writeCSV(basePath+".csv", records, ",")
	}
}

// writeCSV writes records to CSV/TSV.
func (e *Exporter) writeCSV(path string, records []map[string]interface{}, sep string) error {
	if len(records) == 0 {
		return nil
	}

	// Collect all keys
	keySet := make(map[string]bool)
	for _, r := range records {
		for k := range r {
			keySet[k] = true
		}
	}

	// Sort keys for consistent column order
	var keys []string
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if sep == "\t" {
		w.Comma = '\t'
	}

	// Header
	if err := w.Write(keys); err != nil {
		return err
	}

	// Rows
	for _, r := range records {
		row := make([]string, len(keys))
		for i, k := range keys {
			if v, ok := r[k]; ok {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

// flatten converts nested map to flat key-value pairs.
// Nested keys are joined with underscore.
func flatten(m map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "_" + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			for fk, fv := range flatten(val, key) {
				result[fk] = fv
			}
		case []interface{}:
			// Convert arrays to JSON strings
			data, _ := json.Marshal(val)
			result[key] = string(data)
		default:
			result[key] = val
		}
	}

	return result
}
