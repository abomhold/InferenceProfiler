package collectors

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// --- Constants & Global Helpers ---

const JiffiesPerSecond = 100

// GetTimestamp returns current time in nanoseconds
func GetTimestamp() int64 {
	return time.Now().UnixNano()
}

// parseInt64 is a quiet parser for DRYer metric collection
func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

// --- Probing & Filesystem Utilities ---

// ProbeFile reads a file and returns its content with timestamp
func ProbeFile(path string) (string, int64) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ts
	}
	return strings.TrimSpace(string(data)), ts
}

// ProbeFileInt reads a file and converts content to int64
func ProbeFileInt(path string) (int64, int64) {
	content, ts := ProbeFile(path)
	if content == "" {
		return 0, ts
	}
	return parseInt64(content), ts
}

// ProbeFileLines reads a file and returns lines
func ProbeFileLines(path string) ([]string, int64) {
	ts := GetTimestamp()
	file, err := os.Open(path)
	if err != nil {
		return nil, ts
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, ts
}

// ProbeFunction executes a function and returns result with timestamp
func ProbeFunction[T any](fn func() (T, error), defaultVal T) (T, int64) {
	ts := GetTimestamp()
	result, err := fn()
	if err != nil {
		log.Printf("ProbeFunction error: %v", err)
		return defaultVal, ts
	}
	return result, ts
}

// ParseProcKV parses a key-value file like /proc/meminfo or /proc/self/status
func ParseProcKV(path string, separator string) (map[string]string, int64) {
	lines, ts := ProbeFileLines(path)
	data := make(map[string]string)
	for _, line := range lines {
		if before, after, found := strings.Cut(line, separator); found {
			data[strings.TrimSpace(before)] = strings.TrimSpace(after)
		}
	}
	return data, ts
}

// --- Metric Structures ---

// MetricValue represents a single metric data point
type MetricValue struct {
	Value interface{} `json:"value"`
	Time  int64       `json:"time,omitempty"`
}

func NewMetric(value interface{}) MetricValue {
	return MetricValue{Value: value}
}

func NewMetricWithTime(value interface{}, ts int64) MetricValue {
	return MetricValue{Value: value, Time: ts}
}

type BaseMap[T any] map[string]T

func (m BaseMap[T]) Merge(src map[string]T) BaseMap[T] {
	for k, v := range src {
		m[k] = v
	}
	return m
}

func (m BaseMap[T]) MergeWithPrefix(prefix string, src map[string]T) BaseMap[T] {
	for k, v := range src {
		m[prefix+k] = v
	}
	return m
}

type DynamicMetrics = BaseMap[MetricValue]

func NewDynamicMetrics() DynamicMetrics {
	m := make(DynamicMetrics)
	m["timestamp"] = NewMetric(GetTimestamp())
	return m
}

type StaticMetrics = BaseMap[interface{}]

func NewStaticMetrics(sessionUUID uuid.UUID) StaticMetrics {
	s := make(StaticMetrics)
	s["uuid"] = sessionUUID.String()
	return s
}
