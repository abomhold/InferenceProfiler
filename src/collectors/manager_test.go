package collectors

import (
	"os"
	"testing"
)

// ============================================================================
// Dynamic Collection Benchmarks (require /proc filesystem)
// ============================================================================

func BenchmarkCollectMemoryDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/meminfo"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/meminfo not available")
	}

	m := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CollectMemoryDynamic(m)
	}
}

func BenchmarkCollectCPUDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/stat"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/stat not available")
	}

	m := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CollectCPUDynamic(m)
	}
}

func BenchmarkCollectDiskDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/diskstats"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/diskstats not available")
	}

	m := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CollectDiskDynamic(m)
	}
}

func BenchmarkCollectNetworkDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/net/dev"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/net/dev not available")
	}

	m := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CollectNetworkDynamic(m)
	}
}

func BenchmarkCollectContainerDynamic(b *testing.B) {
	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		b.Skip("Skipping: /sys/fs/cgroup not available")
	}

	m := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CollectContainerDynamic(m)
	}
}

func BenchmarkCollectProcessesDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/1/stat"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/[pid]/stat not available")
	}

	m := &DynamicMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Processes = nil // Reset slice to avoid accumulation
		CollectProcessesDynamic(m)
	}
}

// BenchmarkCollectAllDynamic benchmarks collecting all dynamic metrics
func BenchmarkCollectAllDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/meminfo"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc not available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := &DynamicMetrics{}
		CollectMemoryDynamic(m)
		CollectCPUDynamic(m)
		CollectDiskDynamic(m)
		CollectNetworkDynamic(m)
		CollectContainerDynamic(m)
		CollectProcessesDynamic(m)
	}
}
