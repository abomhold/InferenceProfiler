package base

import (
	"InferenceProfiler/pkg/utils"
	"context"
)

type Writer interface {
	Static(name string, data any) error
	Dynamic(name string, data any) error
	Flush() error
	Close() error
}

type Collector interface {
	Name() string
	Init(cfg *utils.Config) error
	Static() any
	Poll(ctx context.Context) any
	Close() error
}

type MetricInt struct {
	V int64 `json:"V"`
	T int64 `json:"T"`
}

type MetricFloat struct {
	V float64 `json:"V"`
	T int64   `json:"T"`
}

type MetricStr struct {
	V string `json:"V"`
	T int64  `json:"T"`
}
