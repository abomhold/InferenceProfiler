package exporting

import (
	"InferenceProfiler/pkg/utils"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

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

// Open opens the file and reads the header row.
func (r *DelimitedReader) Open(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	r.file = file
	r.reader = csv.NewReader(file)
	r.reader.Comma = r.delimiter
	r.reader.FieldsPerRecord = -1
	r.reader.LazyQuotes = true

	header, err := r.reader.Read()
	if err != nil {
		_ = r.file.Close()
		return fmt.Errorf("failed to read header: %w", err)
	}
	r.header = header
	return nil
}

// Read parses all records from the file.
func (r *DelimitedReader) Read() ([]Record, error) {
	var records []Record

	for {
		row, err := r.reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
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

		if f, err := strconv.ParseFloat(val, 64); err == nil {
			if strings.Contains(val, ".") {
				record[key] = f
			} else {
				if i64, err := strconv.ParseInt(val, 10, 64); err == nil {
					record[key] = i64
				} else {
					record[key] = f
				}
			}
		} else if strings.EqualFold(val, "true") {
			record[key] = true
		} else if strings.EqualFold(val, "false") {
			record[key] = false
		} else {
			record[key] = val
		}
	}

	return record
}

// Close closes the underlying file handle.
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

// Init creates the file and prepares the writer.
func (w *DelimitedWriter) Init(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	w.path = path
	w.file = file
	w.writer = csv.NewWriter(file)
	w.writer.Comma = w.delimiter

	return nil
}

// Write writes a single record to the file.
func (w *DelimitedWriter) Write(record Record) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.writeRow(record)
}

// WriteBatch writes multiple records to the file efficiently.
func (w *DelimitedWriter) WriteBatch(records []Record) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, r := range records {
		if err := w.writeRow(r); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}
	return nil
}

// writeRow handles the internal writing logic without locking.
func (w *DelimitedWriter) writeRow(record Record) error {
	if !w.headerSet {
		w.header = w.extractSortedKeys(record)
		if err := w.writer.Write(w.header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		w.headerSet = true
	}

	row := make([]string, len(w.header))
	for i, key := range w.header {
		if val, ok := record[key]; ok {
			row[i] = utils.FormatValue(val)
		}
	}

	if err := w.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}
	return nil
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *DelimitedWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer != nil {
		w.writer.Flush()
		return w.writer.Error()
	}
	return nil
}

// Close flushes the buffer and closes the file.
func (w *DelimitedWriter) Close() error {
	if err := w.Flush(); err != nil {
		if w.file != nil {
			_ = w.file.Close()
		}
		return err
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Path returns the file path.
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
