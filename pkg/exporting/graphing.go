// Package graphing provides visualization generation for profiler metrics.
package exporting

import (
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const (
	defaultWidth  = 12 * vg.Inch
	defaultHeight = 4 * vg.Inch
)

// Generator creates visualizations from profiler metrics.
type Generator struct {
	inputPath string
	outputDir string
	records   []Record
}

// NewGenerator creates a new graph generator.
func NewGenerator(inputPath, outputDir, _ string) (*Generator, error) {
	if inputPath == "" {
		return nil, fmt.Errorf("input path is required")
	}
	if outputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	return &Generator{
		inputPath: inputPath,
		outputDir: outputDir,
	}, nil
}

// Generate creates PNG visualizations in the output directory.
func (g *Generator) Generate() error {
	records, err := LoadRecords(g.inputPath)
	if err != nil {
		return fmt.Errorf("failed to load records: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("need at least 2 records to generate graphs, got %d", len(records))
	}

	g.records = records

	// Sort records by timestamp
	sort.Slice(g.records, func(i, j int) bool {
		return utils.ToFloat64(g.records[i]["timestamp"]) < utils.ToFloat64(g.records[j]["timestamp"])
	})

	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build and render series
	series := buildSeries(g.records)
	for _, s := range series {
		if len(s.Values) < 2 {
			continue
		}

		// Raw values chart
		if err := g.renderLineChart(s, false); err != nil {
			log.Printf("Warning: failed to render %s: %v", s.Name, err)
		}

		// Delta chart
		if len(s.Deltas) > 0 {
			if err := g.renderLineChart(s, true); err != nil {
				log.Printf("Warning: failed to render %s delta: %v", s.Name, err)
			}
		}
	}

	// Build and render histograms
	histograms := buildHistograms(g.records)
	for _, h := range histograms {
		if len(h.Buckets) == 0 {
			continue
		}
		if err := g.renderBarChart(h); err != nil {
			log.Printf("Warning: failed to render histogram %s: %v", h.Name, err)
		}
	}

	log.Printf("Generated graphs in: %s", g.outputDir)
	return nil
}

func (g *Generator) renderLineChart(s *Series, isDelta bool) error {
	p := plot.New()

	title := formatName(s.Name)
	suffix := "_raw"
	if isDelta {
		title += " (Delta)"
		suffix = "_delta"
	}
	p.Title.Text = title
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Value"

	var pts plotter.XYs
	var timestamps []int64
	var values []float64

	if isDelta {
		timestamps = s.Timestamps[1:]
		values = s.Deltas
	} else {
		timestamps = s.Timestamps
		values = s.Values
	}

	if len(timestamps) == 0 || len(values) == 0 {
		return nil
	}

	baseTime := timestamps[0]
	for i, ts := range timestamps {
		if i < len(values) {
			pts = append(pts, plotter.XY{
				X: float64(ts-baseTime) / 1e9, // Convert to seconds from start
				Y: values[i],
			})
		}
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	line.Color = plotutil.Color(0)
	p.Add(line)
	p.Add(plotter.NewGrid())

	filename := sanitizeFilename(s.Name) + suffix + ".png"
	return p.Save(defaultWidth, defaultHeight, filepath.Join(g.outputDir, filename))
}

func (g *Generator) renderBarChart(h *Histogram) error {
	p := plot.New()
	p.Title.Text = "vLLM " + formatName(h.Name) + " Distribution"
	p.X.Label.Text = "Bucket"
	p.Y.Label.Text = "Count"

	labels := sortBucketLabels(h.Buckets)
	lastIdx := len(h.Timestamps) - 1
	if lastIdx < 0 {
		return nil
	}

	var values plotter.Values
	var prev float64

	for _, label := range labels {
		vals := h.Buckets[label]
		var cum float64
		if lastIdx < len(vals) {
			cum = vals[lastIdx]
		}
		count := cum - prev
		if count < 0 {
			count = 0
		}
		values = append(values, count)
		prev = cum
	}

	bars, err := plotter.NewBarChart(values, vg.Points(20))
	if err != nil {
		return err
	}
	bars.Color = plotutil.Color(0)
	p.Add(bars)

	// Set X axis labels
	p.NominalX(formatBucketLabels(labels)...)

	filename := "vllm_" + sanitizeFilename(h.Name) + "_dist.png"
	return p.Save(defaultWidth, defaultHeight, filepath.Join(g.outputDir, filename))
}

// GenerateGraphsFromFile is a convenience function for simple graph generation.
func GenerateGraphsFromFile(inputPath, outputPath string) error {
	// If outputPath looks like a file, use its directory
	outputDir := outputPath
	if strings.HasSuffix(outputPath, ".png") || strings.HasSuffix(outputPath, ".html") {
		outputDir = filepath.Dir(outputPath)
		if outputDir == "" || outputDir == "." {
			outputDir = "graphs"
		}
	}

	gen, err := NewGenerator(inputPath, outputDir, "png")
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
func buildSeries(records []Record) []*Series {
	cols := make(map[string]bool)
	for _, r := range records {
		for k, v := range r {
			if k == "timestamp" ||
				strings.HasSuffix(k, "T") ||
				strings.HasSuffix(k, "Json") ||
				strings.HasSuffix(k, "JSON") {
				continue
			}
			if _, ok := utils.ToFloat64Ok(v); ok {
				cols[k] = true
			}
		}
	}

	seriesMap := make(map[string]*Series)
	for col := range cols {
		seriesMap[col] = &Series{Name: col}
	}

	for _, r := range records {
		fallbackTs := int64(utils.ToFloat64(r["timestamp"]))

		for col := range cols {
			s := seriesMap[col]
			if v, ok := utils.ToFloat64Ok(r[col]); ok {
				ts := fallbackTs
				if metricTs, ok := utils.ToFloat64Ok(r[col+"T"]); ok && metricTs > 0 {
					ts = int64(metricTs)
				}
				s.Timestamps = append(s.Timestamps, ts)
				s.Values = append(s.Values, v)
			}
		}
	}

	for _, s := range seriesMap {
		for i := 1; i < len(s.Values); i++ {
			s.Deltas = append(s.Deltas, s.Values[i]-s.Values[i-1])
		}
	}

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
func buildHistograms(records []Record) []*Histogram {
	// First pass: collect all histogram names and bucket labels
	histBuckets := make(map[string]map[string]bool)

	for _, r := range records {
		jsonStr := getHistogramJSON(r)
		if jsonStr == "" {
			continue
		}

		var data map[string]map[string]float64
		if err := parseJSON(jsonStr, &data); err != nil {
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
		ts := int64(utils.ToFloat64(r["timestamp"]))

		jsonStr := getHistogramJSON(r)
		var currentData map[string]map[string]float64
		if jsonStr != "" {
			parseJSON(jsonStr, &currentData)
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

func getHistogramJSON(r Record) string {
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

func parseJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

func sortBucketLabels(buckets map[string][]float64) []string {
	labels := make([]string, 0, len(buckets))
	for l := range buckets {
		labels = append(labels, l)
	}

	sort.Slice(labels, func(i, j int) bool {
		if labels[i] == "+Inf" || labels[i] == "inf" {
			return false
		}
		if labels[j] == "+Inf" || labels[j] == "inf" {
			return true
		}
		vi, _ := strconv.ParseFloat(labels[i], 64)
		vj, _ := strconv.ParseFloat(labels[j], 64)
		return vi < vj
	})

	return labels
}

func formatBucketLabels(labels []string) []string {
	result := make([]string, len(labels))
	for i, l := range labels {
		if l == "inf" {
			result[i] = "+Inf"
		} else {
			result[i] = l
		}
	}
	return result
}

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

func sanitizeFilename(name string) string {
	// Replace problematic characters
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return strings.ToLower(name)
}

// GenerateDeltaGraph creates a bar chart visualization of delta values.
func GenerateDeltaGraph(delta DeltaResult, outputDir string) error {
	if outputDir == "" {
		outputDir = "delta_graphs"
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract numeric fields from delta record
	var fields []fieldValue
	for key, val := range delta.Record {
		// Skip metadata and non-numeric fields
		if strings.HasPrefix(key, "_delta_") {
			continue
		}
		if strings.HasSuffix(key, "Json") || strings.HasSuffix(key, "JSON") {
			continue
		}
		if strings.HasSuffix(key, "T") {
			continue
		}
		if key == "timestamp" || key == "uuid" {
			continue
		}

		if f, ok := utils.ToFloat64Ok(val); ok {
			fields = append(fields, fieldValue{name: key, value: f})
		}
	}

	if len(fields) == 0 {
		return fmt.Errorf("no numeric fields found in delta")
	}

	// Sort by name for consistent output
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].name < fields[j].name
	})

	// Group fields by category for separate charts
	categories := map[string][]fieldValue{
		"cpu":       {},
		"memory":    {},
		"disk":      {},
		"network":   {},
		"gpu":       {},
		"container": {},
		"process":   {},
		"other":     {},
	}

	for _, f := range fields {
		lowerName := strings.ToLower(f.name)
		switch {
		case strings.HasPrefix(lowerName, "vcpu") || strings.HasPrefix(lowerName, "cpu"):
			categories["cpu"] = append(categories["cpu"], f)
		case strings.HasPrefix(lowerName, "vmemory") || strings.HasPrefix(lowerName, "mem"):
			categories["memory"] = append(categories["memory"], f)
		case strings.HasPrefix(lowerName, "vdisk") || strings.HasPrefix(lowerName, "disk"):
			categories["disk"] = append(categories["disk"], f)
		case strings.HasPrefix(lowerName, "vnetwork") || strings.HasPrefix(lowerName, "network"):
			categories["network"] = append(categories["network"], f)
		case strings.HasPrefix(lowerName, "nvidia") || strings.HasPrefix(lowerName, "gpu"):
			categories["gpu"] = append(categories["gpu"], f)
		case strings.HasPrefix(lowerName, "c") && (strings.Contains(lowerName, "cpu") || strings.Contains(lowerName, "mem") || strings.Contains(lowerName, "disk") || strings.Contains(lowerName, "network")):
			categories["container"] = append(categories["container"], f)
		case strings.HasPrefix(lowerName, "p") && (strings.Contains(lowerName, "cpu") || strings.Contains(lowerName, "mem")):
			categories["process"] = append(categories["process"], f)
		default:
			categories["other"] = append(categories["other"], f)
		}
	}

	// Generate chart for each non-empty category
	for category, catFields := range categories {
		if len(catFields) == 0 {
			continue
		}

		if err := generateDeltaBarChart(catFields, category, outputDir); err != nil {
			log.Printf("Warning: failed to generate %s chart: %v", category, err)
		}
	}

	// Generate summary chart with top changes (by absolute value)
	topFields := make([]fieldValue, len(fields))
	copy(topFields, fields)
	sort.Slice(topFields, func(i, j int) bool {
		return absFloat(topFields[i].value) > absFloat(topFields[j].value)
	})
	if len(topFields) > 20 {
		topFields = topFields[:20]
	}
	if err := generateDeltaBarChart(topFields, "top_changes", outputDir); err != nil {
		log.Printf("Warning: failed to generate top changes chart: %v", err)
	}

	return nil
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func generateDeltaBarChart(fields []fieldValue, title, outputDir string) error {
	if len(fields) == 0 {
		return nil
	}

	p := plot.New()
	p.Title.Text = fmt.Sprintf("Delta: %s", formatName(title))
	p.X.Label.Text = "Metric"
	p.Y.Label.Text = "Delta Value"

	// Create bar values
	values := make(plotter.Values, len(fields))
	labels := make([]string, len(fields))

	for i, f := range fields {
		values[i] = f.value
		// Truncate long labels
		label := f.name
		if len(label) > 25 {
			label = label[:22] + "..."
		}
		labels[i] = label
	}

	bars, err := plotter.NewBarChart(values, vg.Points(15))
	if err != nil {
		return err
	}

	// Color bars based on positive/negative
	bars.Color = plotutil.Color(0)
	p.Add(bars)
	p.Add(plotter.NewGrid())

	// Set x-axis labels
	p.NominalX(labels...)

	// Rotate labels if there are many
	if len(fields) > 5 {
		p.X.Tick.Label.Rotation = 0.8 // ~45 degrees
		p.X.Tick.Label.XAlign = 1
		p.X.Tick.Label.YAlign = 0.5
	}

	filename := fmt.Sprintf("delta_%s.png", sanitizeFilename(title))
	return p.Save(defaultWidth, defaultHeight, filepath.Join(outputDir, filename))
}

type fieldValue struct {
	name  string
	value float64
}
