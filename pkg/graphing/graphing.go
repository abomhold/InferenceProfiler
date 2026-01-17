// Package graphing provides visualization generation for profiler metrics.
package graphing

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"InferenceProfiler/pkg/formatting"

	"github.com/go-echarts/go-echarts/v2/components"
)

// Generator creates visualizations from profiler metrics.
type Generator struct {
	inputPath  string
	outputPath string
	format     string
	records    []formatting.Record
	staticInfo map[string]interface{}
}

// NewGenerator creates a new graph generator.
func NewGenerator(inputPath, outputPath, format string) (*Generator, error) {
	if inputPath == "" {
		return nil, fmt.Errorf("input path is required")
	}
	if outputPath == "" {
		return nil, fmt.Errorf("output path is required")
	}

	return &Generator{
		inputPath:  inputPath,
		outputPath: outputPath,
		format:     format,
	}, nil
}

// Generate creates the visualization output.
func (g *Generator) Generate() error {
	// Load records
	records, err := formatting.LoadRecords(g.inputPath)
	if err != nil {
		return fmt.Errorf("failed to load records: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("need at least 2 records to generate graphs, got %d", len(records))
	}

	g.records = records

	// Try to load static info
	g.loadStaticInfo()

	// Sort records by timestamp
	sort.Slice(g.records, func(i, j int) bool {
		return formatting.ToFloat(g.records[i]["timestamp"]) < formatting.ToFloat(g.records[j]["timestamp"])
	})

	switch g.format {
	case "html":
		return g.generateHTML()
	case "png", "svg":
		return fmt.Errorf("format %s not yet implemented", g.format)
	default:
		return g.generateHTML()
	}
}

func (g *Generator) loadStaticInfo() {
	// Try to find static info file
	dir := filepath.Dir(g.inputPath)
	base := filepath.Base(g.inputPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	// Try different naming patterns
	patterns := []string{
		filepath.Join(dir, name+".json"),
		filepath.Join(dir, "static_"+name+".json"),
	}

	for _, pattern := range patterns {
		if data, err := os.ReadFile(pattern); err == nil {
			var info map[string]interface{}
			if json.Unmarshal(data, &info) == nil {
				g.staticInfo = info
				return
			}
		}
	}
}

func (g *Generator) generateHTML() error {
	series := buildSeries(g.records)
	histograms := buildHistograms(g.records)

	page := components.NewPage()
	page.PageTitle = "Profiler Metrics"

	// Add line charts
	for _, s := range series {
		if len(s.Values) < 2 {
			continue
		}
		page.AddCharts(createLineChart(s, false)) // Raw values
		if len(s.Deltas) > 0 {
			page.AddCharts(createLineChart(s, true)) // Delta values
		}
	}

	// Add histogram charts
	for _, h := range histograms {
		if len(h.Buckets) == 0 {
			continue
		}
		if heatmap := createHeatmap(h); heatmap != nil {
			page.AddCharts(heatmap)
		}
		page.AddCharts(createBarChart(h))
	}

	// Render charts HTML
	var chartBuf strings.Builder
	if err := page.Render(&chartBuf); err != nil {
		return fmt.Errorf("failed to render charts: %w", err)
	}

	chartHTML := chartBuf.String()

	// Extract just the body content from echarts output
	// The echarts library generates a full HTML page, we need to integrate it
	var finalHTML string

	if g.staticInfo != nil {
		sessionID := extractSessionID(g.inputPath)

		// Render static info using templates
		staticInfoHTML, err := renderStaticInfoHTML(sessionID, g.staticInfo)
		if err != nil {
			log.Printf("Warning: failed to render static info: %v", err)
			staticInfoHTML = ""
		}

		// Get styles and scripts from template
		stylesScripts, err := renderStylesAndScripts()
		if err != nil {
			log.Printf("Warning: failed to render styles/scripts: %v", err)
			stylesScripts = ""
		}

		// Inject into chart HTML
		finalHTML = strings.Replace(chartHTML, "<body>", "<body>\n"+staticInfoHTML, 1)
		finalHTML = strings.Replace(finalHTML, "</head>", stylesScripts+"</head>", 1)
	} else {
		// Just add styles/scripts
		stylesScripts, _ := renderStylesAndScripts()
		finalHTML = strings.Replace(chartHTML, "</head>", stylesScripts+"</head>", 1)
	}

	// Write output
	if err := os.WriteFile(g.outputPath, []byte(finalHTML), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	log.Printf("Generated graphs: %s", g.outputPath)
	return nil
}

func extractSessionID(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}

// GenerateGraphsFromFile is a convenience function for simple graph generation.
func GenerateGraphsFromFile(inputPath, outputPath string) error {
	format := "html"
	if strings.HasSuffix(outputPath, ".png") {
		format = "png"
	} else if strings.HasSuffix(outputPath, ".svg") {
		format = "svg"
	}

	gen, err := NewGenerator(inputPath, outputPath, format)
	if err != nil {
		return err
	}
	return gen.Generate()
}

// Series holds time series data for a single metric.
type Series struct {
	Name       string
	Timestamps []int64
	Values     []float64
	Deltas     []float64
}

// Histogram holds histogram bucket data over time.
type Histogram struct {
	Name       string
	Timestamps []int64
	Buckets    map[string][]float64
}

// buildSeries extracts numeric columns into time series with deltas.
func buildSeries(records []formatting.Record) []*Series {
	// Find all numeric columns
	cols := make(map[string]bool)
	for _, r := range records {
		for k, v := range r {
			// Skip timestamp, T-suffix columns, and JSON columns
			if k == "timestamp" ||
				strings.HasSuffix(k, "T") ||
				strings.HasSuffix(k, "Json") ||
				strings.HasSuffix(k, "JSON") {
				continue
			}
			if _, ok := formatting.ToFloatOk(v); ok {
				cols[k] = true
			}
		}
	}

	// Build series for each column
	seriesMap := make(map[string]*Series)
	for col := range cols {
		seriesMap[col] = &Series{Name: col}
	}

	for _, r := range records {
		fallbackTs := int64(formatting.ToFloat(r["timestamp"]))

		for col := range cols {
			s := seriesMap[col]
			if v, ok := formatting.ToFloatOk(r[col]); ok {
				ts := fallbackTs
				// Try to get metric-specific timestamp
				if metricTs, ok := formatting.ToFloatOk(r[col+"T"]); ok && metricTs > 0 {
					ts = int64(metricTs)
				}
				s.Timestamps = append(s.Timestamps, ts)
				s.Values = append(s.Values, v)
			}
		}
	}

	// Calculate deltas
	for _, s := range seriesMap {
		for i := 1; i < len(s.Values); i++ {
			s.Deltas = append(s.Deltas, s.Values[i]-s.Values[i-1])
		}
	}

	// Convert to sorted slice
	result := make([]*Series, 0, len(seriesMap))
	for _, s := range seriesMap {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// buildHistograms extracts vLLM histogram data.
func buildHistograms(records []formatting.Record) []*Histogram {
	// First pass: collect all histogram names and bucket labels
	histBuckets := make(map[string]map[string]bool)

	for _, r := range records {
		jsonStr := getHistogramJSON(r)
		if jsonStr == "" {
			continue
		}

		var data map[string]map[string]float64
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}

		for name, buckets := range data {
			if histBuckets[name] == nil {
				histBuckets[name] = make(map[string]bool)
			}
			for label := range buckets {
				histBuckets[name][label] = true
			}
		}
	}

	if len(histBuckets) == 0 {
		return nil
	}

	// Second pass: build histograms with aligned buckets
	histMap := make(map[string]*Histogram)
	for name, labels := range histBuckets {
		h := &Histogram{
			Name:    name,
			Buckets: make(map[string][]float64),
		}
		for label := range labels {
			h.Buckets[label] = []float64{}
		}
		histMap[name] = h
	}

	// Track last known values for carry-forward
	lastValues := make(map[string]map[string]float64)

	for _, r := range records {
		ts := int64(formatting.ToFloat(r["timestamp"]))

		jsonStr := getHistogramJSON(r)
		var currentData map[string]map[string]float64
		if jsonStr != "" {
			json.Unmarshal([]byte(jsonStr), &currentData)
		}

		for name, h := range histMap {
			if lastValues[name] == nil {
				lastValues[name] = make(map[string]float64)
			}

			currentBuckets := currentData[name]
			if currentBuckets != nil {
				h.Timestamps = append(h.Timestamps, ts)
				for label := range h.Buckets {
					var val float64
					if v, exists := currentBuckets[label]; exists {
						val = v
					} else {
						val = lastValues[name][label]
					}
					h.Buckets[label] = append(h.Buckets[label], val)
					lastValues[name][label] = val
				}
			}
		}
	}

	// Convert to sorted slice
	result := make([]*Histogram, 0, len(histMap))
	for _, h := range histMap {
		if len(h.Timestamps) > 0 {
			result = append(result, h)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func getHistogramJSON(r formatting.Record) string {
	keys := []string{
		"vllmHistogramsJson",
		"VLLMHistogramsJSON",
		"vllmHistogramsJSON",
	}

	for _, key := range keys {
		if val, exists := r[key]; exists {
			switch v := val.(type) {
			case string:
				if v != "" && v != "{}" {
					return v
				}
			case []byte:
				s := string(v)
				if s != "" && s != "{}" {
					return s
				}
			}
		}
	}
	return ""
}

// formatName converts camelCase to readable format.
func formatName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return result.String()
}

// formatTime converts a timestamp to readable format.
func formatTime(val interface{}) string {
	ts := int64(formatting.ToFloat(val))
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
