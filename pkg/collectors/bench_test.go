package collectors_test

import (
	"testing"

	"InferenceProfiler/pkg/collectors"
	"InferenceProfiler/pkg/collectors/container"
	"InferenceProfiler/pkg/collectors/nvidia"
	"InferenceProfiler/pkg/collectors/process"
	"InferenceProfiler/pkg/collectors/vllm"
	"InferenceProfiler/pkg/collectors/vm/cpu"
	"InferenceProfiler/pkg/collectors/vm/disk"
	"InferenceProfiler/pkg/collectors/vm/memory"
	"InferenceProfiler/pkg/collectors/vm/network"
	"InferenceProfiler/pkg/config"
)

func BenchmarkCPUCollector_Static(b *testing.B) {
	c := cpu.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic()
	}
}

func BenchmarkCPUCollector_Dynamic(b *testing.B) {
	c := cpu.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkMemoryCollector_Static(b *testing.B) {
	c := memory.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic()
	}
}

func BenchmarkMemoryCollector_Dynamic(b *testing.B) {
	c := memory.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkDiskCollector_Static(b *testing.B) {
	c := disk.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic()
	}
}

func BenchmarkDiskCollector_Dynamic(b *testing.B) {
	c := disk.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkNetworkCollector_Static(b *testing.B) {
	c := network.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic()
	}
}

func BenchmarkNetworkCollector_Dynamic(b *testing.B) {
	c := network.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkContainerCollector_Static(b *testing.B) {
	c := container.New()
	if c == nil {
		b.Skip("container collector not available")
	}
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic()
	}
}

func BenchmarkContainerCollector_Dynamic(b *testing.B) {
	c := container.New()
	if c == nil {
		b.Skip("container collector not available")
	}
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkProcessCollector_Dynamic(b *testing.B) {
	c := process.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkNvidiaCollector_Static(b *testing.B) {
	c := nvidia.New(false)
	if c == nil {
		b.Skip("nvidia collector not available")
	}
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic()
	}
}

func BenchmarkNvidiaCollector_Dynamic(b *testing.B) {
	c := nvidia.New(false)
	if c == nil {
		b.Skip("nvidia collector not available")
	}
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkNvidiaCollector_DynamicWithProcs(b *testing.B) {
	c := nvidia.New(true)
	if c == nil {
		b.Skip("nvidia collector not available")
	}
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkVLLMCollector_Dynamic(b *testing.B) {
	c := vllm.New()
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

// Benchmark the full manager collection cycle
func BenchmarkManager_CollectDynamic_VMOnly(b *testing.B) {
	cfg := &config.Config{
		EnableVM:        true,
		EnableContainer: false,
		EnableProcess:   false,
		EnableNvidia:    false,
		EnableVLLM:      false,
	}
	m := collectors.NewManager(cfg)
	defer m.Close()

	base := &collectors.BaseStatic{UUID: "test", Hostname: "test"}
	m.CollectStatic(base)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&collectors.BaseDynamic{})
	}
}

func BenchmarkManager_CollectDynamic_All(b *testing.B) {
	cfg := &config.Config{
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       true,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
	}
	m := collectors.NewManager(cfg)
	defer m.Close()

	base := &collectors.BaseStatic{UUID: "test", Hostname: "test"}
	m.CollectStatic(base)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&collectors.BaseDynamic{})
	}
}

func BenchmarkManager_CollectDynamic_NoProcess(b *testing.B) {
	cfg := &config.Config{
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       false,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
	}
	m := collectors.NewManager(cfg)
	defer m.Close()

	base := &collectors.BaseStatic{UUID: "test", Hostname: "test"}
	m.CollectStatic(base)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&collectors.BaseDynamic{})
	}
}

// Benchmark ToRecord conversion
func BenchmarkToRecord_CPUDynamic(b *testing.B) {
	c := cpu.New()
	defer c.Close()
	// Warm up to get a real record
	c.CollectDynamic()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

// Benchmark MergeRecords
func BenchmarkMergeRecords_5Records(b *testing.B) {
	records := make([]collectors.Record, 5)
	for i := range records {
		records[i] = collectors.Record{
			"field1": int64(123),
			"field2": "string",
			"field3": float64(1.23),
			"field4": true,
			"field5": int64(456),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectors.MergeRecords(records...)
	}
}

func BenchmarkMergeRecords_10Records(b *testing.B) {
	records := make([]collectors.Record, 10)
	for i := range records {
		records[i] = collectors.Record{
			"field1":  int64(123),
			"field2":  "string",
			"field3":  float64(1.23),
			"field4":  true,
			"field5":  int64(456),
			"field6":  int64(789),
			"field7":  "another",
			"field8":  float64(4.56),
			"field9":  false,
			"field10": int64(101112),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectors.MergeRecords(records...)
	}
}

// Individual collector isolation benchmarks
func BenchmarkIsolated_CPU(b *testing.B) {
	c := cpu.New()
	defer c.Close()
	c.CollectStatic() // warm up
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkIsolated_Memory(b *testing.B) {
	c := memory.New()
	defer c.Close()
	c.CollectStatic()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkIsolated_Disk(b *testing.B) {
	c := disk.New()
	defer c.Close()
	c.CollectStatic()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}

func BenchmarkIsolated_Network(b *testing.B) {
	c := network.New()
	defer c.Close()
	c.CollectStatic()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic()
	}
}
