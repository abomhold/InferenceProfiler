package graphing

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// createLineChart creates a line chart for raw or delta values.
func createLineChart(s *Series, isDelta bool) *charts.Line {
	line := charts.NewLine()

	title := formatName(s.Name)
	if isDelta {
		title += " (Delta)"
	} else {
		title += " (Raw)"
	}

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", AxisLabel: &opts.AxisLabel{Rotate: 45}}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value"}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "400px"}),
	)

	var xLabels []string
	var data []opts.LineData

	if isDelta {
		for i := 1; i < len(s.Timestamps); i++ {
			xLabels = append(xLabels, time.Unix(0, s.Timestamps[i]).Format("15:04:05.000"))
		}
		for _, v := range s.Deltas {
			data = append(data, opts.LineData{Value: v})
		}
	} else {
		for _, ts := range s.Timestamps {
			xLabels = append(xLabels, time.Unix(0, ts).Format("15:04:05.000"))
		}
		for _, v := range s.Values {
			data = append(data, opts.LineData{Value: v})
		}
	}

	line.SetXAxis(xLabels).AddSeries("", data,
		charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true), ShowSymbol: opts.Bool(true)}),
	)

	if isDelta {
		line.SetSeriesOptions(charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(0.2)}))
	}

	return line
}

// createHeatmap creates a heatmap showing histogram evolution over time.
func createHeatmap(h *Histogram) *charts.HeatMap {
	if len(h.Timestamps) < 2 || len(h.Buckets) == 0 {
		return nil
	}

	heatmap := charts.NewHeatMap()
	title := "vllm " + formatName(h.Name) + " (Heatmap)"

	// Sort bucket labels numerically
	labels := sortBucketLabels(h.Buckets)

	// Build time labels
	numPoints := len(h.Timestamps)
	timeLabels := make([]string, numPoints)
	for i, ts := range h.Timestamps {
		timeLabels[i] = time.Unix(0, ts).Format("15:04:05")
	}

	// Build heatmap data (convert cumulative to per-bucket counts)
	var data []opts.HeatMapData
	var maxVal float64

	for xi := 0; xi < numPoints; xi++ {
		var prev float64
		for yi, label := range labels {
			vals := h.Buckets[label]
			var cum float64
			if xi < len(vals) {
				cum = vals[xi]
			} else if len(vals) > 0 {
				cum = vals[len(vals)-1]
			}

			count := cum - prev
			if count < 0 {
				count = 0
			}
			if count > maxVal {
				maxVal = count
			}
			data = append(data, opts.HeatMapData{Value: [3]interface{}{xi, yi, count}})
			prev = cum
		}
	}

	// Format Y-axis labels
	yLabels := make([]string, len(labels))
	for i, l := range labels {
		if l == "inf" {
			yLabels[i] = "+Inf"
		} else {
			yLabels[i] = l
		}
	}

	heatmap.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", AxisLabel: &opts.AxisLabel{Rotate: 45}}),
		charts.WithYAxisOpts(opts.YAxis{Type: "category", Data: yLabels}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: opts.Bool(true),
			Min:        0,
			Max:        float32(maxVal),
			InRange: &opts.VisualMapInRange{
				Color: []string{
					"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8",
					"#ffffbf", "#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026",
				},
			},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "400px"}),
	)

	heatmap.SetXAxis(timeLabels).AddSeries("", data)
	return heatmap
}

// createBarChart creates a bar chart showing final histogram distribution.
func createBarChart(h *Histogram) *charts.Bar {
	bar := charts.NewBar()
	title := "vllm " + formatName(h.Name) + " (Distribution)"

	labels := sortBucketLabels(h.Buckets)

	lastIdx := len(h.Timestamps) - 1
	var xLabels []string
	var data []opts.BarData
	var prev float64

	for _, label := range labels {
		vals := h.Buckets[label]
		var cum float64
		if lastIdx >= 0 && lastIdx < len(vals) {
			cum = vals[lastIdx]
		}

		count := cum - prev
		if count < 0 {
			count = 0
		}

		if label == "inf" {
			xLabels = append(xLabels, "+Inf")
		} else {
			xLabels = append(xLabels, label)
		}
		data = append(data, opts.BarData{Value: count})
		prev = cum
	}

	// Get total from +Inf bucket
	var total float64
	for _, key := range []string{"+Inf", "inf"} {
		if vals, ok := h.Buckets[key]; ok && lastIdx >= 0 && lastIdx < len(vals) {
			total = vals[lastIdx]
			break
		}
	}

	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: fmt.Sprintf("Total: %.0f", total),
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", AxisLabel: &opts.AxisLabel{Rotate: 45}}),
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "400px"}),
	)

	bar.SetXAxis(xLabels).AddSeries("", data)
	return bar
}

// sortBucketLabels sorts histogram bucket labels numerically.
func sortBucketLabels(buckets map[string][]float64) []string {
	labels := make([]string, 0, len(buckets))
	for l := range buckets {
		labels = append(labels, l)
	}

	sort.Slice(labels, func(i, j int) bool {
		// +Inf/inf always goes last
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
