package formatting

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
)

func init() {
	Register(&CSVFormat{})
	Register(&TSVFormat{})
}

// CSVFormat handles CSV files.
type CSVFormat struct{}

func (f *CSVFormat) Name() string         { return "csv" }
func (f *CSVFormat) Extensions() []string { return []string{".csv"} }
func (f *CSVFormat) Reader() Reader       { return &DelimitedReader{delimiter: ','} }
func (f *CSVFormat) Writer() Writer       { return &DelimitedWriter{delimiter: ','} }

// TSVFormat handles TSV files.
type TSVFormat struct{}

func (f *TSVFormat) Name() string         { return "tsv" }
func (f *TSVFormat) Extensions() []string { return []string{".tsv"} }
func (f *TSVFormat) Reader() Reader       { return &DelimitedReader{delimiter: '\t'} }
func (f *TSVFormat) Writer() Writer       { return &DelimitedWriter{delimiter: '\t'} }

// DelimitedReader reads CSV/TSV files.
type DelimitedReader struct {
	file      *os.File
	reader    *csv.Reader
	header    []string
	delimiter rune
}

func (r *DelimitedReader) Open(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	r.file = file
	r.reader = csv.NewReader(file)
	r.reader.Comma = r.delimiter
	r.reader.FieldsPerRecord = -1 // Allow variable field counts
	r.reader.LazyQuotes = true

	// Read header row
	header, err := r.reader.Read()
	if err != nil {
		r.file.Close()
		return fmt.Errorf("failed to read header: %w", err)
	}
	r.header = header
	return nil
}

func (r *DelimitedReader) Read() ([]Record, error) {
	var records []Record

	for {
		row, err := r.reader.Read()
		if err != nil {
			break // EOF or error
		}
		records = append(records, r.rowToRecord(row))
	}

	return records, nil
}

func (r *DelimitedReader) rowToRecord(row []string) Record {
	record := make(Record)

	for i, val := range row {
		if i >= len(r.header) || val == "" {
			continue
		}
		key := r.header[i]

		// Try to parse as number or bool
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			// Check if it's actually an integer
			if i64, err := strconv.ParseInt(val, 10, 64); err == nil {
				record[key] = i64
			} else {
				record[key] = f
			}
		} else if val == "true" || val == "True" || val == "TRUE" {
			record[key] = true
		} else if val == "false" || val == "False" || val == "FALSE" {
			record[key] = false
		} else {
			record[key] = val
		}
	}

	return record
}

func (r *DelimitedReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// DelimitedWriter writes CSV/TSV files.
type DelimitedWriter struct {
	path      string
	file      *os.File
	writer    *csv.Writer
	header    []string
	headerSet bool
	delimiter rune
	mu        sync.Mutex
}

func (w *DelimitedWriter) Init(path string, schema *Schema) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	w.path = path
	w.file = file
	w.writer = csv.NewWriter(file)
	w.writer.Comma = w.delimiter

	// If schema provided, write header immediately
	if schema != nil && len(schema.Columns) > 0 {
		w.header = make([]string, len(schema.Columns))
		for i, col := range schema.Columns {
			w.header[i] = col.Name
		}
		if err := w.writer.Write(w.header); err != nil {
			w.file.Close()
			return fmt.Errorf("failed to write header: %w", err)
		}
		w.headerSet = true
	}

	return nil
}

func (w *DelimitedWriter) Write(record Record) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Initialize header from first record if not set
	if !w.headerSet {
		w.header = w.extractSortedKeys(record)
		if err := w.writer.Write(w.header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		w.headerSet = true
	}

	// Build row in header order
	row := make([]string, len(w.header))
	for i, key := range w.header {
		if val, ok := record[key]; ok {
			row[i] = FormatValue(val)
		}
	}

	if err := w.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	return nil
}

func (w *DelimitedWriter) WriteBatch(records []Record) error {
	for i, r := range records {
		if err := w.Write(r); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}
	return nil
}

func (w *DelimitedWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer != nil {
		w.writer.Flush()
		return w.writer.Error()
	}
	return nil
}

func (w *DelimitedWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *DelimitedWriter) Path() string {
	return w.path
}

func (w *DelimitedWriter) extractSortedKeys(record Record) []string {
	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
