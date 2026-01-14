package collectors

import (
	"io"
	"log"
	"os"
	"testing"

	"InferenceProfiler/src/config"
)

func init() {
	// Disable logging during benchmarks to prevent "Registered collector" spam
	log.SetOutput(io.Discard)
}

// ============================================================================
// Dynamic Collection Benchmarks (using CollectorManager)
// ============================================================================

func BenchmarkCollectMemoryDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/meminfo"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/meminfo not available")
	}

	// Configure Manager with ONLY Memory enabled
	cfg := config.CollectorConfig{Memory: true}
	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}

func BenchmarkNvidiaCollector_CollectNvidiaDynamic(b *testing.B) {
	// Configure Manager with ONLY Nvidia enabled
	cfg := config.CollectorConfig{
		Nvidia:      true,
		NvidiaProcs: true, // Enable processes if that was the intent of the original test
	}
	cm := NewCollectorManager(cfg)
	// Ensure we clean up NVML connections
	defer cm.Close()

	// Check if Nvidia init actually succeeded, otherwise skip
	// Note: checking length because NewCollectorManager initializes the slice
	if len(cm.collectors) == 0 {
		b.Skip("Skipping: Nvidia collector could not be initialized")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}

func BenchmarkCollectVLLMDynamic(b *testing.B) {
	cfg := config.CollectorConfig{VLLM: true}
	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}

func BenchmarkCollectCPUDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/stat"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/stat not available")
	}

	cfg := config.CollectorConfig{CPU: true}
	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}

func BenchmarkCollectDiskDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/diskstats"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/diskstats not available")
	}

	cfg := config.CollectorConfig{Disk: true}
	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}

func BenchmarkCollectNetworkDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/net/dev"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/net/dev not available")
	}

	cfg := config.CollectorConfig{Network: true}
	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}

func BenchmarkCollectContainerDynamic(b *testing.B) {
	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		b.Skip("Skipping: /sys/fs/cgroup not available")
	}

	cfg := config.CollectorConfig{Container: true}
	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}

func BenchmarkCollectProcessesDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/1/stat"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc/[pid]/stat not available")
	}

	cfg := config.CollectorConfig{Processes: true}
	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: CollectorManager.CollectDynamicMetrics creates a fresh
		// DynamicMetrics struct internally, so manual slice resetting is
		// no longer required here.
		cm.CollectDynamicMetrics()
	}
}

// BenchmarkCollectAllDynamic benchmarks collecting ALL dynamic metrics simultaneously
func BenchmarkCollectAllDynamic(b *testing.B) {
	if _, err := os.Stat("/proc/meminfo"); os.IsNotExist(err) {
		b.Skip("Skipping: /proc not available")
	}

	// Enable everything (except Nvidia/VLLM if hardware isn't guaranteed present,
	// adjust as needed for your specific benchmark env)
	cfg := config.CollectorConfig{
		CPU:       true,
		Memory:    true,
		Disk:      true,
		Network:   true,
		Container: true,
		Processes: true,
		Nvidia:    true, // Uncomment if running on GPU node
		// VLLM:    true, // Uncomment if VLLM endpoint is mockable/available
	}

	cm := NewCollectorManager(cfg)
	defer cm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.CollectDynamicMetrics()
	}
}
