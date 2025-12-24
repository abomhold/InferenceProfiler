package src

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// VLLMMetrics contains vLLM inference engine measurements.
// Parsed from Prometheus text format at /metrics endpoint.
type VLLMMetrics struct {
	Timestamp int64 `json:"timestamp,omitempty"`

	// System state
	RequestsRunning  float64 `json:"system_requests_running,omitempty"`
	RequestsWaiting  float64 `json:"system_requests_waiting,omitempty"`
	EngineSleepState float64 `json:"system_engine_sleep_state,omitempty"`
	PreemptionsTotal float64 `json:"system_preemptions_total,omitempty"`

	// Cache utilization
	KVCacheUsagePercent    float64 `json:"cache_kv_usage_percent,omitempty"`
	PrefixCacheHits        float64 `json:"cache_prefix_hits,omitempty"`
	PrefixCacheQueries     float64 `json:"cache_prefix_queries,omitempty"`
	MultimodalCacheHits    float64 `json:"cache_multimodal_hits,omitempty"`
	MultimodalCacheQueries float64 `json:"cache_multimodal_queries,omitempty"`

	// Throughput counters
	RequestsFinished  float64 `json:"requests_finished_total,omitempty"`
	RequestsCorrupted float64 `json:"requests_corrupted_total,omitempty"`
	PromptTokens      float64 `json:"tokens_prompt_total,omitempty"`
	GenerationTokens  float64 `json:"tokens_generation_total,omitempty"`

	// Latency sums (for computing averages with counts)
	LatencyTTFTSum       float64 `json:"latency_ttft_s_sum,omitempty"`
	LatencyE2ESum        float64 `json:"latency_e2e_s_sum,omitempty"`
	LatencyQueueSum      float64 `json:"latency_queue_s_sum,omitempty"`
	LatencyInferenceSum  float64 `json:"latency_inference_s_sum,omitempty"`
	LatencyPrefillSum    float64 `json:"latency_prefill_s_sum,omitempty"`
	LatencyDecodeSum     float64 `json:"latency_decode_s_sum,omitempty"`
	LatencyInterTokenSum float64 `json:"latency_inter_token_s_sum,omitempty"`

	// Histograms stored as bucket->cumulative_count
	Histograms map[string]map[string]float64 `json:"histograms,omitempty"`

	// Config info (extracted from *_info metrics)
	Config map[string]interface{} `json:"config,omitempty"`
}

// vLLM metric name aliases for cleaner JSON output.
var vllmAliases = map[string]string{
	"num_requests_running":           "system_requests_running",
	"num_requests_waiting":           "system_requests_waiting",
	"engine_sleep_state":             "system_engine_sleep_state",
	"num_preemptions":                "system_preemptions_total",
	"kv_cache_usage_perc":            "cache_kv_usage_percent",
	"prefix_cache_hits":              "cache_prefix_hits",
	"prefix_cache_queries":           "cache_prefix_queries",
	"mm_cache_hits":                  "cache_multimodal_hits",
	"mm_cache_queries":               "cache_multimodal_queries",
	"request_success":                "requests_finished_total",
	"corrupted_requests":             "requests_corrupted_total",
	"prompt_tokens":                  "tokens_prompt_total",
	"generation_tokens":              "tokens_generation_total",
	"time_to_first_token_seconds":    "latency_ttft_s",
	"e2e_request_latency_seconds":    "latency_e2e_s",
	"request_queue_time_seconds":     "latency_queue_s",
	"request_inference_time_seconds": "latency_inference_s",
	"request_prefill_time_seconds":   "latency_prefill_s",
	"request_decode_time_seconds":    "latency_decode_s",
	"inter_token_latency_seconds":    "latency_inter_token_s",
}

// Labels to ignore (reduce noise).
var ignoredLabels = map[string]bool{
	"model_name": true, "model": true, "engine_id": true,
	"engine": true, "handler": true, "method": true,
}

// Prometheus line regex: metric{labels} value
var promLineRE = regexp.MustCompile(`^([a-zA-Z0-9_:]+)(?:\{(.+)\})?\s+([0-9\.eE\+\-]+|nan|inf|NaN|Inf)$`)

// CollectVLLM scrapes metrics from vLLM's Prometheus endpoint.
func CollectVLLM() VLLMMetrics {
	url := os.Getenv("VLLM_METRICS_URL")
	if url == "" {
		url = "http://localhost:8000/metrics"
	}

	var m VLLMMetrics
	ts := time.Now().UnixMilli()

	client := http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		return m
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return m
	}

	m = parsePrometheus(string(body))
	if len(m.Histograms) > 0 || m.RequestsRunning > 0 || m.RequestsWaiting > 0 {
		m.Timestamp = ts
	}
	return m
}

// parsePrometheus parses Prometheus text format into VLLMMetrics.
func parsePrometheus(text string) VLLMMetrics {
	m := VLLMMetrics{
		Histograms: make(map[string]map[string]float64),
		Config:     make(map[string]interface{}),
	}

	// Temporary storage for histogram buckets
	histoBuckets := make(map[string][]bucketEntry)

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		match := promLineRE.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		name, labelStr, valStr := match[1], match[2], match[3]
		val := parsePromValue(valStr)

		// Parse and filter labels
		labels := parseLabels(labelStr)

		// Strip vllm: prefix and handle suffixes
		name = strings.TrimPrefix(name, "vllm:")
		isBucket := strings.HasSuffix(name, "_bucket")
		isSum := strings.HasSuffix(name, "_sum")
		isCount := strings.HasSuffix(name, "_count")
		isInfo := strings.HasSuffix(name, "_info")

		baseName := name
		if isBucket {
			baseName = name[:len(name)-7]
		} else if isSum {
			baseName = name[:len(name)-4]
		} else if isCount {
			baseName = name[:len(name)-6]
		}

		// Get alias if exists
		cleanName := baseName
		if alias, ok := vllmAliases[baseName]; ok {
			cleanName = alias
		}

		// Handle _info metrics -> extract to config
		if isInfo && len(labels) > 0 {
			for k, v := range labels {
				m.Config[cleanName+"_"+k] = tryParseNumber(v)
			}
			continue
		}

		// Handle histogram buckets
		if isBucket {
			le, ok := labels["le"]
			if !ok {
				continue
			}
			delete(labels, "le")

			// Build histogram key
			histoKey := buildKey(cleanName, labels)
			histoBuckets[histoKey] = append(histoBuckets[histoKey], bucketEntry{
				le:    parseLe(le),
				count: val,
			})
			continue
		}

		// Handle _sum and _count
		if isSum {
			assignSum(&m, cleanName, val)
			continue
		}

		if isCount {
			// Store count metrics if needed
			continue
		}

		// Handle standard metrics
		assignMetric(&m, cleanName, val, labels)
	}

	// Convert histogram buckets to sorted maps
	for key, buckets := range histoBuckets {
		sort.Slice(buckets, func(i, j int) bool {
			return buckets[i].le < buckets[j].le
		})

		histo := make(map[string]float64)
		for _, b := range buckets {
			leKey := "inf"
			if b.le != posInf {
				leKey = strconv.FormatFloat(b.le, 'f', -1, 64)
			}
			histo[leKey] = b.count
		}
		m.Histograms[key+"_histogram"] = histo
	}

	return m
}

type bucketEntry struct {
	le    float64
	count float64
}

const posInf = 1e308

func parsePromValue(s string) float64 {
	s = strings.ToLower(s)
	if s == "nan" {
		return 0
	}
	if s == "inf" || s == "+inf" {
		return posInf
	}
	if s == "-inf" {
		return -posInf
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseLe(s string) float64 {
	s = strings.ToLower(s)
	if s == "+inf" || s == "inf" {
		return posInf
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseLabels(s string) map[string]string {
	m := make(map[string]string)
	if s == "" {
		return m
	}

	for _, part := range strings.Split(s, ",") {
		idx := strings.Index(part, "=")
		if idx < 0 {
			continue
		}
		k := strings.TrimSpace(part[:idx])
		v := strings.Trim(strings.TrimSpace(part[idx+1:]), `"`)

		if !ignoredLabels[k] {
			m[k] = v
		}
	}
	return m
}

func buildKey(base string, labels map[string]string) string {
	if len(labels) == 0 {
		return base
	}

	var parts []string
	for k, v := range labels {
		parts = append(parts, k+"_"+v)
	}
	sort.Strings(parts)
	return base + "_" + strings.Join(parts, "_")
}

func tryParseNumber(s string) interface{} {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func assignSum(m *VLLMMetrics, name string, val float64) {
	switch name {
	case "latency_ttft_s":
		m.LatencyTTFTSum = val
	case "latency_e2e_s":
		m.LatencyE2ESum = val
	case "latency_queue_s":
		m.LatencyQueueSum = val
	case "latency_inference_s":
		m.LatencyInferenceSum = val
	case "latency_prefill_s":
		m.LatencyPrefillSum = val
	case "latency_decode_s":
		m.LatencyDecodeSum = vald
	case "latency_inter_token_s":
		m.LatencyInterTokenSum = val
	}
}

func assignMetric(m *VLLMMetrics, name string, val float64, labels map[string]string) {
	switch name {
	case "system_requests_running":
		m.RequestsRunning = val
	case "system_requests_waiting":
		m.RequestsWaiting = val
	case "system_engine_sleep_state":
		m.EngineSleepState = val
	case "system_preemptions_total":
		m.PreemptionsTotal = val
	case "cache_kv_usage_percent":
		m.KVCacheUsagePercent = val
	case "cache_prefix_hits":
		m.PrefixCacheHits = val
	case "cache_prefix_queries":
		m.PrefixCacheQueries = val
	case "cache_multimodal_hits":
		m.MultimodalCacheHits = val
	case "cache_multimodal_queries":
		m.MultimodalCacheQueries = val
	case "requests_finished_total":
		m.RequestsFinished = val
	case "requests_corrupted_total":
		m.RequestsCorrupted = val
	case "tokens_prompt_total":
		m.PromptTokens = val
	case "tokens_generation_total":
		m.GenerationTokens = val
	}
}
