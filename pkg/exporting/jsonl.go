package exporting

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const (
	DefaultBufferSize = 64 * 1024
	MaxLineSize       = 10 * 1024 * 1024
)

func init() {
	Register(&JSONLFormat{})
}

// JSONLFormat handles JSON Lines format.
type JSONLFormat struct{}

func (f *JSONLFormat) Name() string         { return "jsonl" }
func (f *JSONLFormat) Extensions() []string { return []string{".jsonl"} }
func (f *JSONLFormat) Reader() Reader       { return &JSONLReader{} }
func (f *JSONLFormat) Writer() Writer       { return &JSONLWriter{} }

// JSONLReader reads JSONL files.
type JSONLReader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func (r *JSONLReader) Open(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	r.file = file
	r.scanner = bufio.NewScanner(file)
	r.scanner.Buffer(make([]byte, DefaultBufferSize), MaxLineSize)
	return nil
}

func (r *JSONLReader) Read() ([]Record, error) {
	var records []Record
	lineNum := 0
	for r.scanner.Scan() {
		lineNum++
		line := r.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var record Record
		if err := json.Unmarshal(line, &record); err != nil {
			// Skip malformed lines but continue reading
			continue
		}
		records = append(records, record)
	}

	if err := r.scanner.Err(); err != nil {
		return records, fmt.Errorf("scanner error: %w", err)
	}

	return records, nil
}

func (r *JSONLReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// JSONLWriter writes JSONL files.
type JSONLWriter struct {
	path   string
	file   *os.File
	writer *bufio.Writer
	mu     sync.Mutex
}

func (w *JSONLWriter) Init(path string, _ *Schema) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	w.path = path
	w.file = file
	w.writer = bufio.NewWriterSize(file, DefaultBufferSize)
	return nil
}

func (w *JSONLWriter) Write(record Record) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	if err := w.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

func (w *JSONLWriter) WriteBatch(records []Record) error {
	for i, r := range records {
		if err := w.Write(r); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}
	return nil
}

func (w *JSONLWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer != nil {
		return w.writer.Flush()
	}
	return nil
}

func (w *JSONLWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *JSONLWriter) Path() string {
	return w.path
}
