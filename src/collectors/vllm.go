package collectors

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var vllmMetricsURL = getVLLMMetricsURL()

func getVLLMMetricsURL() string {
	if url := os.Getenv("VLLM_METRICS_URL"); url != "" {
		return url
	}
	return "http://localhost:8000/metrics"
}

var prometheusRegex = regexp.MustCompile(`^([a-zA-Z0-9_:]+)(?:\{(.+)\})?\s+([0-9\.eE\+\-]+|nan|inf|NaN|Inf)$`)

var ignoredLabels = map[string]bool{
	"model_name": true,
	"model":      true,
	"engine_id":  true,
	"engine":     true,
	"handler":    true,
	"method":     true,
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

// CollectVLLM collects vLLM inference engine metrics
func CollectVLLM() map[string]MetricValue {
	metrics := make(map[string]MetricValue)
	scrapeTime := GetTimestamp()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(vllmMetricsURL)
	if err != nil {
		return metrics
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return metrics
	}

	parsed := parsePrometheus(string(body))
	for k, v := range parsed {
		metrics[k] = NewMetricWithTime(v, scrapeTime)
	}

	if len(metrics) > 0 {
		metrics["vllmTimestamp"] = NewMetricWithTime(scrapeTime, scrapeTime)
	}

	return metrics
}

func getCleanName(rawName string) string {
	lookupName := strings.Replace(rawName, "vllm:", "", 1)
	if alias, ok := metricAliases[lookupName]; ok {
		return alias
	}
	return strings.ReplaceAll(rawName, ":", "_")
}

func parsePrometheus(text string) map[string]interface{} {
	data := make(map[string]interface{})
	histograms := make(map[string]map[string]float64)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		match := prometheusRegex.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		name := match[1]
		labelStr := match[2]
		valueStr := match[3]

		var val float64
		lowerVal := strings.ToLower(valueStr)
		if strings.Contains(lowerVal, "nan") {
			val = 0.0
		} else if strings.Contains(lowerVal, "inf") {
			val = 0.0
		} else {
			var err error
			val, err = strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}
		}

		labels := make(map[string]string)
		if labelStr != "" {
			parts := strings.Split(labelStr, ",")
			for _, p := range parts {
				if idx := strings.Index(p, "="); idx != -1 {
					k := strings.TrimSpace(p[:idx])
					if !ignoredLabels[k] {
						v := strings.Trim(strings.TrimSpace(p[idx+1:]), "\"")
						labels[k] = v
					}
				}
			}
		}

		isBucket := strings.HasSuffix(name, "_bucket")
		isSum := strings.HasSuffix(name, "_sum")
		isCount := strings.HasSuffix(name, "_count")
		isInfo := strings.HasSuffix(name, "_info")

		baseLookup := name
		suffix := ""

		if isBucket {
			baseLookup = name[:len(name)-7]
			suffix = "_bucket"
		} else if isSum {
			baseLookup = name[:len(name)-4]
			suffix = "_sum"
		} else if isCount {
			baseLookup = name[:len(name)-6]
			suffix = "_count"
		}

		cleanBase := getCleanName(baseLookup)

		if isInfo && len(labels) > 0 {
			for k, v := range labels {
				infoKey := cleanBase + "_" + k
				data[infoKey] = tryParseNumber(v)
			}
			continue
		}

		if isBucket {
			if le, ok := labels["le"]; ok {
				delete(labels, "le")

				var leVal float64
				if strings.Contains(strings.ToLower(le), "inf") {
					leVal = -1
				} else {
					leVal, _ = strconv.ParseFloat(le, 64)
				}

				histoKey := cleanBase
				for k, v := range labels {
					histoKey += "_" + k + "_" + v
				}

				if histograms[histoKey] == nil {
					histograms[histoKey] = make(map[string]float64)
				}

				leKeyStr := le
				if leVal == -1 {
					leKeyStr = "inf"
				}
				histograms[histoKey][leKeyStr] = val
			}
		} else {
			finalKey := cleanBase + suffix

			for k, v := range labels {
				finalKey += "_" + k + "_" + v
			}

			data[finalKey] = val
		}
	}

	for key, buckets := range histograms {
		data[key+"_histogram"] = buckets
	}

	return data
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
