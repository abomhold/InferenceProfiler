package exporting

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/parquet-go/parquet-go"
)

const ParquetBatchSize = 1000

func init() {
	Register(&ParquetFormat{})
}

// ParquetFormat handles Parquet files.
type ParquetFormat struct{}

func (f *ParquetFormat) Name() string         { return "parquet" }
func (f *ParquetFormat) Extensions() []string { return []string{".parquet"} }
func (f *ParquetFormat) Reader() Reader       { return &ParquetReader{} }
func (f *ParquetFormat) Writer() Writer       { return &ParquetWriter{} }

// ParquetReader reads Parquet files.
type ParquetReader struct {
	file  *os.File
	pfile *parquet.File
}

func (r *ParquetReader) Open(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	r.file = file

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat file: %w", err)
	}

	pf, err := parquet.OpenFile(file, stat.Size())
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to open parquet file: %w", err)
	}
	r.pfile = pf

	return nil
}

func (r *ParquetReader) Read() ([]Record, error) {
	if r.pfile == nil {
		return nil, fmt.Errorf("reader not initialized")
	}

	// Get schema field names
	schema := r.pfile.Schema()
	fields := schema.Fields()
	fieldNames := make([]string, len(fields))
	for i, f := range fields {
		fieldNames[i] = f.Name()
	}

	records := make([]Record, 0, r.pfile.NumRows())
	rowBuf := make([]parquet.Row, 100)

	// Iterate over each RowGroup in the file
	for _, rg := range r.pfile.RowGroups() {
		rows := rg.Rows()

		for {
			n, err := rows.ReadRows(rowBuf)
			if n == 0 {
				break
			}

			for i := 0; i < n; i++ {
				record := make(Record, len(fields))
				row := rowBuf[i]

				for j, val := range row {
					if j >= len(fieldNames) {
						continue
					}
					if val.IsNull() {
						continue
					}
					record[fieldNames[j]] = parquetValueToGo(val)
				}
				records = append(records, record)
			}

			if err != nil {
				if err != io.EOF {
					rows.Close()
					return nil, fmt.Errorf("failed to read rows: %w", err)
				}
				break
			}
		}
		rows.Close()
	}

	return records, nil
}

func parquetValueToGo(v parquet.Value) interface{} {
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
		return string(v.ByteArray())
	default:
		return v.String()
	}
}

func (r *ParquetReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// ParquetWriter writes Parquet files using the Row API.
type ParquetWriter struct {
	path       string
	file       *os.File
	writer     *parquet.Writer
	schema     *parquet.Schema
	columns    []string
	schemaInit bool
	buffer     []parquet.Row
	mu         sync.Mutex
}

func (w *ParquetWriter) Init(path string, schema *Schema) error {
	w.path = path
	w.buffer = make([]parquet.Row, 0, ParquetBatchSize)
	return nil
}

func (w *ParquetWriter) initSchema(record Record) error {
	// Build sorted column list
	w.columns = make([]string, 0, len(record))
	for k := range record {
		w.columns = append(w.columns, k)
	}
	sort.Strings(w.columns)

	// Build schema group
	group := make(parquet.Group)
	for _, name := range w.columns {
		val := record[name]
		group[name] = valueToParquetNode(val)
	}

	w.schema = parquet.NewSchema("record", group)

	// Create file and writer
	file, err := os.Create(w.path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	w.file = file

	w.writer = parquet.NewWriter(file, w.schema,
		parquet.Compression(&parquet.Snappy),
	)
	w.schemaInit = true
	return nil
}

func valueToParquetNode(val interface{}) parquet.Node {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return parquet.Optional(parquet.Int(64))
	case float32, float64:
		return parquet.Optional(parquet.Leaf(parquet.DoubleType))
	case bool:
		return parquet.Optional(parquet.Leaf(parquet.BooleanType))
	default:
		return parquet.Optional(parquet.String())
	}
}

func (w *ParquetWriter) recordToRow(record Record) parquet.Row {
	row := make(parquet.Row, len(w.columns))
	for i, name := range w.columns {
		val, ok := record[name]
		if !ok || val == nil {
			row[i] = parquet.NullValue()
			continue
		}
		row[i] = goToParquetValue(val, i)
	}
	return row
}

func goToParquetValue(val interface{}, columnIndex int) parquet.Value {
	switch v := val.(type) {
	case bool:
		return parquet.BooleanValue(v).Level(0, 1, columnIndex)
	case int:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case int8:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case int16:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case int32:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case int64:
		return parquet.Int64Value(v).Level(0, 1, columnIndex)
	case uint:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case uint8:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case uint16:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case uint32:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case uint64:
		return parquet.Int64Value(int64(v)).Level(0, 1, columnIndex)
	case float32:
		return parquet.DoubleValue(float64(v)).Level(0, 1, columnIndex)
	case float64:
		return parquet.DoubleValue(v).Level(0, 1, columnIndex)
	case string:
		return parquet.ByteArrayValue([]byte(v)).Level(0, 1, columnIndex)
	default:
		return parquet.ByteArrayValue([]byte(fmt.Sprintf("%v", v))).Level(0, 1, columnIndex)
	}
}

func (w *ParquetWriter) Write(record Record) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.schemaInit {
		if err := w.initSchema(record); err != nil {
			return err
		}
	}

	row := w.recordToRow(record)
	w.buffer = append(w.buffer, row)

	if len(w.buffer) >= ParquetBatchSize {
		return w.flushBuffer()
	}

	return nil
}

func (w *ParquetWriter) WriteBatch(records []Record) error {
	for i, r := range records {
		if err := w.Write(r); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}
	return nil
}

func (w *ParquetWriter) flushBuffer() error {
	if len(w.buffer) == 0 || w.writer == nil {
		return nil
	}

	for _, row := range w.buffer {
		if _, err := w.writer.WriteRows([]parquet.Row{row}); err != nil {
			return fmt.Errorf("failed to write parquet row: %w", err)
		}
	}

	w.buffer = w.buffer[:0]
	return nil
}

func (w *ParquetWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.flushBuffer(); err != nil {
		return err
	}

	if w.writer != nil {
		return w.writer.Flush()
	}
	return nil
}

func (w *ParquetWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}

	if w.writer != nil {
		if err := w.writer.Close(); err != nil {
			return err
		}
	}

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *ParquetWriter) Path() string {
	return w.path
}
