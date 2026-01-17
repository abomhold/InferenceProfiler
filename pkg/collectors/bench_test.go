package collectors

import (
	"testing"

	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/formatting"
)

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

func BenchmarkMemoryCollector_Static(b *testing.B) {
	c := NewMemoryCollector()
	defer c.Close()
	s := &StaticMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic(s)
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

func BenchmarkDiskCollector_Static(b *testing.B) {
	c := NewDiskCollector()
	defer c.Close()
	s := &StaticMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic(s)
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

func BenchmarkNetworkCollector_Static(b *testing.B) {
	c := NewNetworkCollector()
	defer c.Close()
	s := &StaticMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic(s)
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

func BenchmarkContainerCollector_Static(b *testing.B) {
	c := NewContainerCollector()
	if c == nil {
		b.Skip("container collector not available")
	}
	defer c.Close()
	s := &StaticMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic(s)
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

func BenchmarkProcessCollector_Dynamic(b *testing.B) {
	c := NewProcessCollector()
	defer c.Close()
	d := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectDynamic(d)
	}
}

func BenchmarkNvidiaCollector_Static(b *testing.B) {
	c := NewNvidiaCollector(false)
	if c == nil {
		b.Skip("nvidia collector not available")
	}
	defer c.Close()
	s := &StaticMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CollectStatic(s)
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

func BenchmarkManager_CollectDynamic_VMOnly(b *testing.B) {
	cfg := &config.Config{
		EnableVM:        true,
		EnableContainer: false,
		EnableProcess:   false,
		EnableNvidia:    false,
		EnableVLLM:      false,
	}
	m := NewManager(cfg)
	defer m.Close()

	s := &StaticMetrics{UUID: "test", Hostname: "test"}
	m.CollectStatic(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
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
	m := NewManager(cfg)
	defer m.Close()

	s := &StaticMetrics{UUID: "test", Hostname: "test"}
	m.CollectStatic(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CollectDynamic(&DynamicMetrics{})
	}
}

func BenchmarkManager_CollectAndFlatten_All(b *testing.B) {
	cfg := &config.Config{
		EnableVM:            true,
		EnableContainer:     true,
		EnableProcess:       true,
		EnableNvidia:        true,
		EnableVLLM:          true,
		CollectGPUProcesses: true,
	}
	m := NewManager(cfg)
	defer m.Close()

	s := &StaticMetrics{UUID: "test", Hostname: "test"}
	m.CollectStatic(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := m.CollectDynamic(&DynamicMetrics{})
		formatting.FlattenRecord(record)
	}
}

func BenchmarkManager_CollectAndFlatten_VMOnly(b *testing.B) {
	cfg := &config.Config{
		EnableVM:        true,
		EnableContainer: false,
		EnableProcess:   false,
		EnableNvidia:    false,
		EnableVLLM:      false,
	}
	m := NewManager(cfg)
	defer m.Close()

	s := &StaticMetrics{UUID: "test", Hostname: "test"}
	m.CollectStatic(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := m.CollectDynamic(&DynamicMetrics{})
		formatting.FlattenRecord(record)
	}
}
