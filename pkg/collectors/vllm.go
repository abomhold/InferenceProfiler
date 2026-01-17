package collectors

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

// Collector collects vLLM metrics.
type Collector struct {
	client   *http.Client
	endpoint string
}

// New creates a new vLLM collector.
func New() *Collector {
	endpoint := os.Getenv(config.VLLMEnvVar)
	if endpoint == "" {
		endpoint = config.DefaultVLLMEndpoint
	}
	return &Collector{
		client:   &http.Client{Timeout: 500 * time.Millisecond},
		endpoint: endpoint,
	}
}

// Name returns the collector name.
func (c *Collector) Name() string {
	return "vLLM"
}

// Close releases any resources.
func (c *Collector) Close() error {
	return nil
}

// CollectStatic returns nil as vLLM collector has no static data.
func (c *Collector) CollectStatic() types.Record {
	return nil
}

// CollectDynamic collects dynamic vLLM metrics.
func (c *Collector) CollectDynamic() types.Record {
	d := &Dynamic{
		VLLMTimestamp: probing.GetTimestamp(),
	}

	resp, err := c.client.Get(c.endpoint)
	if err != nil {
		d.VLLMAvailable = false
		return d.ToRecord()
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		d.VLLMAvailable = false
		return d.ToRecord()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.VLLMAvailable = false
		return d.ToRecord()
	}

	d.VLLMAvailable = true
	c.parseMetrics(d, string(body))

	return d.ToRecord()
}

func (c *Collector) parseMetrics(d *Dynamic, body string) {
	histograms := &Histograms{}

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, labels, value, ok := parseMetricLine(line)
		if !ok {
			continue
		}

		name = strings.TrimPrefix(name, "vllm:")

		switch {
		case strings.HasSuffix(name, "_bucket"):
			baseName := strings.TrimSuffix(name, "_bucket")
			le := labels["le"]
			addHistogramBucket(histograms, baseName, le, value)

		case strings.HasSuffix(name, "_sum"):
			baseName := strings.TrimSuffix(name, "_sum")
			setHistogramSum(d, baseName, value)

		case strings.HasSuffix(name, "_count"):
			baseName := strings.TrimSuffix(name, "_count")
			setHistogramCount(d, baseName, value)

		case strings.HasSuffix(name, "_total"):
			baseName := strings.TrimSuffix(name, "_total")
			setGaugeOrCounter(d, baseName, value)

		default:
			setGaugeOrCounter(d, name, value)
		}
	}

	if hasHistogramData(histograms) {
		data, _ := json.Marshal(histograms)
		d.VLLMHistogramsJSON = string(data)
	}
}

func parseMetricLine(line string) (string, map[string]string, float64, bool) {
	labels := make(map[string]string)

	lastSpace := strings.LastIndex(line, " ")
	if lastSpace == -1 {
		return "", nil, 0, false
	}

	valueStr := strings.TrimSpace(line[lastSpace+1:])
	metricPart := strings.TrimSpace(line[:lastSpace])

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return "", nil, 0, false
	}

	labelStart := strings.Index(metricPart, "{")
	if labelStart == -1 {
		return metricPart, labels, value, true
	}

	name := metricPart[:labelStart]
	labelEnd := strings.LastIndex(metricPart, "}")
	if labelEnd == -1 || labelEnd <= labelStart {
		return "", nil, 0, false
	}

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
		val = strings.Trim(val, "\"")
		labels[key] = val
	}

	return name, labels, value, true
}

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

func addHistogramBucket(h *Histograms, name, le string, value float64) {
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

func setHistogramSum(d *Dynamic, name string, value float64) {
	value = sanitizeFloat(value)
	switch name {
	case "time_to_first_token_seconds":
		d.VLLMLatencyTtftSum = value
	case "e2e_request_latency_seconds":
		d.VLLMLatencyE2eSum = value
	case "request_queue_time_seconds":
		d.VLLMLatencyQueueSum = value
	case "request_inference_time_seconds":
		d.VLLMLatencyInferenceSum = value
	case "request_prefill_time_seconds":
		d.VLLMLatencyPrefillSum = value
	case "request_decode_time_seconds":
		d.VLLMLatencyDecodeSum = value
	}
}

func setHistogramCount(d *Dynamic, name string, value float64) {
	value = sanitizeFloat(value)
	switch name {
	case "time_to_first_token_seconds":
		d.VLLMLatencyTtftCount = value
	case "e2e_request_latency_seconds":
		d.VLLMLatencyE2eCount = value
	case "request_queue_time_seconds":
		d.VLLMLatencyQueueCount = value
	case "request_inference_time_seconds":
		d.VLLMLatencyInferenceCount = value
	case "request_prefill_time_seconds":
		d.VLLMLatencyPrefillCount = value
	case "request_decode_time_seconds":
		d.VLLMLatencyDecodeCount = value
	}
}

func setGaugeOrCounter(d *Dynamic, name string, value float64) {
	value = sanitizeFloat(value)
	switch name {
	case "num_requests_running":
		d.VLLMRequestsRunning = value
	case "num_requests_waiting":
		d.VLLMRequestsWaiting = value
	case "engine_sleep_state":
		d.VLLMEngineSleepState = value
	case "num_preemptions":
		d.VLLMPreemptionsTotal = value
	case "kv_cache_usage_perc":
		d.VLLMKvCacheUsagePercent = value
	case "prefix_cache_hits":
		d.VLLMPrefixCacheHits = value
	case "prefix_cache_queries":
		d.VLLMPrefixCacheQueries = value
	case "request_success":
		d.VLLMRequestsFinishedTotal = value
	case "corrupted_requests":
		d.VLLMRequestsCorruptedTotal = value
	case "prompt_tokens":
		d.VLLMTokensPromptTotal = value
	case "generation_tokens":
		d.VLLMTokensGenerationTotal = value
	}
}

func sanitizeFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

func hasHistogramData(h *Histograms) bool {
	return len(h.LatencyTtft) > 0 ||
		len(h.LatencyE2e) > 0 ||
		len(h.LatencyQueue) > 0 ||
		len(h.LatencyInference) > 0 ||
		len(h.LatencyPrefill) > 0 ||
		len(h.LatencyDecode) > 0 ||
		len(h.LatencyInterToken) > 0 ||
		len(h.TokensPerStep) > 0 ||
		len(h.ReqSizePromptTokens) > 0 ||
		len(h.ReqSizeGenerationTokens) > 0 ||
		len(h.ReqParamsMaxTokens) > 0 ||
		len(h.ReqParamsN) > 0
}
