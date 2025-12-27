package collectors

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

var vllmMetricsURL = getEnv("VLLM_METRICS_URL", "http://localhost:8000/metrics")

// Metric name mappings from Prometheus names to our struct fields
var metricMappings = map[string]string{
	"num_requests_running":           "RequestsRunning",
	"num_requests_waiting":           "RequestsWaiting",
	"engine_sleep_state":             "EngineSleepState",
	"num_preemptions":                "PreemptionsTotal",
	"kv_cache_usage_perc":            "KvCacheUsagePercent",
	"prefix_cache_hits":              "PrefixCacheHits",
	"prefix_cache_queries":           "PrefixCacheQueries",
	"mm_cache_hits":                  "MultimodalCacheHits",
	"mm_cache_queries":               "MultimodalCacheQueries",
	"request_success":                "RequestsFinishedTotal",
	"corrupted_requests":             "RequestsCorruptedTotal",
	"prompt_tokens":                  "TokensPromptTotal",
	"generation_tokens":              "TokensGenerationTotal",
	"time_to_first_token_seconds":    "LatencyTtft",
	"e2e_request_latency_seconds":    "LatencyE2e",
	"request_queue_time_seconds":     "LatencyQueue",
	"request_inference_time_seconds": "LatencyInference",
	"request_prefill_time_seconds":   "LatencyPrefill",
	"request_decode_time_seconds":    "LatencyDecode",
	"inter_token_latency_seconds":    "LatencyInterToken",
	"request_prompt_tokens":          "ReqSizePromptTokens",
	"request_generation_tokens":      "ReqSizeGenerationTokens",
	"iteration_tokens_total":         "TokensPerStep",
	"request_params_max_tokens":      "ReqParamsMaxTokens",
	"request_params_n":               "ReqParamsN",
}

// CollectVLLMDynamic populates vLLM metrics
func CollectVLLMDynamic(m *DynamicMetrics) {
	scrapeTime := GetTimestamp()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(vllmMetricsURL)
	if err != nil {
		m.VLLMAvailable = false
		return
	}
	defer resp.Body.Close()

	families, err := (&expfmt.TextParser{}).TextToMetricFamilies(resp.Body)
	if err != nil {
		m.VLLMAvailable = false
		return
	}

	m.VLLMAvailable = true
	m.VLLMTimestamp = scrapeTime

	histograms := &VLLMHistograms{}

	for name, mf := range families {
		baseName := strings.TrimPrefix(name, "vllm:")
		mapping, ok := metricMappings[baseName]
		if !ok {
			continue
		}

		for _, metric := range mf.GetMetric() {
			// Handle histogram metrics
			if h := metric.GetHistogram(); h != nil {
				sum := sanitize(h.GetSampleSum())
				count := float64(h.GetSampleCount())
				buckets := extractBuckets(h)

				setHistogramValues(m, histograms, mapping, sum, count, buckets)
				continue
			}

			// Handle summary metrics
			if s := metric.GetSummary(); s != nil {
				sum := sanitize(s.GetSampleSum())
				count := float64(s.GetSampleCount())
				setSummaryValues(m, mapping, sum, count)
				continue
			}

			// Handle gauge/counter metrics
			val := getValue(metric)
			setGaugeValue(m, mapping, sanitize(val))
		}
	}

	// Serialize histograms to JSON
	if data, err := json.Marshal(histograms); err == nil && string(data) != "{}" {
		m.VLLMHistogramsJSON = string(data)
	}
}

func extractBuckets(h *dto.Histogram) map[string]float64 {
	buckets := make(map[string]float64)
	for _, b := range h.GetBucket() {
		le := b.GetUpperBound()
		key := "inf"
		if !math.IsInf(le, 1) {
			key = fmt.Sprintf("%g", le)
		}
		buckets[key] = float64(b.GetCumulativeCount())
	}
	return buckets
}

func setHistogramValues(m *DynamicMetrics, h *VLLMHistograms, mapping string, sum, count float64, buckets map[string]float64) {
	switch mapping {
	case "LatencyTtft":
		m.VLLMLatencyTtftSum = sum
		m.VLLMLatencyTtftCount = count
		h.LatencyTtft = buckets
	case "LatencyE2e":
		m.VLLMLatencyE2eSum = sum
		m.VLLMLatencyE2eCount = count
		h.LatencyE2e = buckets
	case "LatencyQueue":
		m.VLLMLatencyQueueSum = sum
		m.VLLMLatencyQueueCount = count
		h.LatencyQueue = buckets
	case "LatencyInference":
		m.VLLMLatencyInferenceSum = sum
		m.VLLMLatencyInferenceCount = count
		h.LatencyInference = buckets
	case "LatencyPrefill":
		m.VLLMLatencyPrefillSum = sum
		m.VLLMLatencyPrefillCount = count
		h.LatencyPrefill = buckets
	case "LatencyDecode":
		m.VLLMLatencyDecodeSum = sum
		m.VLLMLatencyDecodeCount = count
		h.LatencyDecode = buckets
	case "LatencyInterToken":
		// No flattened fields for inter-token, just histogram
		h.LatencyInterToken = buckets
	case "ReqSizePromptTokens":
		h.ReqSizePromptTokens = buckets
	case "ReqSizeGenerationTokens":
		h.ReqSizeGenerationTokens = buckets
	case "TokensPerStep":
		h.TokensPerStep = buckets
	case "ReqParamsMaxTokens":
		h.ReqParamsMaxTokens = buckets
	case "ReqParamsN":
		h.ReqParamsN = buckets
	}
}

func setSummaryValues(m *DynamicMetrics, mapping string, sum, count float64) {
	switch mapping {
	case "LatencyTtft":
		m.VLLMLatencyTtftSum = sum
		m.VLLMLatencyTtftCount = count
	case "LatencyE2e":
		m.VLLMLatencyE2eSum = sum
		m.VLLMLatencyE2eCount = count
	case "LatencyQueue":
		m.VLLMLatencyQueueSum = sum
		m.VLLMLatencyQueueCount = count
	case "LatencyInference":
		m.VLLMLatencyInferenceSum = sum
		m.VLLMLatencyInferenceCount = count
	case "LatencyPrefill":
		m.VLLMLatencyPrefillSum = sum
		m.VLLMLatencyPrefillCount = count
	case "LatencyDecode":
		m.VLLMLatencyDecodeSum = sum
		m.VLLMLatencyDecodeCount = count
	}
}

func setGaugeValue(m *DynamicMetrics, mapping string, val float64) {
	switch mapping {
	case "RequestsRunning":
		m.VLLMRequestsRunning = val
	case "RequestsWaiting":
		m.VLLMRequestsWaiting = val
	case "EngineSleepState":
		m.VLLMEngineSleepState = val
	case "PreemptionsTotal":
		m.VLLMPreemptionsTotal = val
	case "KvCacheUsagePercent":
		m.VLLMKvCacheUsagePercent = val
	case "PrefixCacheHits":
		m.VLLMPrefixCacheHits = val
	case "PrefixCacheQueries":
		m.VLLMPrefixCacheQueries = val
	case "RequestsFinishedTotal":
		m.VLLMRequestsFinishedTotal = val
	case "RequestsCorruptedTotal":
		m.VLLMRequestsCorruptedTotal = val
	case "TokensPromptTotal":
		m.VLLMTokensPromptTotal = val
	case "TokensGenerationTotal":
		m.VLLMTokensGenerationTotal = val
	}
}

func getValue(m *dto.Metric) float64 {
	if c := m.GetCounter(); c != nil {
		return c.GetValue()
	}
	if g := m.GetGauge(); g != nil {
		return g.GetValue()
	}
	if u := m.GetUntyped(); u != nil {
		return u.GetValue()
	}
	return 0
}

func sanitize(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
