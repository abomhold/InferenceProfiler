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

func BenchmarkNvidiaCollector_Sequential(b *testing.B) {
	c := NewNvidiaCollector(false, false)
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

func BenchmarkNvidiaCollector_Concurrent(b *testing.B) {
	c := NewNvidiaCollector(false, true)
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
	c := NewNvidiaCollector(true, false)
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

func BenchmarkManager_VMOnly_Sequential(b *testing.B) {
	cfg := &utils.Config{
		DisableVM:        false,
		DisableContainer: true,
		DisableProcess:   true,
		DisableNvidia:    true,
		DisableVLLM:      true,
		Concurrent:       false,
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
		DisableVM:        false,
		DisableContainer: true,
		DisableProcess:   true,
		DisableNvidia:    true,
		DisableVLLM:      true,
		Concurrent:       true,
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
		DisableVM:           false,
		DisableContainer:    false,
		DisableProcess:      false,
		DisableNvidia:       false,
		DisableVLLM:         false,
		DisableGPUProcesses: false,
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
		DisableVM:           false,
		DisableContainer:    false,
		DisableProcess:      false,
		DisableNvidia:       false,
		DisableVLLM:         false,
		DisableGPUProcesses: false,
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
		DisableVM:        false,
		DisableContainer: false,
		DisableProcess:   false,
		DisableNvidia:    true,
		DisableVLLM:      true,
		Concurrent:       false,
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
		DisableVM:        false,
		DisableContainer: false,
		DisableProcess:   false,
		DisableNvidia:    true,
		DisableVLLM:      true,
		Concurrent:       true,
	}
	m := NewManager(cfg)
	defer m.Close()
	m.CollectStatic(&StaticMetrics{UUID: "bench", Hostname: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

func BenchmarkManager_Flatten_Sequential(b *testing.B) {
	cfg := &utils.Config{
		DisableVM:           false,
		DisableContainer:    false,
		DisableProcess:      false,
		DisableNvidia:       false,
		DisableVLLM:         false,
		DisableGPUProcesses: false,
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
		DisableVM:           false,
		DisableContainer:    false,
		DisableProcess:      false,
		DisableNvidia:       false,
		DisableVLLM:         false,
		DisableGPUProcesses: false,
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
