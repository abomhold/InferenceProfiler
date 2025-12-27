package collectors

import (
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

var ignoredLabels = map[string]bool{
	"model_name": true, "model": true, "engine_id": true,
	"engine": true, "handler": true, "method": true,
}

var metricAliases = map[string]string{
	"num_requests_running":           "vllmSystemRequestsRunning",
	"num_requests_waiting":           "vllmSystemRequestsWaiting",
	"engine_sleep_state":             "vllmSystemEngineSleepState",
	"num_preemptions":                "vllmSystemPreemptionsTotal",
	"cache_config_info":              "vllmConfigCache",
	"kv_cache_usage_perc":            "vllmCacheKvUsagePercent",
	"prefix_cache_hits":              "vllmCachePrefixHits",
	"prefix_cache_queries":           "vllmCachePrefixQueries",
	"mm_cache_hits":                  "vllmCacheMultimodalHits",
	"mm_cache_queries":               "vllmCacheMultimodalQueries",
	"request_success":                "vllmRequestsFinishedTotal",
	"corrupted_requests":             "vllmRequestsCorruptedTotal",
	"prompt_tokens":                  "vllmTokensPromptTotal",
	"generation_tokens":              "vllmTokensGenerationTotal",
	"iteration_tokens_total":         "vllmTokensPerStepHistogram",
	"time_to_first_token_seconds":    "vllmLatencyTtftS",
	"e2e_request_latency_seconds":    "vllmLatencyE2eS",
	"request_queue_time_seconds":     "vllmLatencyQueueS",
	"request_inference_time_seconds": "vllmLatencyInferenceS",
	"request_prefill_time_seconds":   "vllmLatencyPrefillS",
	"request_decode_time_seconds":    "vllmLatencyDecodeS",
	"inter_token_latency_seconds":    "vllmLatencyInterTokenS",
	"request_prompt_tokens":          "vllmReqSizePromptTokens",
	"request_generation_tokens":      "vllmReqSizeGenerationTokens",
	"request_params_max_tokens":      "vllmReqParamsMaxTokens",
	"request_params_n":               "vllmReqParamsN",
}

func CollectVLLM() map[string]MetricValue {
	metrics := make(map[string]MetricValue)
	scrapeTime := GetTimestamp()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(vllmMetricsURL)
	if err != nil {
		return metrics
	}
	defer resp.Body.Close()

	families, err := (&expfmt.TextParser{}).TextToMetricFamilies(resp.Body)
	if err != nil {
		return metrics
	}

	for name, mf := range families {
		baseName := getCleanName(name)
		isInfo := strings.HasSuffix(name, "_info")

		for _, m := range mf.GetMetric() {
			labels := filterLabels(m.GetLabel())
			suffix := labelSuffix(labels)

			// Info metrics: extract label values as separate metrics
			if isInfo {
				for k, v := range labels {
					metrics[baseName+"_"+k] = NewMetricWithTime(parseNumber(v), scrapeTime)
				}
				continue
			}

			// Histogram: emit sum, count, and bucket map
			if h := m.GetHistogram(); h != nil {
				metrics[baseName+"_sum"+suffix] = NewMetricWithTime(sanitize(h.GetSampleSum()), scrapeTime)
				metrics[baseName+"_count"+suffix] = NewMetricWithTime(float64(h.GetSampleCount()), scrapeTime)

				buckets := make(map[string]float64)
				for _, b := range h.GetBucket() {
					le := b.GetUpperBound()
					key := "inf"
					if !math.IsInf(le, 1) {
						key = fmt.Sprintf("%g", le)
					}
					buckets[key] = float64(b.GetCumulativeCount())
				}
				metrics[baseName+suffix+"_histogram"] = NewMetricWithTime(buckets, scrapeTime)
				continue
			}

			// Summary: emit sum, count, and quantiles
			if s := m.GetSummary(); s != nil {
				metrics[baseName+"_sum"+suffix] = NewMetricWithTime(sanitize(s.GetSampleSum()), scrapeTime)
				metrics[baseName+"_count"+suffix] = NewMetricWithTime(float64(s.GetSampleCount()), scrapeTime)
				for _, q := range s.GetQuantile() {
					qKey := fmt.Sprintf("%s_quantile_%g%s", baseName, q.GetQuantile(), suffix)
					metrics[qKey] = NewMetricWithTime(sanitize(q.GetValue()), scrapeTime)
				}
				continue
			}

			// Counter, Gauge, Untyped: extract single value
			val := getValue(m)
			metrics[baseName+suffix] = NewMetricWithTime(sanitize(val), scrapeTime)
		}
	}

	if len(metrics) > 0 {
		metrics["vllmTimestamp"] = NewMetricWithTime(scrapeTime, scrapeTime)
	}
	return metrics
}

func getCleanName(name string) string {
	lookup := strings.TrimPrefix(name, "vllm:")
	if alias, ok := metricAliases[lookup]; ok {
		return alias
	}
	return strings.ReplaceAll(name, ":", "_")
}

func filterLabels(pairs []*dto.LabelPair) map[string]string {
	labels := make(map[string]string)
	for _, p := range pairs {
		if !ignoredLabels[p.GetName()] {
			labels[p.GetName()] = p.GetValue()
		}
	}
	return labels
}

func labelSuffix(labels map[string]string) string {
	var b strings.Builder
	for k, v := range labels {
		b.WriteString("_")
		b.WriteString(k)
		b.WriteString("_")
		b.WriteString(v)
	}
	return b.String()
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

func parseNumber(s string) interface{} {
	var i int64
	if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
		return i
	}
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
		return f
	}
	return s
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
