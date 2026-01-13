package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
)

// FormatWriter defines the interface for writing records in different formats.
// Implementations handle format-specific initialization, record writing, and cleanup.
type FormatWriter interface {
	// Init initializes the writer with the output path and schema (column names).
	// For formats that require upfront schema definition (CSV, Parquet), the schema
	// determines column order. For schema-less formats (JSONL), schema may be ignored.
	Init(path string, schema []string) error

	// Write writes a single record to the output.
	// The record is a map of column names to values.
	Write(record map[string]interface{}) error

	// Close finalizes the output and releases resources.
	// Must be called to ensure all data is flushed.
	Close() error

	// Extension returns the file extension for this format (e.g., ".jsonl", ".parquet")
	Extension() string
}

// NewFormatWriter creates a FormatWriter for the specified format.
// Supported formats: "jsonl", "parquet", "csv", "tsv"
// Returns an error for unknown formats.
func NewFormatWriter(format string) (FormatWriter, error) {
	switch strings.ToLower(format) {
	case "jsonl":
		return &JSONLWriter{}, nil
	case "parquet":
		return &ParquetWriter{}, nil
	case "csv":
		return &CSVWriter{delimiter: ','}, nil
	case "tsv":
		return &CSVWriter{delimiter: '\t'}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s (supported: jsonl, parquet, csv, tsv)", format)
	}
}

// =============================================================================
// JSONL Writer - JSON Lines format (one JSON object per line)
// =============================================================================

// JSONLWriter writes records as newline-delimited JSON (JSON Lines format).
// Each record is written as a single-line JSON object.
type JSONLWriter struct {
	file   *os.File
	writer *bufio.Writer
	schema []string // stored for consistent key ordering
}

func (w *JSONLWriter) Init(path string, schema []string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create JSONL file: %w", err)
	}
	w.file = f
	w.writer = bufio.NewWriter(f)
	w.schema = schema
	return nil
}

func (w *JSONLWriter) Write(record map[string]interface{}) error {
	// Use schema order if available, otherwise sort keys for deterministic output
	var orderedRecord map[string]interface{}
	if len(w.schema) > 0 {
		orderedRecord = make(map[string]interface{}, len(w.schema))
		for _, key := range w.schema {
			if val, ok := record[key]; ok {
				orderedRecord[key] = val
			}
		}
	} else {
		orderedRecord = record
	}

	data, err := json.Marshal(orderedRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	if _, err := w.writer.Write(data); err != nil {
		return err
	}
	return w.writer.WriteByte('\n')
}

func (w *JSONLWriter) Close() error {
	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return err
		}
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *JSONLWriter) Extension() string {
	return ".jsonl"
}

// =============================================================================
// CSV/TSV Writer - Delimited text format
// =============================================================================

// CSVWriter writes records as CSV or TSV (delimiter-configurable).
// First row contains column headers; subsequent rows contain values.
type CSVWriter struct {
	file      *os.File
	writer    *csv.Writer
	schema    []string
	delimiter rune
	headerWritten bool
}

func (w *CSVWriter) Init(path string, schema []string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	w.file = f
	w.writer = csv.NewWriter(f)
	w.writer.Comma = w.delimiter
	w.schema = schema
	w.headerWritten = false
	return nil
}

func (w *CSVWriter) Write(record map[string]interface{}) error {
	// Write header on first record
	if !w.headerWritten {
		if err := w.writer.Write(w.schema); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
		w.headerWritten = true
	}

	// Build row in schema order
	row := make([]string, len(w.schema))
	for i, key := range w.schema {
		if val, ok := record[key]; ok {
			row[i] = formatCSVValue(val)
		} else {
			row[i] = ""
		}
	}

	return w.writer.Write(row)
}

func (w *CSVWriter) Close() error {
	if w.writer != nil {
		w.writer.Flush()
		if err := w.writer.Error(); err != nil {
			return err
		}
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *CSVWriter) Extension() string {
	if w.delimiter == '\t' {
		return ".tsv"
	}
	return ".csv"
}

// formatCSVValue converts a value to its CSV string representation
func formatCSVValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case float64:
		// Handle special float values
		if val != val { // NaN
			return ""
		}
		return fmt.Sprintf("%v", val)
	case float32:
		if val != val { // NaN
			return ""
		}
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// =============================================================================
// Parquet Writer - Columnar format for analytics
// =============================================================================

// ParquetWriter writes records in Apache Parquet format.
// Uses a dynamic schema based on the first record's structure.
type ParquetWriter struct {
	path       string
	schema     []string
	pw         *writer.JSONWriter
	localFile  *local.LocalFile
	records    []map[string]interface{} // buffer for batch writing
	batchSize  int
	schemaJSON string
}

func (w *ParquetWriter) Init(path string, schema []string) error {
	w.path = path
	w.schema = schema
	w.batchSize = 1000 // Write in batches for efficiency
	w.records = make([]map[string]interface{}, 0, w.batchSize)
	return nil
}

func (w *ParquetWriter) Write(record map[string]interface{}) error {
	// Initialize writer on first record (we need data to infer types)
	if w.pw == nil {
		if err := w.initParquetWriter(record); err != nil {
			return err
		}
	}

	// Add to buffer
	w.records = append(w.records, record)

	// Flush if batch is full
	if len(w.records) >= w.batchSize {
		return w.flushBatch()
	}
	return nil
}

func (w *ParquetWriter) Close() error {
	// Flush remaining records
	if len(w.records) > 0 {
		if err := w.flushBatch(); err != nil {
			return err
		}
	}

	if w.pw != nil {
		if err := w.pw.WriteStop(); err != nil {
			return fmt.Errorf("failed to finalize parquet: %w", err)
		}
	}

	if w.localFile != nil {
		return w.localFile.Close()
	}
	return nil
}

func (w *ParquetWriter) Extension() string {
	return ".parquet"
}

// initParquetWriter creates the parquet writer with schema inferred from the first record
func (w *ParquetWriter) initParquetWriter(sample map[string]interface{}) error {
	// Build parquet schema from schema order and sample values
	schema := w.buildParquetSchema(sample)
	w.schemaJSON = schema

	// Create local file
	var err error
	w.localFile, err = local.NewLocalFileWriter(w.path)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}

	// Create JSON writer (easier than native parquet for dynamic schemas)
	w.pw, err = writer.NewJSONWriter(schema, w.localFile, 4) // 4 = compression codec (snappy)
	if err != nil {
		return fmt.Errorf("failed to create parquet writer: %w", err)
	}

	return nil
}

// buildParquetSchema generates a parquet schema JSON from the schema and sample record
func (w *ParquetWriter) buildParquetSchema(sample map[string]interface{}) string {
	var fields []string

	for _, key := range w.schema {
		val := sample[key]
		parquetType := inferParquetType(val)
		
		// Parquet field definition
		field := fmt.Sprintf(`{"Tag": "name=%s, type=%s, repetitiontype=OPTIONAL"}`, key, parquetType)
		fields = append(fields, field)
	}

	return fmt.Sprintf(`{"Tag": "name=metrics, repetitiontype=REQUIRED", "Fields": [%s]}`, 
		strings.Join(fields, ", "))
}

// inferParquetType maps Go types to Parquet types
func inferParquetType(v interface{}) string {
	switch v.(type) {
	case int, int32, int64:
		return "INT64"
	case float32, float64:
		return "DOUBLE"
	case bool:
		return "BOOLEAN"
	case string:
		return "BYTE_ARRAY, convertedtype=UTF8"
	default:
		// Default to string for unknown types
		return "BYTE_ARRAY, convertedtype=UTF8"
	}
}

// flushBatch writes buffered records to the parquet file
func (w *ParquetWriter) flushBatch() error {
	for _, record := range w.records {
		// Convert record to JSON for the JSON writer
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record for parquet: %w", err)
		}
		
		if err := w.pw.Write(string(data)); err != nil {
			return fmt.Errorf("failed to write parquet record: %w", err)
		}
	}
	
	// Clear buffer
	w.records = w.records[:0]
	return nil
}

// =============================================================================
// Record Loader Interface - For reading existing data files
// =============================================================================

// RecordLoader defines the interface for reading records from files.
// Used for graph generation from existing data files.
type RecordLoader interface {
	// Load reads all records from the specified path.
	// Returns records as a slice of maps, where each map represents one row.
	Load(path string) ([]map[string]interface{}, error)
}

// NewRecordLoader creates a RecordLoader based on file extension.
// Supported extensions: .jsonl, .parquet, .csv, .tsv
func NewRecordLoader(path string) (RecordLoader, error) {
	ext := strings.ToLower(getFileExtension(path))
	switch ext {
	case ".jsonl":
		return &JSONLLoader{}, nil
	case ".parquet":
		return &ParquetLoader{}, nil
	case ".csv":
		return &CSVLoader{delimiter: ','}, nil
	case ".tsv":
		return &CSVLoader{delimiter: '\t'}, nil
	default:
		return nil, fmt.Errorf("unknown file format: %s", ext)
	}
}

// getFileExtension returns the file extension including the dot
func getFileExtension(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ""
}

// =============================================================================
// JSONL Loader
// =============================================================================

type JSONLLoader struct{}

func (l *JSONLLoader) Load(path string) ([]map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSONL file: %w", err)
	}
	defer f.Close()

	var records []map[string]interface{}
	scanner := bufio.NewScanner(f)
	
	// Increase buffer size for large lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return records, nil
}

// =============================================================================
// CSV/TSV Loader
// =============================================================================

type CSVLoader struct {
	delimiter rune
}

func (l *CSVLoader) Load(path string) ([]map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = l.delimiter
	reader.LazyQuotes = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Read all rows
	var records []map[string]interface{}
	for {
		row, err := reader.Read()
		if err != nil {
			break // EOF or error
		}

		record := make(map[string]interface{}, len(header))
		for i, key := range header {
			if i < len(row) {
				record[key] = parseCSVValue(row[i])
			}
		}
		records = append(records, record)
	}

	return records, nil
}

// parseCSVValue attempts to parse a CSV value as a number, falling back to string
func parseCSVValue(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Try integer first
	var intVal int64
	if _, err := fmt.Sscanf(s, "%d", &intVal); err == nil {
		// Check it's actually an integer (no decimal point)
		if !strings.Contains(s, ".") {
			return intVal
		}
	}

	// Try float
	var floatVal float64
	if _, err := fmt.Sscanf(s, "%f", &floatVal); err == nil {
		return floatVal
	}

	// Return as string
	return s
}

// =============================================================================
// Parquet Loader
// =============================================================================

type ParquetLoader struct{}

func (l *ParquetLoader) Load(path string) ([]map[string]interface{}, error) {
	fr, err := local.NewLocalFileReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}
	defer fr.Close()

	// Use JSON reader for flexibility with unknown schemas
	pr, err := reader.NewParquetReader(fr, nil, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer pr.ReadStop()

	numRows := int(pr.GetNumRows())
	records := make([]map[string]interface{}, 0, numRows)

	// Read in batches
	batchSize := 1000
	for {
		rows, err := pr.ReadByNumber(batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read parquet rows: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			// Row comes as interface{}, convert to map
			if rowMap, ok := row.(map[string]interface{}); ok {
				records = append(records, rowMap)
			}
		}
	}

	return records, nil
}

// =============================================================================
// Helper: Extract Schema from Records
// =============================================================================

// ExtractSchema extracts a sorted list of all unique keys from a set of records.
// This is useful for building CSV headers or parquet schemas from data.
func ExtractSchema(records []map[string]interface{}) []string {
	keySet := make(map[string]struct{})
	for _, record := range records {
		for key := range record {
			keySet[key] = struct{}{}
		}
	}

	schema := make([]string, 0, len(keySet))
	for key := range keySet {
		schema = append(schema, key)
	}
	sort.Strings(schema)
	return schema
}
