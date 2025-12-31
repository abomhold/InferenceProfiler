package output

import (
	"InferenceProfiler/src/collectors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/parquet-go/parquet-go"
)

// TestStreamingParquet tests the streaming parquet mode
func TestStreamingParquet(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "profiler-stream-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()

	// Create exporter in streaming mode
	exp, err := NewExporter(tmpDir, sessionUUID, true, true)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}
	defer exp.CloseStream()

	metrics1 := &collectors.DynamicMetrics{
		Timestamp:       1735166000000,
		CPUTimeUserMode: 12345,
		MemoryTotal:     32768000000,
		LoadAvg:         2.45,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{Index: 0, UtilizationGPU: 0, MemoryUsedMb: 0},
		},
	}

	metrics2 := &collectors.DynamicMetrics{
		Timestamp:       1735166001000,
		CPUTimeUserMode: 12346,
		MemoryTotal:     32768000000,
		LoadAvg:         2.46,
	}

	metrics3 := &collectors.DynamicMetrics{
		Timestamp:       1735166002000,
		CPUTimeUserMode: 12347,
		MemoryTotal:     32768000000,
		LoadAvg:         2.47,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{Index: 0, UtilizationGPU: 85, MemoryUsedMb: 8192},
		},
	}

	// Write snapshots
	if err := exp.SaveSnapshot(metrics1); err != nil {
		t.Fatalf("Failed to save snapshot 1: %v", err)
	}
	if err := exp.SaveSnapshot(metrics2); err != nil {
		t.Fatalf("Failed to save snapshot 2: %v", err)
	}
	if err := exp.SaveSnapshot(metrics3); err != nil {
		t.Fatalf("Failed to save snapshot 3: %v", err)
	}

	// Close stream
	if err := exp.CloseStream(); err != nil {
		t.Fatalf("Failed to close stream: %v", err)
	}

	// Verify file exists
	parquetPath := filepath.Join(tmpDir, sessionUUID.String()+"-stream.parquet")
	if _, err := os.Stat(parquetPath); os.IsNotExist(err) {
		t.Fatalf("Parquet file not created: %s", parquetPath)
	}

	// Read and verify parquet file
	file, err := os.Open(parquetPath)
	if err != nil {
		t.Fatalf("Failed to open parquet file: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat parquet file: %v", err)
	}

	pf, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		t.Fatalf("Failed to open parquet file: %v", err)
	}

	reader := parquet.NewReader(pf)

	// Read all rows
	rows := make([]map[string]interface{}, 0, 3)
	for {
		row := make(map[string]interface{})
		err := reader.Read(&row)
		if err != nil {
			break
		}
		rows = append(rows, row)
	}

	// Verify row count
	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	// Verify first row data
	if rows[0]["timestamp"] != int64(1735166000000) {
		t.Errorf("First row timestamp = %v; want 1735166000000", rows[0]["timestamp"])
	}

	// Verify GPU data in third row
	if rows[2]["nvidiaGpuCount"] != int64(1) {
		t.Errorf("Third row nvidiaGpuCount = %v; want 1", rows[2]["nvidiaGpuCount"])
	}
	if rows[2]["nvidia0UtilizationGpu"] != int64(85) {
		t.Errorf("Third row nvidia0UtilizationGpu = %v; want 85", rows[2]["nvidia0UtilizationGpu"])
	}
}

// TestStreamingVsBatch compares output between streaming and batch modes
func TestStreamingVsBatch(t *testing.T) {
	// Create temp directories
	streamDir, err := os.MkdirTemp("", "profiler-stream")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(streamDir)

	batchDir, err := os.MkdirTemp("", "profiler-batch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(batchDir)

	sessionUUID := uuid.New()

	// Create test metrics
	metrics := []*collectors.DynamicMetrics{
		{
			Timestamp:       1735166000000,
			CPUTimeUserMode: 12345,
			MemoryTotal:     32768000000,
			NvidiaGPUs: []collectors.NvidiaGPUDynamic{
				{Index: 0, UtilizationGPU: 85, MemoryUsedMb: 8192},
			},
		},
		{
			Timestamp:       1735166001000,
			CPUTimeUserMode: 12346,
			MemoryTotal:     32768000000,
			NvidiaGPUs: []collectors.NvidiaGPUDynamic{
				{Index: 0, UtilizationGPU: 86, MemoryUsedMb: 8193},
			},
		},
	}

	// Streaming mode
	streamExp, err := NewExporter(streamDir, sessionUUID, true, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metrics {
		if err := streamExp.SaveSnapshot(m); err != nil {
			t.Fatal(err)
		}
	}
	streamExp.CloseStream()

	// Batch mode
	batchExp, err := NewExporter(batchDir, sessionUUID, true, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metrics {
		if err := batchExp.SaveSnapshot(m); err != nil {
			t.Fatal(err)
		}
	}
	if err := batchExp.ProcessSession("parquet"); err != nil {
		t.Fatal(err)
	}

	// Read both files
	streamFile := filepath.Join(streamDir, sessionUUID.String()+"-stream.parquet")
	batchFile := filepath.Join(batchDir, sessionUUID.String()+".parquet")

	streamRows := readParquetFile(t, streamFile)
	batchRows := readParquetFile(t, batchFile)

	// Compare row counts
	if len(streamRows) != len(batchRows) {
		t.Errorf("Row count mismatch: stream=%d, batch=%d", len(streamRows), len(batchRows))
	}

	// Compare data (timestamps should match)
	for i := range streamRows {
		if streamRows[i]["timestamp"] != batchRows[i]["timestamp"] {
			t.Errorf("Row %d timestamp mismatch: stream=%v, batch=%v",
				i, streamRows[i]["timestamp"], batchRows[i]["timestamp"])
		}
	}
}

// TestStreamingNoSnapshots tests streaming mode with no data written
func TestStreamingNoSnapshots(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler-stream-empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()

	// Create exporter but don't write any snapshots
	exp, err := NewExporter(tmpDir, sessionUUID, true, true)
	if err != nil {
		t.Fatal(err)
	}

	// Close without writing - should not create parquet file
	if err := exp.CloseStream(); err != nil {
		t.Fatal(err)
	}

	// Verify no parquet file created
	parquetPath := filepath.Join(tmpDir, sessionUUID.String()+"-stream.parquet")
	if _, err := os.Stat(parquetPath); !os.IsNotExist(err) {
		t.Error("Parquet file should not exist when no snapshots written")
	}
}

// TestStreamingMultipleClose tests that multiple close calls are safe
func TestStreamingMultipleClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler-stream-close")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()

	exp, err := NewExporter(tmpDir, sessionUUID, true, true)
	if err != nil {
		t.Fatal(err)
	}

	// Write one snapshot
	metrics := &collectors.DynamicMetrics{
		Timestamp:       1735166000000,
		CPUTimeUserMode: 12345,
	}
	if err := exp.SaveSnapshot(metrics); err != nil {
		t.Fatal(err)
	}

	// Close multiple times - should not panic or error
	if err := exp.CloseStream(); err != nil {
		t.Fatalf("First close failed: %v", err)
	}
	if err := exp.CloseStream(); err != nil {
		t.Fatalf("Second close failed: %v", err)
	}
	if err := exp.CloseStream(); err != nil {
		t.Fatalf("Third close failed: %v", err)
	}
}

// Helper function to read parquet file
func readParquetFile(t *testing.T, path string) []map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", path, err)
	}
	defer file.Close()

	// Get file info for size
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Create parquet file reader
	pf, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		t.Fatalf("Failed to open parquet file: %v", err)
	}

	// Read all rows
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

// BenchmarkStreamingWrite benchmarks streaming write performance
func BenchmarkStreamingWrite(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "profiler-bench")
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, _ := NewExporter(tmpDir, sessionUUID, true, true)
	defer exp.CloseStream()

	metrics := &collectors.DynamicMetrics{
		Timestamp:       1735166000000,
		CPUTimeUserMode: 12345,
		MemoryTotal:     32768000000,
		LoadAvg:         2.45,
	}

	// First write initializes schema
	exp.SaveSnapshot(metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Timestamp = int64(1735166000000 + i)
		exp.SaveSnapshot(metrics)
	}
}

// BenchmarkBatchWrite benchmarks batch write performance
func BenchmarkBatchWrite(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "profiler-bench")
	defer os.RemoveAll(tmpDir)

	sessionUUID := uuid.New()
	exp, _ := NewExporter(tmpDir, sessionUUID, true, false)

	metrics := &collectors.DynamicMetrics{
		Timestamp:       1735166000000,
		CPUTimeUserMode: 12345,
		MemoryTotal:     32768000000,
		LoadAvg:         2.45,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.Timestamp = int64(1735166000000 + i)
		exp.SaveSnapshot(metrics)
	}
	exp.ProcessSession("parquet")
}
