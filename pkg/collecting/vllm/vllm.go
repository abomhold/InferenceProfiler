package vllm

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type vllmDynamic struct {
	Available              bool             `json:"Available"`
	NumRequestsRunning     base.MetricFloat `json:"NumRequestsRunning"`
	NumRequestsWaiting     base.MetricFloat `json:"NumRequestsWaiting"`
	KvCacheUsagePercent    base.MetricFloat `json:"KvCacheUsagePercent"`
	NumPreemptionsTotal    base.MetricFloat `json:"NumPreemptionsTotal"`
	PrefixCacheHits        base.MetricFloat `json:"PrefixCacheHits"`
	PrefixCacheQueries     base.MetricFloat `json:"PrefixCacheQueries"`
	TtftHist               base.MetricStr   `json:"TtftHist"`
	E2eLatencyHist         base.MetricStr   `json:"E2eLatencyHist"`
	QueueTimeHist          base.MetricStr   `json:"QueueTimeHist"`
	InferenceTimeHist      base.MetricStr   `json:"InferenceTimeHist"`
	PrefillTimeHist        base.MetricStr   `json:"PrefillTimeHist"`
	DecodeTimeHist         base.MetricStr   `json:"DecodeTimeHist"`
	InterTokenLatencyHist  base.MetricStr   `json:"InterTokenLatencyHist"`
	PromptTokensHist       base.MetricStr   `json:"PromptTokensHist"`
	GenerationTokensHist   base.MetricStr   `json:"GenerationTokensHist"`
	TimePerOutputTokenHist base.MetricStr   `json:"TimePerOutputTokenHist"`
}

func parseVllm(r io.Reader, collectHistograms bool, m *vllmDynamic) {
	m.Available = false
	foundData := false

	floatMetrics := map[string]*base.MetricFloat{
		"vllm_num_requests_running":       &m.NumRequestsRunning,
		"vllm_num_requests_waiting":       &m.NumRequestsWaiting,
		"vllm_kv_cache_usage_perc":        &m.KvCacheUsagePercent,
		"vllm_num_preemptions_total":      &m.NumPreemptionsTotal,
		"vllm_prefix_cache_hits_total":    &m.PrefixCacheHits,
		"vllm_prefix_cache_queries_total": &m.PrefixCacheQueries,
	}

	histMetrics := map[string]*base.MetricStr{
		"vllm_time_to_first_token_seconds":           &m.TtftHist,
		"vllm_e2e_request_latency_seconds":           &m.E2eLatencyHist,
		"vllm_request_queue_time_seconds":            &m.QueueTimeHist,
		"vllm_request_inference_time_seconds":        &m.InferenceTimeHist,
		"vllm_request_prefill_time_seconds":          &m.PrefillTimeHist,
		"vllm_request_decode_time_seconds":           &m.DecodeTimeHist,
		"vllm_inter_token_latency_seconds":           &m.InterTokenLatencyHist,
		"vllm_request_prompt_tokens":                 &m.PromptTokensHist,
		"vllm_request_generation_tokens":             &m.GenerationTokensHist,
		"vllm_request_time_per_output_token_seconds": &m.TimePerOutputTokenHist,
	}

	buckets := make(map[string]map[string]float64)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		name, value := parseLine(line)
		if name == "" {
			continue
		}

		name = strings.Replace(name, "vllm:", "vllm_", 1)

		if collectHistograms && strings.Contains(name, "_bucket") {
			baseName := name[:strings.Index(name, "_bucket")]
			if _, ok := histMetrics[baseName]; ok {
				if le := extractLabel(name, "le"); le != "" {
					if buckets[baseName] == nil {
						buckets[baseName] = make(map[string]float64)
					}
					buckets[baseName][le] = value
					foundData = true
				}
			}
			continue
		}

		cleanName := name
		if idx := strings.IndexByte(name, '{'); idx > 0 {
			cleanName = name[:idx]
		}

		if ptr, ok := floatMetrics[cleanName]; ok {
			*ptr = base.MetricFloat{V: value, T: utils.GetTimestamp()}
			foundData = true
		}
	}

	if foundData {
		m.Available = true
	}

	if collectHistograms {
		for baseName, b := range buckets {
			if ptr, ok := histMetrics[baseName]; ok {
				if data, err := json.Marshal(b); err == nil {
					*ptr = base.MetricStr{V: string(data), T: utils.GetTimestamp()}
				}
			}
		}
	}
}

func parseLine(line string) (string, float64) {
	idx := strings.LastIndexByte(line, ' ')
	if idx <= 0 {
		return "", 0
	}
	return line[:idx], utils.ParseFloat64(line[idx+1:])
}

func extractLabel(s, label string) string {
	key := label + `="`
	start := strings.Index(s, key)
	if start < 0 {
		return ""
	}
	start += len(key)
	end := strings.IndexByte(s[start:], '"')
	if end < 0 {
		return ""
	}
	return s[start : start+end]
}
