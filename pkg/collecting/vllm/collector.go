package vllm

import (
	"InferenceProfiler/pkg/utils"
	"context"
	"log"
	"net/http"
	"time"
)

type Collector struct {
	endpoint    string
	collectHist bool
	client      *http.Client
	last        *vllmDynamic
}

func New() *Collector { return &Collector{} }

func (c *Collector) Name() string { return "Vllm" }

func (c *Collector) Init(cfg *utils.Config) error {
	c.endpoint = cfg.VLLMEndpoint
	c.collectHist = !cfg.DisableVLLMHistograms
	c.client = utils.NewHTTPClient(1*time.Second, 100*time.Millisecond, 500*time.Millisecond, 1)

	log.Printf("vllm: endpoint=%s histograms=%v", c.endpoint, c.collectHist)
	return nil
}

func (c *Collector) Static() any { return nil }

func (c *Collector) Poll(ctx context.Context) any {
	body, err := utils.HTTPGet(ctx, c.client, c.endpoint)
	if err != nil {
		if c.last == nil {
			log.Printf("vllm: endpoint not available")
		}
		if c.last != nil {
			cached := *c.last
			cached.Available = false
			return &cached
		}
		return nil
	}
	defer body.Close()

	var m vllmDynamic
	parseVllm(body, c.collectHist, &m)

	c.last = &m
	return &m
}

func (c *Collector) Close() error { return nil }
