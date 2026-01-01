package output

import (
	"InferenceProfiler/src/collectors"
	"encoding/json"
	"testing"
)

func TestFlattenMetrics_ScalarFields(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp:       1735166000000,
		CPUTimeUserMode: 12345678,
		MemoryTotal:     32768000000,
		LoadAvg:         2.45,
		VLLMAvailable:   true,
	}

	flat := FlattenMetrics(m)

	// Check scalar fields are preserved
	if flat["timestamp"] != int64(1735166000000) {
		t.Errorf("timestamp = %v; want 1735166000000", flat["timestamp"])
	}
	if flat["vCpuTimeUserMode"] != int64(12345678) {
		t.Errorf("vCpuTimeUserMode = %v; want 12345678", flat["vCpuTimeUserMode"])
	}
	if flat["vMemoryTotal"] != int64(32768000000) {
		t.Errorf("vMemoryTotal = %v; want 32768000000", flat["vMemoryTotal"])
	}
	if flat["vLoadAvg"] != 2.45 {
		t.Errorf("vLoadAvg = %v; want 2.45", flat["vLoadAvg"])
	}
	if flat["vllmAvailable"] != true {
		t.Errorf("vllmAvailable = %v; want true", flat["vllmAvailable"])
	}
}

func TestFlattenMetrics_NoGPUs(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp:  1735166000000,
		NvidiaGPUs: nil,
	}

	flat := FlattenMetrics(m)

	if flat["nvidiaGpuCount"] != 0 {
		t.Errorf("nvidiaGpuCount = %v; want 0", flat["nvidiaGpuCount"])
	}

	// Should not have any nvidia0* keys
	for key := range flat {
		if len(key) >= 7 && key[:7] == "nvidia0" {
			t.Errorf("Found unexpected GPU key: %s", key)
		}
	}
}

func TestFlattenMetrics_SingleGPU(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{
				Index:            0,
				UtilizationGPU:   85,
				UtilizationGPUT:  1735166000001,
				MemoryUsedBytes:  8589934592, // 8 GiB in bytes
				MemoryFreeBytes:  8589934592,
				TemperatureGpuC:  65,
				PowerUsageMw:     245500, // milliwatts
				PerformanceState: 0,
				ProcessCount:     3,
				ProcessesJSON:    `[{"pid":1234,"name":"python","usedMemoryBytes":4294967296}]`,
			},
		},
	}

	flat := FlattenMetrics(m)

	if flat["nvidiaGpuCount"] != 1 {
		t.Errorf("nvidiaGpuCount = %v; want 1", flat["nvidiaGpuCount"])
	}
	if flat["nvidia0UtilizationGpu"] != int64(85) {
		t.Errorf("nvidia0UtilizationGpu = %v; want 85", flat["nvidia0UtilizationGpu"])
	}
	if flat["nvidia0MemoryUsedBytes"] != int64(8589934592) {
		t.Errorf("nvidia0MemoryUsedBytes = %v; want 8589934592", flat["nvidia0MemoryUsedBytes"])
	}
	if flat["nvidia0TemperatureGpuC"] != int64(65) {
		t.Errorf("nvidia0TemperatureGpuC = %v; want 65", flat["nvidia0TemperatureGpuC"])
	}
	if flat["nvidia0PowerUsageMw"] != int64(245500) {
		t.Errorf("nvidia0PowerUsageMw = %v; want 245500", flat["nvidia0PowerUsageMw"])
	}
	if flat["nvidia0PerformanceState"] != 0 {
		t.Errorf("nvidia0PerformanceState = %v; want 0", flat["nvidia0PerformanceState"])
	}
	if flat["nvidia0ProcessesJson"] != `[{"pid":1234,"name":"python","usedMemoryBytes":4294967296}]` {
		t.Errorf("nvidia0ProcessesJson = %v", flat["nvidia0ProcessesJson"])
	}
}

func TestFlattenMetrics_MultipleGPUs(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{Index: 0, UtilizationGPU: 85, MemoryUsedBytes: 8589934592},
			{Index: 1, UtilizationGPU: 72, MemoryUsedBytes: 6442450944},
			{Index: 2, UtilizationGPU: 90, MemoryUsedBytes: 10737418240},
		},
	}

	flat := FlattenMetrics(m)

	if flat["nvidiaGpuCount"] != 3 {
		t.Errorf("nvidiaGpuCount = %v; want 3", flat["nvidiaGpuCount"])
	}

	// Check each GPU
	gpuTests := []struct {
		prefix  string
		util    int64
		memUsed int64
	}{
		{"nvidia0", 85, 8589934592},
		{"nvidia1", 72, 6442450944},
		{"nvidia2", 90, 10737418240},
	}

	for _, tt := range gpuTests {
		if flat[tt.prefix+"UtilizationGpu"] != tt.util {
			t.Errorf("%sUtilizationGpu = %v; want %d", tt.prefix, flat[tt.prefix+"UtilizationGpu"], tt.util)
		}
		if flat[tt.prefix+"MemoryUsedBytes"] != tt.memUsed {
			t.Errorf("%sMemoryUsedBytes = %v; want %d", tt.prefix, flat[tt.prefix+"MemoryUsedBytes"], tt.memUsed)
		}
	}
}

func TestFlattenMetrics_NoProcesses(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: nil,
	}

	flat := FlattenMetrics(m)

	if flat["processCount"] != 0 {
		t.Errorf("processCount = %v; want 0", flat["processCount"])
	}
}

func TestFlattenMetrics_SingleProcess(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: []collectors.ProcessMetrics{
			{
				PID:               1234,
				PIDT:              1735166000001,
				Name:              "python3",
				Cmdline:           "python3 train.py --epochs 10",
				NumThreads:        8,
				CPUTimeUserMode:   789000,
				CPUTimeKernelMode: 123000,
				ResidentSetSize:   4294967296,
			},
		},
	}

	flat := FlattenMetrics(m)

	if flat["processCount"] != 1 {
		t.Errorf("processCount = %v; want 1", flat["processCount"])
	}
	if flat["process0Pid"] != int64(1234) {
		t.Errorf("process0Pid = %v; want 1234", flat["process0Pid"])
	}
	if flat["process0Name"] != "python3" {
		t.Errorf("process0Name = %v; want python3", flat["process0Name"])
	}
	if flat["process0Cmdline"] != "python3 train.py --epochs 10" {
		t.Errorf("process0Cmdline = %v", flat["process0Cmdline"])
	}
	if flat["process0NumThreads"] != int64(8) {
		t.Errorf("process0NumThreads = %v; want 8", flat["process0NumThreads"])
	}
	if flat["process0ResidentSetSize"] != int64(4294967296) {
		t.Errorf("process0ResidentSetSize = %v; want 4294967296", flat["process0ResidentSetSize"])
	}
}

func TestFlattenMetrics_MultipleProcesses(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: []collectors.ProcessMetrics{
			{PID: 1, Name: "systemd"},
			{PID: 2, Name: "kthreadd"},
			{PID: 1234, Name: "python3"},
		},
	}

	flat := FlattenMetrics(m)

	if flat["processCount"] != 3 {
		t.Errorf("processCount = %v; want 3", flat["processCount"])
	}
	if flat["process0Pid"] != int64(1) {
		t.Errorf("process0Pid = %v; want 1", flat["process0Pid"])
	}
	if flat["process1Name"] != "kthreadd" {
		t.Errorf("process1Name = %v; want kthreadd", flat["process1Name"])
	}
	if flat["process2Pid"] != int64(1234) {
		t.Errorf("process2Pid = %v; want 1234", flat["process2Pid"])
	}
}

// ============================================================================
// ToJSONMode Tests
// ============================================================================

func TestToJSONMode_ScalarFields(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp:       1735166000000,
		CPUTimeUserMode: 12345678,
		MemoryTotal:     32768000000,
		LoadAvg:         2.45,
	}

	result := ToJSONMode(m)

	// Check scalar fields are preserved (same as flatten mode)
	if result["timestamp"] != int64(1735166000000) {
		t.Errorf("timestamp = %v; want 1735166000000", result["timestamp"])
	}
	if result["vCpuTimeUserMode"] != int64(12345678) {
		t.Errorf("vCpuTimeUserMode = %v; want 12345678", result["vCpuTimeUserMode"])
	}
}

func TestToJSONMode_GPUsAsJSON(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{Index: 0, UtilizationGPU: 85, MemoryUsedBytes: 8589934592},
			{Index: 1, UtilizationGPU: 72, MemoryUsedBytes: 6442450944},
		},
	}

	result := ToJSONMode(m)

	if result["nvidiaGpuCount"] != 2 {
		t.Errorf("nvidiaGpuCount = %v; want 2", result["nvidiaGpuCount"])
	}

	// Should have nvidiaGpusJson as a string
	gpusJSON, ok := result["nvidiaGpusJson"].(string)
	if !ok {
		t.Fatalf("nvidiaGpusJson is not a string: %T", result["nvidiaGpusJson"])
	}

	// Parse the JSON to verify contents
	var gpus []collectors.NvidiaGPUDynamic
	if err := json.Unmarshal([]byte(gpusJSON), &gpus); err != nil {
		t.Fatalf("Failed to parse nvidiaGpusJson: %v", err)
	}

	if len(gpus) != 2 {
		t.Errorf("Parsed %d GPUs; want 2", len(gpus))
	}
	if gpus[0].UtilizationGPU != 85 {
		t.Errorf("GPU 0 utilization = %d; want 85", gpus[0].UtilizationGPU)
	}
	if gpus[1].MemoryUsedBytes != 6442450944 {
		t.Errorf("GPU 1 memoryUsedBytes = %d; want 6442450944", gpus[1].MemoryUsedBytes)
	}

	// Should NOT have flattened nvidia0* keys
	if _, exists := result["nvidia0UtilizationGpu"]; exists {
		t.Error("ToJSONMode should not have flattened GPU keys")
	}
}

func TestToJSONMode_ProcessesAsJSON(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: []collectors.ProcessMetrics{
			{PID: 1, Name: "systemd", ResidentSetSize: 8388608},
			{PID: 1234, Name: "python3", ResidentSetSize: 4294967296},
		},
	}

	result := ToJSONMode(m)

	if result["processCount"] != 2 {
		t.Errorf("processCount = %v; want 2", result["processCount"])
	}

	// Should have processesJson as a string
	procsJSON, ok := result["processesJson"].(string)
	if !ok {
		t.Fatalf("processesJson is not a string: %T", result["processesJson"])
	}

	// Parse the JSON to verify contents
	var procs []collectors.ProcessMetrics
	if err := json.Unmarshal([]byte(procsJSON), &procs); err != nil {
		t.Fatalf("Failed to parse processesJson: %v", err)
	}

	if len(procs) != 2 {
		t.Errorf("Parsed %d processes; want 2", len(procs))
	}
	if procs[0].Name != "systemd" {
		t.Errorf("Process 0 name = %s; want systemd", procs[0].Name)
	}
	if procs[1].PID != 1234 {
		t.Errorf("Process 1 PID = %d; want 1234", procs[1].PID)
	}

	// Should NOT have flattened process0* keys
	if _, exists := result["process0Pid"]; exists {
		t.Error("ToJSONMode should not have flattened process keys")
	}
}

func TestToJSONMode_NoGPUs(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp:  1735166000000,
		NvidiaGPUs: nil,
	}

	result := ToJSONMode(m)

	if result["nvidiaGpuCount"] != 0 {
		t.Errorf("nvidiaGpuCount = %v; want 0", result["nvidiaGpuCount"])
	}

	// Should NOT have nvidiaGpusJson key when empty
	if _, exists := result["nvidiaGpusJson"]; exists {
		t.Error("nvidiaGpusJson should not exist when no GPUs")
	}
}

func TestToJSONMode_NoProcesses(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: nil,
	}

	result := ToJSONMode(m)

	if result["processCount"] != 0 {
		t.Errorf("processCount = %v; want 0", result["processCount"])
	}

	// Should NOT have processesJson key when empty
	if _, exists := result["processesJson"]; exists {
		t.Error("processesJson should not exist when no processes")
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestFlattenMetrics_EmptyMetrics(t *testing.T) {
	m := &collectors.DynamicMetrics{}

	flat := FlattenMetrics(m)

	// Should still have basic keys
	if _, exists := flat["timestamp"]; !exists {
		t.Error("timestamp key should exist even for empty metrics")
	}
	if flat["nvidiaGpuCount"] != 0 {
		t.Errorf("nvidiaGpuCount = %v; want 0", flat["nvidiaGpuCount"])
	}
	if flat["processCount"] != 0 {
		t.Errorf("processCount = %v; want 0", flat["processCount"])
	}
}

func TestFlattenMetrics_VLLMFields(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp:               1735166000000,
		VLLMAvailable:           true,
		VLLMRequestsRunning:     5.0,
		VLLMKvCacheUsagePercent: 0.75,
		VLLMLatencyTtftSum:      1.234,
		VLLMLatencyTtftCount:    100.0,
		VLLMHistogramsJSON:      `{"latencyTtft":{"0.1":50,"0.5":90,"inf":100}}`,
	}

	flat := FlattenMetrics(m)

	if flat["vllmAvailable"] != true {
		t.Errorf("vllmAvailable = %v; want true", flat["vllmAvailable"])
	}
	if flat["vllmRequestsRunning"] != 5.0 {
		t.Errorf("vllmRequestsRunning = %v; want 5.0", flat["vllmRequestsRunning"])
	}
	if flat["vllmKvCacheUsagePercent"] != 0.75 {
		t.Errorf("vllmKvCacheUsagePercent = %v; want 0.75", flat["vllmKvCacheUsagePercent"])
	}
	if flat["vllmHistogramsJson"] != `{"latencyTtft":{"0.1":50,"0.5":90,"inf":100}}` {
		t.Errorf("vllmHistogramsJson = %v", flat["vllmHistogramsJson"])
	}
}

func TestFlattenMetrics_ContainerFields(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp:                1735166000000,
		ContainerCPUTime:         123456789000,
		ContainerMemoryUsed:      8589934592,
		ContainerPgFault:         987654321,
		ContainerPerCPUTimesJSON: `[12345678901234,23456789012345]`,
	}

	flat := FlattenMetrics(m)

	if flat["cCpuTime"] != int64(123456789000) {
		t.Errorf("cCpuTime = %v; want 123456789000", flat["cCpuTime"])
	}
	if flat["cMemoryUsed"] != int64(8589934592) {
		t.Errorf("cMemoryUsed = %v; want 8589934592", flat["cMemoryUsed"])
	}
	if flat["cCpuPerCpuJson"] != `[12345678901234,23456789012345]` {
		t.Errorf("cCpuPerCpuJson = %v", flat["cCpuPerCpuJson"])
	}
}

func TestFlattenMetrics_GPUWithEmptyProcessesJSON(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{Index: 0, UtilizationGPU: 50, ProcessesJSON: ""},
		},
	}

	flat := FlattenMetrics(m)

	// ProcessesJSON should not be in the output if empty
	if _, exists := flat["nvidia0ProcessesJson"]; exists {
		t.Error("nvidia0ProcessesJson should not exist when empty")
	}
}

func TestFlattenMetrics_ProcessWithSpecialCharacters(t *testing.T) {
	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: []collectors.ProcessMetrics{
			{
				PID:     1234,
				Name:    "Web Content",
				Cmdline: "/usr/bin/firefox --profile=/home/user/.mozilla/firefox/abc123.default",
			},
		},
	}

	flat := FlattenMetrics(m)

	if flat["process0Name"] != "Web Content" {
		t.Errorf("process0Name = %v; want 'Web Content'", flat["process0Name"])
	}
	if flat["process0Cmdline"] != "/usr/bin/firefox --profile=/home/user/.mozilla/firefox/abc123.default" {
		t.Errorf("process0Cmdline = %v", flat["process0Cmdline"])
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkFlattenMetrics_Simple(b *testing.B) {
	m := &collectors.DynamicMetrics{
		Timestamp:         1735166000000,
		CPUTimeUserMode:   12345678,
		MemoryTotal:       32768000000,
		NetworkBytesRecvd: 987654321,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FlattenMetrics(m)
	}
}

func BenchmarkFlattenMetrics_WithGPUs(b *testing.B) {
	m := &collectors.DynamicMetrics{
		Timestamp:         1735166000000,
		CPUTimeUserMode:   12345678,
		MemoryTotal:       32768000000,
		NetworkBytesRecvd: 987654321,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{Index: 0, UtilizationGPU: 85, MemoryUsedBytes: 8589934592},
			{Index: 1, UtilizationGPU: 72, MemoryUsedBytes: 6442450944},
			{Index: 2, UtilizationGPU: 90, MemoryUsedBytes: 10737418240},
			{Index: 3, UtilizationGPU: 65, MemoryUsedBytes: 4294967296},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FlattenMetrics(m)
	}
}

func BenchmarkFlattenMetrics_WithProcesses(b *testing.B) {
	procs := make([]collectors.ProcessMetrics, 100)
	for i := 0; i < 100; i++ {
		procs[i] = collectors.ProcessMetrics{
			PID:             int64(i + 1),
			Name:            "process",
			CPUTimeUserMode: 12345,
			ResidentSetSize: 1048576,
		}
	}

	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: procs,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FlattenMetrics(m)
	}
}

func BenchmarkToJSONMode_WithGPUs(b *testing.B) {
	m := &collectors.DynamicMetrics{
		Timestamp:         1735166000000,
		CPUTimeUserMode:   12345678,
		MemoryTotal:       32768000000,
		NetworkBytesRecvd: 987654321,
		NvidiaGPUs: []collectors.NvidiaGPUDynamic{
			{Index: 0, UtilizationGPU: 85, MemoryUsedBytes: 8589934592},
			{Index: 1, UtilizationGPU: 72, MemoryUsedBytes: 6442450944},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToJSONMode(m)
	}
}

func BenchmarkToJSONMode_WithProcesses(b *testing.B) {
	procs := make([]collectors.ProcessMetrics, 100)
	for i := 0; i < 100; i++ {
		procs[i] = collectors.ProcessMetrics{
			PID:             int64(i + 1),
			Name:            "process",
			CPUTimeUserMode: 12345,
			ResidentSetSize: 1048576,
		}
	}

	m := &collectors.DynamicMetrics{
		Timestamp: 1735166000000,
		Processes: procs,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToJSONMode(m)
	}
}
