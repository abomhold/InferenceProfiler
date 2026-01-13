package collectors

import (
	"bufio"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var vllmMetricsURL = getEnv("VLLM_METRICS_URL", "http://localhost:8000/metrics")

// CollectVLLMDynamic scrapes and parses vLLM Prometheus metrics
func CollectVLLMDynamic(m *DynamicMetrics) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	scrapeTime := GetTimestamp()

	resp, err := client.Get(vllmMetricsURL)
	if err != nil {
		m.VLLMAvailable = false
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.VLLMAvailable = false
		return
	}

	m.VLLMAvailable = true
	m.VLLMTimestamp = scrapeTime

	histograms := &VLLMHistograms{}

	// Parse Prometheus text format directly
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, labels, value, ok := parseMetricLine(line)
		if !ok {
			continue
		}

		// Strip "vllm:" prefix
		name = strings.TrimPrefix(name, "vllm:")

		// Route to appropriate handler based on suffix
		switch {
		case strings.HasSuffix(name, "_bucket"):
			baseName := strings.TrimSuffix(name, "_bucket")
			le := labels["le"]
			addHistogramBucket(histograms, baseName, le, value)

		case strings.HasSuffix(name, "_sum"):
			baseName := strings.TrimSuffix(name, "_sum")
			setHistogramSum(m, baseName, value)

		case strings.HasSuffix(name, "_count"):
			baseName := strings.TrimSuffix(name, "_count")
			setHistogramCount(m, baseName, value)

		case strings.HasSuffix(name, "_total"):
			// Counter with _total suffix
			baseName := strings.TrimSuffix(name, "_total")
			setGaugeOrCounter(m, baseName, value)

		default:
			setGaugeOrCounter(m, name, value)
		}
	}

	if data, err := json.Marshal(histograms); err == nil && string(data) != "{}" {
		m.VLLMHistogramsJSON = string(data)
	}
}

// parseMetricLine parses a Prometheus metric line
// Format: metric_name{label="value",label2="value2"} 123.45
// Returns: name, labels map, value, success
func parseMetricLine(line string) (string, map[string]string, float64, bool) {
	labels := make(map[string]string)

	// Find value (last space-separated field)
	lastSpace := strings.LastIndex(line, " ")
	if lastSpace == -1 {
		return "", nil, 0, false
	}

	valueStr := strings.TrimSpace(line[lastSpace+1:])
	metricPart := strings.TrimSpace(line[:lastSpace])

	// Parse value
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return "", nil, 0, false
	}

	// Check for labels
	labelStart := strings.Index(metricPart, "{")
	if labelStart == -1 {
		// No labels
		return metricPart, labels, value, true
	}

	name := metricPart[:labelStart]
	labelEnd := strings.LastIndex(metricPart, "}")
	if labelEnd == -1 || labelEnd <= labelStart {
		return "", nil, 0, false
	}

	// Parse labels
	labelStr := metricPart[labelStart+1 : labelEnd]
	for _, pair := range splitLabels(labelStr) {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		eqIdx := strings.Index(pair, "=")
		if eqIdx == -1 {
			continue
		}
		key := strings.TrimSpace(pair[:eqIdx])
		val := strings.TrimSpace(pair[eqIdx+1:])
		// Remove quotes
		val = strings.Trim(val, "\"")
		labels[key] = val
	}

	return name, labels, value, true
}

// splitLabels splits label string handling quoted commas
func splitLabels(s string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for _, c := range s {
		switch c {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(c)
		case ',':
			if inQuotes {
				current.WriteRune(c)
			} else {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(c)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

func addHistogramBucket(h *VLLMHistograms, name, le string, value float64) {
	// Normalize le value
	if le == "+Inf" {
		le = "inf"
	}

	switch name {
	case "time_to_first_token_seconds":
		if h.LatencyTtft == nil {
			h.LatencyTtft = make(map[string]float64)
		}
		h.LatencyTtft[le] = value
	case "e2e_request_latency_seconds":
		if h.LatencyE2e == nil {
			h.LatencyE2e = make(map[string]float64)
		}
		h.LatencyE2e[le] = value
	case "request_queue_time_seconds":
		if h.LatencyQueue == nil {
			h.LatencyQueue = make(map[string]float64)
		}
		h.LatencyQueue[le] = value
	case "request_inference_time_seconds":
		if h.LatencyInference == nil {
			h.LatencyInference = make(map[string]float64)
		}
		h.LatencyInference[le] = value
	case "request_prefill_time_seconds":
		if h.LatencyPrefill == nil {
			h.LatencyPrefill = make(map[string]float64)
		}
		h.LatencyPrefill[le] = value
	case "request_decode_time_seconds":
		if h.LatencyDecode == nil {
			h.LatencyDecode = make(map[string]float64)
		}
		h.LatencyDecode[le] = value
	case "inter_token_latency_seconds":
		if h.LatencyInterToken == nil {
			h.LatencyInterToken = make(map[string]float64)
		}
		h.LatencyInterToken[le] = value
	case "request_prompt_tokens":
		if h.ReqSizePromptTokens == nil {
			h.ReqSizePromptTokens = make(map[string]float64)
		}
		h.ReqSizePromptTokens[le] = value
	case "request_generation_tokens":
		if h.ReqSizeGenerationTokens == nil {
			h.ReqSizeGenerationTokens = make(map[string]float64)
		}
		h.ReqSizeGenerationTokens[le] = value
	case "iteration_tokens":
		if h.TokensPerStep == nil {
			h.TokensPerStep = make(map[string]float64)
		}
		h.TokensPerStep[le] = value
	case "request_params_max_tokens":
		if h.ReqParamsMaxTokens == nil {
			h.ReqParamsMaxTokens = make(map[string]float64)
		}
		h.ReqParamsMaxTokens[le] = value
	case "request_params_n":
		if h.ReqParamsN == nil {
			h.ReqParamsN = make(map[string]float64)
		}
		h.ReqParamsN[le] = value
	}
}

func setHistogramSum(m *DynamicMetrics, name string, value float64) {
	value = sanitizeFloat(value)
	switch name {
	case "time_to_first_token_seconds":
		m.VLLMLatencyTtftSum = value
	case "e2e_request_latency_seconds":
		m.VLLMLatencyE2eSum = value
	case "request_queue_time_seconds":
		m.VLLMLatencyQueueSum = value
	case "request_inference_time_seconds":
		m.VLLMLatencyInferenceSum = value
	case "request_prefill_time_seconds":
		m.VLLMLatencyPrefillSum = value
	case "request_decode_time_seconds":
		m.VLLMLatencyDecodeSum = value
	}
}

func setHistogramCount(m *DynamicMetrics, name string, value float64) {
	value = sanitizeFloat(value)
	switch name {
	case "time_to_first_token_seconds":
		m.VLLMLatencyTtftCount = value
	case "e2e_request_latency_seconds":
		m.VLLMLatencyE2eCount = value
	case "request_queue_time_seconds":
		m.VLLMLatencyQueueCount = value
	case "request_inference_time_seconds":
		m.VLLMLatencyInferenceCount = value
	case "request_prefill_time_seconds":
		m.VLLMLatencyPrefillCount = value
	case "request_decode_time_seconds":
		m.VLLMLatencyDecodeCount = value
	}
}

func setGaugeOrCounter(m *DynamicMetrics, name string, value float64) {
	value = sanitizeFloat(value)
	switch name {
	case "num_requests_running":
		m.VLLMRequestsRunning = value
	case "num_requests_waiting":
		m.VLLMRequestsWaiting = value
	case "engine_sleep_state":
		m.VLLMEngineSleepState = value
	case "num_preemptions":
		m.VLLMPreemptionsTotal = value
	case "kv_cache_usage_perc":
		m.VLLMKvCacheUsagePercent = value
	case "prefix_cache_hits":
		m.VLLMPrefixCacheHits = value
	case "prefix_cache_queries":
		m.VLLMPrefixCacheQueries = value
	case "request_success":
		m.VLLMRequestsFinishedTotal = value
	case "corrupted_requests":
		m.VLLMRequestsCorruptedTotal = value
	case "prompt_tokens":
		m.VLLMTokensPromptTotal = value
	case "generation_tokens":
		m.VLLMTokensGenerationTotal = value
	}
}

func sanitizeFloat(v float64) float64 {
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
