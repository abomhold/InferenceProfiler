// Package formatting provides unified read/write interfaces for data formats.
package formatting

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Record is a generic map representing a single metrics record.
type Record = map[string]interface{}

// Schema defines the structure of records.
type Schema struct {
	Columns []Column
}

// Column describes a single field in the schema.
type Column struct {
	Name string
	Type ColumnType
}

// ColumnType indicates the data type of a column.
type ColumnType int

const (
	TypeString ColumnType = iota
	TypeInt64
	TypeFloat64
	TypeBool
)

// Format defines the interface for a data format.
type Format interface {
	Name() string
	Extensions() []string
	Reader() Reader
	Writer() Writer
}

// Reader reads records from a file.
type Reader interface {
	Open(path string) error
	Read() ([]Record, error)
	Close() error
}

// Writer writes records to a file.
type Writer interface {
	Init(path string, schema *Schema) error
	Write(record Record) error
	WriteBatch(records []Record) error
	Flush() error
	Close() error
	Path() string
}

// Registry management
var (
	registry    = make(map[string]Format)
	extRegistry = make(map[string]Format)
)

// Register adds a format to the registry.
func Register(f Format) {
	name := strings.ToLower(f.Name())
	registry[name] = f
	for _, ext := range f.Extensions() {
		extRegistry[strings.ToLower(ext)] = f
	}
}

// Get returns a format by name.
func Get(name string) (Format, bool) {
	f, ok := registry[strings.ToLower(name)]
	return f, ok
}

// GetByExtension returns a format by file extension.
func GetByExtension(ext string) (Format, bool) {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	f, ok := extRegistry[ext]
	return f, ok
}

// GetByPath returns a format based on the file's extension.
func GetByPath(path string) (Format, bool) {
	return GetByExtension(filepath.Ext(path))
}

// List returns all registered format names.
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// GetExtension returns the file extension for a format name.
func GetExtension(format string) string {
	switch strings.ToLower(format) {
	case "jsonl", "json":
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

// LoadRecords loads all records from a file.
func LoadRecords(path string) ([]Record, error) {
	f, ok := GetByPath(path)
	if !ok {
		return nil, fmt.Errorf("unsupported format for file: %s", path)
	}

	reader := f.Reader()
	if err := reader.Open(path); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()

	records, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read records: %w", err)
	}

	return records, nil
}

// SaveRecords writes records to a file.
func SaveRecords(path string, records []Record, schema *Schema) error {
	f, ok := GetByPath(path)
	if !ok {
		return fmt.Errorf("unsupported format for file: %s", path)
	}

	writer := f.Writer()
	if err := writer.Init(path, schema); err != nil {
		return fmt.Errorf("failed to initialize writer: %w", err)
	}

	if err := writer.WriteBatch(records); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write records: %w", err)
	}

	if err := writer.Flush(); err != nil {
		writer.Close()
		return fmt.Errorf("failed to flush: %w", err)
	}

	return writer.Close()
}

// ExtractSchema infers schema from records.
func ExtractSchema(records []Record) *Schema {
	if len(records) == 0 {
		return &Schema{}
	}

	columns := make(map[string]ColumnType)
	for _, r := range records {
		for k, v := range r {
			if _, exists := columns[k]; !exists {
				columns[k] = inferType(v)
			}
		}
	}

	schema := &Schema{Columns: make([]Column, 0, len(columns))}
	for name, typ := range columns {
		schema.Columns = append(schema.Columns, Column{Name: name, Type: typ})
	}
	return schema
}

func inferType(v interface{}) ColumnType {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return TypeInt64
	case float32, float64:
		return TypeFloat64
	case bool:
		return TypeBool
	default:
		return TypeString
	}
}

// SplitLines splits a string into lines.
func SplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
