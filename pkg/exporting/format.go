// Package formatting provides unified read/write interfaces for data formats.
package exporting

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Record is a generic map representing a single metrics record.
type Record = map[string]interface{}

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
	Init(path string) error
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
func SaveRecords(path string, records []Record) error {
	f, ok := GetByPath(path)
	if !ok {
		return fmt.Errorf("unsupported format for file: %s", path)
	}

	writer := f.Writer()
	if err := writer.Init(path); err != nil {
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
