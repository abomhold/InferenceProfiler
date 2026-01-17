package collecting

import (
	"InferenceProfiler/pkg/utils"
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type VLLMCollector struct {
	endpoint string
	client   *http.Client
}

func NewVLLMCollector() *VLLMCollector {
	endpoint := os.Getenv(utils.VLLMEnvVar)
	if endpoint == "" {
		endpoint = utils.DefaultVLLMEndpoint
	}

	log.Printf("vLLM collector: endpoint=%s", endpoint)

	return &VLLMCollector{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *VLLMCollector) Name() string { return "vLLM" }
func (c *VLLMCollector) Close() error { return nil }

func (c *VLLMCollector) CollectStatic(m *StaticMetrics) {}

func (c *VLLMCollector) CollectDynamic(m *DynamicMetrics) {
	ts := utils.GetTimestamp()

	resp, err := c.client.Get(c.endpoint)
	if err != nil {
		// vLLM endpoint not available - this is expected if vLLM isn't running
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	m.VLLMAvailable = true
	m.VLLMTimestamp = ts

	histograms := &VLLMHistograms{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		name := parts[0]
		value := utils.ParseFloat64(parts[1])

		switch {
		case name == "vllm:num_requests_running":
			m.VLLMRequestsRunning = value
		case name == "vllm:num_requests_waiting":
			m.VLLMRequestsWaiting = value
		case name == "vllm:engine_sleep_state":
			m.VLLMEngineSleepState = value
		case name == "vllm:num_preemptions_total":
			m.VLLMPreemptionsTotal = value
		case name == "vllm:gpu_cache_usage_perc":
			m.VLLMKvCacheUsagePercent = value
		case name == "vllm:prefix_cache_hit_rate":
			m.VLLMPrefixCacheHits = value
		case name == "vllm:prefix_cache_queries_total":
			m.VLLMPrefixCacheQueries = value
		case name == "vllm:request_success_total":
			m.VLLMRequestsFinishedTotal = value
		case name == "vllm:request_corrupted_total":
			m.VLLMRequestsCorruptedTotal = value
		case name == "vllm:prompt_tokens_total":
			m.VLLMTokensPromptTotal = value
		case name == "vllm:generation_tokens_total":
			m.VLLMTokensGenerationTotal = value
		case name == "vllm:time_to_first_token_seconds_sum":
			m.VLLMLatencyTtftSum = value
		case name == "vllm:time_to_first_token_seconds_count":
			m.VLLMLatencyTtftCount = value
		case name == "vllm:e2e_request_latency_seconds_sum":
			m.VLLMLatencyE2eSum = value
		case name == "vllm:e2e_request_latency_seconds_count":
			m.VLLMLatencyE2eCount = value
		case name == "vllm:request_queue_time_seconds_sum":
			m.VLLMLatencyQueueSum = value
		case name == "vllm:request_queue_time_seconds_count":
			m.VLLMLatencyQueueCount = value
		case name == "vllm:request_inference_time_seconds_sum":
			m.VLLMLatencyInferenceSum = value
		case name == "vllm:request_inference_time_seconds_count":
			m.VLLMLatencyInferenceCount = value
		case name == "vllm:request_prefill_time_seconds_sum":
			m.VLLMLatencyPrefillSum = value
		case name == "vllm:request_prefill_time_seconds_count":
			m.VLLMLatencyPrefillCount = value
		case name == "vllm:request_decode_time_seconds_sum":
			m.VLLMLatencyDecodeSum = value
		case name == "vllm:request_decode_time_seconds_count":
			m.VLLMLatencyDecodeCount = value
		default:
			parseHistogramBucket(name, value, histograms)
		}
	}

	if hasHistogramData(histograms) {
		data, _ := json.Marshal(histograms)
		m.VLLMHistogramsJSON = string(data)
	}
}

func parseHistogramBucket(name string, value float64, h *VLLMHistograms) {
	if !strings.Contains(name, "_bucket{") {
		return
	}

	start := strings.Index(name, "{le=\"")
	if start == -1 {
		return
	}
	end := strings.Index(name[start+5:], "\"}")
	if end == -1 {
		return
	}
	bucket := name[start+5 : start+5+end]
	metricName := name[:strings.Index(name, "_bucket")]

	var target *map[string]float64
	switch metricName {
	case "vllm:num_tokens_generated_per_step":
		target = &h.TokensPerStep
	case "vllm:time_to_first_token_seconds":
		target = &h.LatencyTtft
	case "vllm:e2e_request_latency_seconds":
		target = &h.LatencyE2e
	case "vllm:request_queue_time_seconds":
		target = &h.LatencyQueue
	case "vllm:request_inference_time_seconds":
		target = &h.LatencyInference
	case "vllm:request_prefill_time_seconds":
		target = &h.LatencyPrefill
	case "vllm:request_decode_time_seconds":
		target = &h.LatencyDecode
	case "vllm:inter_token_latency_seconds":
		target = &h.LatencyInterToken
	case "vllm:request_prompt_tokens":
		target = &h.ReqSizePromptTokens
	case "vllm:request_generation_tokens":
		target = &h.ReqSizeGenerationTokens
	case "vllm:request_max_tokens":
		target = &h.ReqParamsMaxTokens
	case "vllm:request_n":
		target = &h.ReqParamsN
	}

	if target != nil {
		if *target == nil {
			*target = make(map[string]float64)
		}
		(*target)[bucket] = value
	}
}

func hasHistogramData(h *VLLMHistograms) bool {
	return len(h.TokensPerStep) > 0 || len(h.LatencyTtft) > 0 ||
		len(h.LatencyE2e) > 0 || len(h.LatencyQueue) > 0 ||
		len(h.LatencyInference) > 0 || len(h.LatencyPrefill) > 0 ||
		len(h.LatencyDecode) > 0 || len(h.LatencyInterToken) > 0
}
