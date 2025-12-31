package output

import (
	"InferenceProfiler/src/collectors"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/parquet-go/parquet-go"
)

// TestMain acts as the entry point for tests in this package.
// It detects if benchmarks are running and silences the logs to ensure clean output.
func TestMain(m *testing.M) {
	// Check if the -test.bench flag is present in arguments
	isBench := false
	for _, arg := range os.Args {
		if strings.Contains(arg, "-test.bench") {
			isBench = true
			break
		}
	}

	// If benchmarking, discard all log output to prevent console spam
	if isBench {
		log.SetOutput(io.Discard)
	}

	os.Exit(m.Run())
}

// ============================================================================
// Streaming Tests - All Formats
// ============================================================================

func TestStreamingJSONL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler-stream-jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, err := NewExporter(tmpDir, sessionUUID, true, true, "jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer exp.CloseStream()

	// Write snapshots
	metrics := []*collectors.DynamicMetrics{
		{Timestamp: 1000, CPUTimeUserMode: 100},
		{Timestamp: 2000, CPUTimeUserMode: 200},
		{Timestamp: 3000, CPUTimeUserMode: 300},
	}

	for _, m := range metrics {
		if err := exp.SaveSnapshot(m); err != nil {
			t.Fatal(err)
		}
	}

	if err := exp.CloseStream(); err != nil {
		t.Fatal(err)
	}

	// Verify output
	path := filepath.Join(tmpDir, sessionUUID.String()+".jsonl")
	lines := readJSONLFile(t, path)

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0]["timestamp"] != float64(1000) {
		t.Errorf("First record timestamp = %v; want 1000", lines[0]["timestamp"])
	}
	if lines[2]["vCpuTimeUserMode"] != float64(300) {
		t.Errorf("Third record vCpuTimeUserMode = %v; want 300", lines[2]["vCpuTimeUserMode"])
	}
}

func TestStreamingParquet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler-stream-parquet")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, err := NewExporter(tmpDir, sessionUUID, true, true, "parquet")
	if err != nil {
		t.Fatal(err)
	}
	defer exp.CloseStream()

	// Write snapshots with GPUs
	metrics := []*collectors.DynamicMetrics{
		{
			Timestamp:       1000,
			CPUTimeUserMode: 100,
			NvidiaGPUs: []collectors.NvidiaGPUDynamic{
				{Index: 0, UtilizationGPU: 50},
			},
		},
		{Timestamp: 2000, CPUTimeUserMode: 200},
		{
			Timestamp:       3000,
			CPUTimeUserMode: 300,
			NvidiaGPUs: []collectors.NvidiaGPUDynamic{
				{Index: 0, UtilizationGPU: 85},
			},
		},
	}

	for _, m := range metrics {
		if err := exp.SaveSnapshot(m); err != nil {
			t.Fatal(err)
		}
	}

	if err := exp.CloseStream(); err != nil {
		t.Fatal(err)
	}

	// Verify output
	path := filepath.Join(tmpDir, sessionUUID.String()+".parquet")
	rows := readParquetFile(t, path)

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	if rows[2]["nvidia0UtilizationGpu"] != int64(85) {
		t.Errorf("Third row nvidia0UtilizationGpu = %v; want 85", rows[2]["nvidia0UtilizationGpu"])
	}
}

func TestStreamingCSV(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler-stream-csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, err := NewExporter(tmpDir, sessionUUID, true, true, "csv")
	if err != nil {
		t.Fatal(err)
	}
	defer exp.CloseStream()

	// Write snapshots
	metrics := []*collectors.DynamicMetrics{
		{Timestamp: 1000, CPUTimeUserMode: 100, LoadAvg: 1.5},
		{Timestamp: 2000, CPUTimeUserMode: 200, LoadAvg: 2.5},
	}

	for _, m := range metrics {
		if err := exp.SaveSnapshot(m); err != nil {
			t.Fatal(err)
		}
	}

	if err := exp.CloseStream(); err != nil {
		t.Fatal(err)
	}

	// Verify output
	path := filepath.Join(tmpDir, sessionUUID.String()+".csv")
	records := readCSVFile(t, path, ',')

	if len(records) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 rows (header + 2 data), got %d", len(records))
	}

	// Check header exists
	header := records[0]
	hasTimestamp := false
	for _, col := range header {
		if col == "timestamp" {
			hasTimestamp = true
			break
		}
	}
	if !hasTimestamp {
		t.Error("Header missing 'timestamp' column")
	}
}

func TestStreamingTSV(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler-stream-tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, err := NewExporter(tmpDir, sessionUUID, true, true, "tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer exp.CloseStream()

	// Write snapshots
	metrics := []*collectors.DynamicMetrics{
		{Timestamp: 1000, CPUTimeUserMode: 100},
		{Timestamp: 2000, CPUTimeUserMode: 200},
	}

	for _, m := range metrics {
		if err := exp.SaveSnapshot(m); err != nil {
			t.Fatal(err)
		}
	}

	if err := exp.CloseStream(); err != nil {
		t.Fatal(err)
	}

	// Verify output
	path := filepath.Join(tmpDir, sessionUUID.String()+".tsv")
	records := readCSVFile(t, path, '\t')

	if len(records) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 rows, got %d", len(records))
	}
}

// ============================================================================
// Batch vs Stream Comparison Tests
// ============================================================================

func TestBatchVsStreamJSONL(t *testing.T) {
	streamDir, _ := os.MkdirTemp("", "stream-jsonl")
	defer os.RemoveAll(streamDir)
	batchDir, _ := os.MkdirTemp("", "batch-jsonl")
	defer os.RemoveAll(batchDir)

	sessionUUID := uuid.New()
	metrics := []*collectors.DynamicMetrics{
		{Timestamp: 1000, CPUTimeUserMode: 100, MemoryTotal: 8192},
		{Timestamp: 2000, CPUTimeUserMode: 200, MemoryTotal: 8192},
	}

	// Stream mode
	streamExp, _ := NewExporter(streamDir, sessionUUID, true, true, "jsonl")
	for _, m := range metrics {
		streamExp.SaveSnapshot(m)
	}
	streamExp.CloseStream()

	// Batch mode
	batchExp, _ := NewExporter(batchDir, sessionUUID, true, false, "jsonl")
	for _, m := range metrics {
		batchExp.SaveSnapshot(m)
	}
	batchExp.ProcessSession()

	// Compare outputs
	streamPath := filepath.Join(streamDir, sessionUUID.String()+".jsonl")
	batchPath := filepath.Join(batchDir, sessionUUID.String()+".jsonl")

	streamLines := readJSONLFile(t, streamPath)
	batchLines := readJSONLFile(t, batchPath)

	if len(streamLines) != len(batchLines) {
		t.Errorf("Line count mismatch: stream=%d, batch=%d", len(streamLines), len(batchLines))
	}

	// Compare timestamps
	for i := range streamLines {
		if streamLines[i]["timestamp"] != batchLines[i]["timestamp"] {
			t.Errorf("Row %d timestamp mismatch", i)
		}
	}
}

func TestBatchVsStreamParquet(t *testing.T) {
	streamDir, _ := os.MkdirTemp("", "stream-pq")
	defer os.RemoveAll(streamDir)
	batchDir, _ := os.MkdirTemp("", "batch-pq")
	defer os.RemoveAll(batchDir)

	sessionUUID := uuid.New()
	metrics := []*collectors.DynamicMetrics{
		{Timestamp: 1000, CPUTimeUserMode: 100},
		{Timestamp: 2000, CPUTimeUserMode: 200},
	}

	// Stream mode
	streamExp, _ := NewExporter(streamDir, sessionUUID, true, true, "parquet")
	for _, m := range metrics {
		streamExp.SaveSnapshot(m)
	}
	streamExp.CloseStream()

	// Batch mode
	batchExp, _ := NewExporter(batchDir, sessionUUID, true, false, "parquet")
	for _, m := range metrics {
		batchExp.SaveSnapshot(m)
	}
	batchExp.ProcessSession()

	// Compare
	streamPath := filepath.Join(streamDir, sessionUUID.String()+".parquet")
	batchPath := filepath.Join(batchDir, sessionUUID.String()+".parquet")

	streamRows := readParquetFile(t, streamPath)
	batchRows := readParquetFile(t, batchPath)

	if len(streamRows) != len(batchRows) {
		t.Errorf("Row count mismatch: stream=%d, batch=%d", len(streamRows), len(batchRows))
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestStreamingNoSnapshots(t *testing.T) {
	for _, format := range []string{"jsonl", "parquet", "csv", "tsv"} {
		t.Run(format, func(t *testing.T) {
			tmpDir, _ := os.MkdirTemp("", "stream-empty-"+format)
			defer os.RemoveAll(tmpDir)

			sessionUUID := uuid.New()
			exp, _ := NewExporter(tmpDir, sessionUUID, true, true, format)

			// Close without writing
			if err := exp.CloseStream(); err != nil {
				t.Fatal(err)
			}

			// No file should exist
			ext := format
			path := filepath.Join(tmpDir, sessionUUID.String()+"."+ext)
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("File should not exist when no snapshots written: %s", format)
			}
		})
	}
}

func TestStreamingMultipleClose(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "stream-close")
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, _ := NewExporter(tmpDir, sessionUUID, true, true, "jsonl")

	// Write one snapshot
	exp.SaveSnapshot(&collectors.DynamicMetrics{Timestamp: 1000})

	// Multiple closes should be safe
	for i := 0; i < 3; i++ {
		if err := exp.CloseStream(); err != nil {
			t.Fatalf("Close #%d failed: %v", i+1, err)
		}
	}
}

func TestInvalidFormatForStreaming(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "stream-invalid")
	defer os.RemoveAll(tmpDir)

	_, err := NewExporter(tmpDir, uuid.New(), true, true, "invalid")
	if err == nil {
		t.Error("Expected error for invalid streaming format")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func readJSONLFile(t *testing.T, path string) []map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", path, err)
	}
	defer file.Close()

	var lines []map[string]interface{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("Failed to parse JSON line: %v", err)
		}
		lines = append(lines, record)
	}

	return lines
}

func readParquetFile(t *testing.T, path string) []map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	pf, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		t.Fatalf("Failed to open parquet file: %v", err)
	}

	var rows []map[string]interface{}
	reader := parquet.NewReader(pf)

	for {
		row := make(map[string]interface{})
		err := reader.Read(&row)
		if err != nil {
			break
		}
		rows = append(rows, row)
	}

	return rows
}

func readCSVFile(t *testing.T, path string, delimiter rune) [][]string {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = delimiter

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	return records
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkStreamingJSONL(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "bench-jsonl")
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, _ := NewExporter(tmpDir, sessionUUID, true, true, "jsonl")
	defer exp.CloseStream()

	metrics := &collectors.DynamicMetrics{
		Timestamp:       1000,
		CPUTimeUserMode: 100,
	}

	// First write to initialize
	exp.SaveSnapshot(metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Timestamp = int64(1000 + i)
		exp.SaveSnapshot(metrics)
	}
}

func BenchmarkStreamingParquet(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "bench-pq")
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, _ := NewExporter(tmpDir, sessionUUID, true, true, "parquet")
	defer exp.CloseStream()

	metrics := &collectors.DynamicMetrics{
		Timestamp:       1000,
		CPUTimeUserMode: 100,
	}

	// First write to initialize
	exp.SaveSnapshot(metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Timestamp = int64(1000 + i)
		exp.SaveSnapshot(metrics)
	}
}

func BenchmarkStreamingCSV(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "bench-csv")
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, _ := NewExporter(tmpDir, sessionUUID, true, true, "csv")
	defer exp.CloseStream()

	metrics := &collectors.DynamicMetrics{
		Timestamp:       1000,
		CPUTimeUserMode: 100,
	}

	// First write to initialize
	exp.SaveSnapshot(metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Timestamp = int64(1000 + i)
		exp.SaveSnapshot(metrics)
	}
}

func BenchmarkBatchJSONL(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "bench-batch")
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, _ := NewExporter(tmpDir, sessionUUID, true, false, "jsonl")

	metrics := &collectors.DynamicMetrics{
		Timestamp:       1000,
		CPUTimeUserMode: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Timestamp = int64(1000 + i)
		exp.SaveSnapshot(metrics)
	}
	exp.ProcessSession()
}
