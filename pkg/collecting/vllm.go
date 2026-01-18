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
	resp, err := c.client.Get(c.endpoint)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	m.VLLMAvailable = true
	m.VLLMTimestamp = utils.GetTimestamp()
	histograms := &VLLMHistograms{}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fullKey, valStr, found := strings.Cut(line, " ")
		if !found {
			continue
		}

		key := strings.TrimPrefix(fullKey, vllmBucketEnd)
		value := utils.ParseFloat64(valStr)

		if name, labels, hasLabels := strings.Cut(key, "{"); hasLabels {
			if strings.HasSuffix(name, vllmBucketSuffix) {
				parseBucket(name, labels, value, histograms)
			}
			continue
		}

		switch key {
		case "num_requests_running":
			m.VLLMRequestsRunning = value
		case "num_requests_waiting":
			m.VLLMRequestsWaiting = value
		case "engine_sleep_state":
			m.VLLMEngineSleepState = value
		case "num_preemptions_total":
			m.VLLMPreemptionsTotal = value
		case "gpu_cache_usage_perc":
			m.VLLMKvCacheUsagePercent = value
		case "prefix_cache_hit_rate":
			m.VLLMPrefixCacheHits = value
		case "prefix_cache_queries_total":
			m.VLLMPrefixCacheQueries = value
		case "request_success_total":
			m.VLLMRequestsFinishedTotal = value
		case "request_corrupted_total":
			m.VLLMRequestsCorruptedTotal = value
		case "prompt_tokens_total":
			m.VLLMTokensPromptTotal = value
		case "generation_tokens_total":
			m.VLLMTokensGenerationTotal = value
		case "time_to_first_token_seconds_sum":
			m.VLLMLatencyTtftSum = value
		case "time_to_first_token_seconds_count":
			m.VLLMLatencyTtftCount = value
		case "e2e_request_latency_seconds_sum":
			m.VLLMLatencyE2eSum = value
		case "e2e_request_latency_seconds_count":
			m.VLLMLatencyE2eCount = value
		case "request_queue_time_seconds_sum":
			m.VLLMLatencyQueueSum = value
		case "request_queue_time_seconds_count":
			m.VLLMLatencyQueueCount = value
		case "request_inference_time_seconds_sum":
			m.VLLMLatencyInferenceSum = value
		case "request_inference_time_seconds_count":
			m.VLLMLatencyInferenceCount = value
		case "request_prefill_time_seconds_sum":
			m.VLLMLatencyPrefillSum = value
		case "request_prefill_time_seconds_count":
			m.VLLMLatencyPrefillCount = value
		case "request_decode_time_seconds_sum":
			m.VLLMLatencyDecodeSum = value
		case "request_decode_time_seconds_count":
			m.VLLMLatencyDecodeCount = value
		}
	}

	if hasHistogramData(histograms) {
		if data, err := json.Marshal(histograms); err == nil {
			m.VLLMHistogramsJSON = string(data)
		}
	}
}

func parseBucket(name, labels string, value float64, h *VLLMHistograms) {
	_, afterLe, found := strings.Cut(labels, "le=\"")
	if !found {
		return
	}
	bucket, _, found := strings.Cut(afterLe, "\"")
	if !found {
		return
	}

	baseName := strings.TrimSuffix(name, vllmBucketSuffix)
	var target *map[string]float64

	switch baseName {
	case "num_tokens_generated_per_step":
		target = &h.TokensPerStep
	case "time_to_first_token_seconds":
		target = &h.LatencyTtft
	case "e2e_request_latency_seconds":
		target = &h.LatencyE2e
	case "request_queue_time_seconds":
		target = &h.LatencyQueue
	case "request_inference_time_seconds":
		target = &h.LatencyInference
	case "request_prefill_time_seconds":
		target = &h.LatencyPrefill
	case "request_decode_time_seconds":
		target = &h.LatencyDecode
	case "inter_token_latency_seconds":
		target = &h.LatencyInterToken
	case "request_prompt_tokens":
		target = &h.ReqSizePromptTokens
	case "request_generation_tokens":
		target = &h.ReqSizeGenerationTokens
	case "request_max_tokens":
		target = &h.ReqParamsMaxTokens
	case "request_n":
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
	return h.TokensPerStep != nil || h.LatencyTtft != nil || h.LatencyE2e != nil ||
		h.LatencyQueue != nil || h.LatencyInference != nil || h.LatencyPrefill != nil ||
		h.LatencyDecode != nil || h.LatencyInterToken != nil
}
