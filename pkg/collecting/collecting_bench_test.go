package collecting

import (
	"io"
	"log"
	"testing"

	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
)

func init() {
	log.SetOutput(io.Discard)
}

// ============================================================================
// Individual Collector Benchmarks
// ============================================================================

func BenchmarkCPUCollector_Static(b *testing.B) {
	c := NewCPUCollector()
	defer c.Close()
	s := &StaticMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic(s)
	}
}

func BenchmarkCPUCollector_Dynamic(b *testing.B) {
	c := NewCPUCollector()
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkMemoryCollector_Dynamic(b *testing.B) {
	c := NewMemoryCollector()
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkDiskCollector_Dynamic(b *testing.B) {
	c := NewDiskCollector()
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkNetworkCollector_Dynamic(b *testing.B) {
	c := NewNetworkCollector()
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkContainerCollector_Dynamic(b *testing.B) {
	c := NewContainerCollector()
	if c == nil {
		b.Skip("container collector not available")
	}
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkProcessCollector_Sequential(b *testing.B) {
	c := NewProcessCollector(false)
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkProcessCollector_Concurrent(b *testing.B) {
	c := NewProcessCollector(true)
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkNvidiaCollector_Dynamic(b *testing.B) {
	c := NewNvidiaCollector(false)
	if c == nil {
		b.Skip("nvidia collector not available")
	}
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkNvidiaCollector_DynamicWithProcs(b *testing.B) {
	c := NewNvidiaCollector(true)
	if c == nil {
		b.Skip("nvidia collector not available")
	}
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkVLLMCollector_Dynamic(b *testing.B) {
	c := NewVLLMCollector()
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

// ============================================================================
// Manager Benchmarks - Sequential vs Concurrent
// ============================================================================

func BenchmarkManager_VMOnly_Sequential(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:        true,
		EnableContainer: false,
		EnableProcess:   false,
		EnableNvidia:    false,
		EnableVLLM:      false,
		Concurrent:      false,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

func BenchmarkManager_VMOnly_Concurrent(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:        true,
		EnableContainer: false,
		EnableProcess:   false,
		EnableNvidia:    false,
		EnableVLLM:      false,
		Concurrent:      true,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

func BenchmarkManager_All_Sequential(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       true,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
		Concurrent:          false,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

func BenchmarkManager_All_Concurrent(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       true,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
		Concurrent:          true,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

func BenchmarkManager_NoGPU_Sequential(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:        true,
		EnableContainer: true,
		EnableProcess:   true,
		EnableNvidia:    false,
		EnableVLLM:      false,
		Concurrent:      false,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

func BenchmarkManager_NoGPU_Concurrent(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:        true,
		EnableContainer: true,
		EnableProcess:   true,
		EnableNvidia:    false,
		EnableVLLM:      false,
		Concurrent:      true,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

// ============================================================================
// End-to-End Benchmarks (Collect + Flatten)
// ============================================================================

func BenchmarkManager_Flatten_Sequential(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       true,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
		Concurrent:          false,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := m.CollectDynamic(&DynamicMetrics{})
		exporting.FlattenRecord(record)
	}
}

func BenchmarkManager_Flatten_Concurrent(b *testing.B) {
	cfg := &utils.Config{
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       true,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
		Concurrent:          true,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := m.CollectDynamic(&DynamicMetrics{})
		exporting.FlattenRecord(record)
	}
}
