package collectors

import (
	"github.com/google/uuid"
)

// MetricValue represents a single metric with its value and timestamp
type MetricValue struct {
	Value interface{} `json:"value"`
	Time  int64       `json:"time,omitempty"`
}

func NewMetric(value interface{}) MetricValue {
	return MetricValue{
		Value: value,
	}
}

// NewMetricWithTime creates a MetricValue with specified timestamp
func NewMetricWithTime(value interface{}, ts int64) MetricValue {
	return MetricValue{
		Value: value,
		Time:  ts,
	}
}

// DynamicMetrics is a flat map of all dynamic metrics collected at runtime.
// Keys use prefixes to indicate source:
//   - v*: VM-level metrics (CPU, memory, disk, network)
//   - c*: Container-level metrics
//   - p*: Process-level metrics (when enabled)
//   - nvidia*: NVIDIA GPU metrics (nvidia_0_*, nvidia_1_*, etc. for multi-GPU)
//   - vllm*: vLLM inference metrics
type DynamicMetrics map[string]MetricValue

// StaticInfo is a flat map of all static system information collected once at startup.
// Keys use prefixes to indicate source:
//   - uuid: Session UUID
//   - vId: VM/instance identifier
//   - v*: VM-level static info (hostname, CPU, memory, kernel)
//   - nvidia*: NVIDIA driver/CUDA versions
//   - nvidia_0_*, nvidia_1_*: Per-GPU static info
type StaticInfo map[string]interface{}

// NewDynamicMetrics creates a new DynamicMetrics map with timestamp
func NewDynamicMetrics() DynamicMetrics {
	m := make(DynamicMetrics)
	m["timestamp"] = NewMetric(GetTimestamp())
	return m
}

// NewStaticInfo creates a new StaticInfo map with session UUID
func NewStaticInfo(sessionUUID uuid.UUID) StaticInfo {
	s := make(StaticInfo)
	s["uuid"] = sessionUUID.String()
	return s
}

// Merge adds all metrics from src into the DynamicMetrics map
func (d DynamicMetrics) Merge(src map[string]MetricValue) {
	for k, v := range src {
		d[k] = v
	}
}

// MergeWithPrefix adds all metrics from src with a prefix
func (d DynamicMetrics) MergeWithPrefix(prefix string, src map[string]MetricValue) {
	for k, v := range src {
		d[prefix+k] = v
	}
}

// MergeStatic adds all values from src into the StaticInfo map
func (s StaticInfo) Merge(src map[string]interface{}) {
	for k, v := range src {
		s[k] = v
	}
}

// MergeStaticWithPrefix adds all values from src with a prefix
func (s StaticInfo) MergeWithPrefix(prefix string, src map[string]interface{}) {
	for k, v := range src {
		s[prefix+k] = v
	}
}
