package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/parquet-go/parquet-go"
)

// FormatWriter defines the interface for writing records in different formats.
type FormatWriter interface {
	Init(path string, schema []string) error
	Write(record map[string]interface{}) error
	Close() error
	Extension() string
}

// NewFormatWriter creates a FormatWriter for the specified format.
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
// JSONL Writer
// =============================================================================

type JSONLWriter struct {
	file   *os.File
	writer *bufio.Writer
	schema []string
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
// CSV/TSV Writer
// =============================================================================

type CSVWriter struct {
	file          *os.File
	writer        *csv.Writer
	schema        []string
	delimiter     rune
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
	if !w.headerWritten {
		if err := w.writer.Write(w.schema); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
		w.headerWritten = true
	}

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

func formatCSVValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case float64:
		if val != val { // NaN
			return ""
		}
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// =============================================================================
// Parquet Writer (Segmentio Implementation)
// =============================================================================

type ParquetWriter struct {
	file        *os.File
	writer      *parquet.GenericWriter[any]
	schemaKeys  []string
	initialized bool
}

func (w *ParquetWriter) Init(path string, schema []string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	w.file = f
	w.schemaKeys = schema
	w.initialized = false
	return nil
}

func (w *ParquetWriter) Write(record map[string]interface{}) error {
	if !w.initialized {
		if err := w.initWriter(record); err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]any{record})
	return err
}

func (w *ParquetWriter) Close() error {
	if w.writer != nil {
		if err := w.writer.Close(); err != nil {
			return fmt.Errorf("failed to close parquet writer: %w", err)
		}
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *ParquetWriter) Extension() string {
	return ".parquet"
}

func (w *ParquetWriter) initWriter(sample map[string]interface{}) error {
	fields := make(map[string]parquet.Node)
	for _, name := range w.schemaKeys {
		val := sample[name]
		fields[name] = parquet.Optional(inferParquetNode(val))
	}
	schema := parquet.NewSchema("metrics", parquet.Group(fields))
	w.writer = parquet.NewGenericWriter[any](w.file, schema)
	w.initialized = true
	return nil
}

// inferParquetNode maps Go values to Parquet Nodes
func inferParquetNode(v interface{}) parquet.Node {

	var t parquet.Type
	switch v.(type) {
	case int, int32, int64:
		t = parquet.Int64Type
	case float32, float64:
		t = parquet.DoubleType
	case bool:
		t = parquet.BooleanType
	default:
		return parquet.String()
	}
	return parquet.Leaf(t)
}

// =============================================================================
// Record Loader Interface
// =============================================================================

type RecordLoader interface {
	Load(path string) ([]map[string]interface{}, error)
}

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

	return records, scanner.Err()
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

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var records []map[string]interface{}
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
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

func parseCSVValue(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var intVal int64
	if _, err := fmt.Sscanf(s, "%d", &intVal); err == nil {
		if !strings.Contains(s, ".") {
			return intVal
		}
	}
	var floatVal float64
	if _, err := fmt.Sscanf(s, "%f", &floatVal); err == nil {
		return floatVal
	}
	return s
}

// =============================================================================
// Parquet Loader
// =============================================================================

type ParquetLoader struct{}

func (l *ParquetLoader) Load(path string) ([]map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	pf, err := parquet.OpenFile(f, info.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}

	// Dynamic row reading: inspect schema to find column names
	schema := pf.Schema()
	columnPaths := schema.Columns()

	// Convert column paths to dot-delimited strings for map keys
	colNames := make([]string, len(columnPaths))
	for i, colPath := range columnPaths {
		colNames[i] = strings.Join(colPath, ".")
	}

	var records []map[string]interface{}

	// Iterate through row groups
	for _, rowGroup := range pf.RowGroups() {
		rows := rowGroup.Rows()
		buffer := make([]parquet.Row, 100)

		for {
			n, err := rows.ReadRows(buffer)
			if n == 0 && err != nil {
				if err == io.EOF {
					break
				}
				rows.Close()
				return nil, err
			}

			for i := 0; i < n; i++ {
				row := buffer[i]
				record := make(map[string]interface{})
				for _, v := range row {
					colIdx := v.Column()
					if colIdx >= 0 && colIdx < len(colNames) {
						record[colNames[colIdx]] = convertParquetValue(v)
					}
				}
				records = append(records, record)
			}

			if err == io.EOF {
				break
			}
		}
		rows.Close()
	}

	return records, nil
}

func convertParquetValue(v parquet.Value) interface{} {
	switch v.Kind() {
	case parquet.Boolean:
		return v.Boolean()
	case parquet.Int32:
		return int64(v.Int32())
	case parquet.Int64:
		return v.Int64()
	case parquet.Float:
		return float64(v.Float())
	case parquet.Double:
		return v.Double()
	case parquet.ByteArray, parquet.FixedLenByteArray:
		return v.String()
	default:
		return v.String()
	}
}

// =============================================================================
// Helper: Extract Schema from Records
// =============================================================================

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
