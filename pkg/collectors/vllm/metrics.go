package vllm

import "InferenceProfiler/pkg/collectors/types"

// Dynamic contains dynamic vLLM metrics.
type Dynamic struct {
	VLLMAvailable              bool
	VLLMTimestamp              int64
	VLLMRequestsRunning        float64
	VLLMRequestsWaiting        float64
	VLLMEngineSleepState       float64
	VLLMPreemptionsTotal       float64
	VLLMKvCacheUsagePercent    float64
	VLLMPrefixCacheHits        float64
	VLLMPrefixCacheQueries     float64
	VLLMRequestsFinishedTotal  float64
	VLLMRequestsCorruptedTotal float64
	VLLMTokensPromptTotal      float64
	VLLMTokensGenerationTotal  float64
	VLLMLatencyTtftSum         float64
	VLLMLatencyTtftCount       float64
	VLLMLatencyE2eSum          float64
	VLLMLatencyE2eCount        float64
	VLLMLatencyQueueSum        float64
	VLLMLatencyQueueCount      float64
	VLLMLatencyInferenceSum    float64
	VLLMLatencyInferenceCount  float64
	VLLMLatencyPrefillSum      float64
	VLLMLatencyPrefillCount    float64
	VLLMLatencyDecodeSum       float64
	VLLMLatencyDecodeCount     float64
	VLLMHistogramsJSON         string
}

// ToRecord converts Dynamic to a Record.
func (d *Dynamic) ToRecord() types.Record {
	r := types.Record{
		"vllmAvailable":              d.VLLMAvailable,
		"vllmTimestamp":              d.VLLMTimestamp,
		"vllmRequestsRunning":        d.VLLMRequestsRunning,
		"vllmRequestsWaiting":        d.VLLMRequestsWaiting,
		"vllmEngineSleepState":       d.VLLMEngineSleepState,
		"vllmPreemptionsTotal":       d.VLLMPreemptionsTotal,
		"vllmKvCacheUsagePercent":    d.VLLMKvCacheUsagePercent,
		"vllmPrefixCacheHits":        d.VLLMPrefixCacheHits,
		"vllmPrefixCacheQueries":     d.VLLMPrefixCacheQueries,
		"vllmRequestsFinishedTotal":  d.VLLMRequestsFinishedTotal,
		"vllmRequestsCorruptedTotal": d.VLLMRequestsCorruptedTotal,
		"vllmTokensPromptTotal":      d.VLLMTokensPromptTotal,
		"vllmTokensGenerationTotal":  d.VLLMTokensGenerationTotal,
		"vllmLatencyTtftSum":         d.VLLMLatencyTtftSum,
		"vllmLatencyTtftCount":       d.VLLMLatencyTtftCount,
		"vllmLatencyE2eSum":          d.VLLMLatencyE2eSum,
		"vllmLatencyE2eCount":        d.VLLMLatencyE2eCount,
		"vllmLatencyQueueSum":        d.VLLMLatencyQueueSum,
		"vllmLatencyQueueCount":      d.VLLMLatencyQueueCount,
		"vllmLatencyInferenceSum":    d.VLLMLatencyInferenceSum,
		"vllmLatencyInferenceCount":  d.VLLMLatencyInferenceCount,
		"vllmLatencyPrefillSum":      d.VLLMLatencyPrefillSum,
		"vllmLatencyPrefillCount":    d.VLLMLatencyPrefillCount,
		"vllmLatencyDecodeSum":       d.VLLMLatencyDecodeSum,
		"vllmLatencyDecodeCount":     d.VLLMLatencyDecodeCount,
	}
	if d.VLLMHistogramsJSON != "" {
		r["vllmHistogramsJson"] = d.VLLMHistogramsJSON
	}
	return r
}

// Histograms contains vLLM histogram data.
type Histograms struct {
	TokensPerStep           map[string]float64 `json:"tokensPerStep,omitempty"`
	LatencyTtft             map[string]float64 `json:"latencyTtft,omitempty"`
	LatencyE2e              map[string]float64 `json:"latencyE2e,omitempty"`
	LatencyQueue            map[string]float64 `json:"latencyQueue,omitempty"`
	LatencyInference        map[string]float64 `json:"latencyInference,omitempty"`
	LatencyPrefill          map[string]float64 `json:"latencyPrefill,omitempty"`
	LatencyDecode           map[string]float64 `json:"latencyDecode,omitempty"`
	LatencyInterToken       map[string]float64 `json:"latencyInterToken,omitempty"`
	ReqSizePromptTokens     map[string]float64 `json:"reqSizePromptTokens,omitempty"`
	ReqSizeGenerationTokens map[string]float64 `json:"reqSizeGenerationTokens,omitempty"`
	ReqParamsMaxTokens      map[string]float64 `json:"reqParamsMaxTokens,omitempty"`
	ReqParamsN              map[string]float64 `json:"reqParamsN,omitempty"`
}
