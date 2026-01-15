package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/parquet-go/parquet-go"
)

// GenerateGraphsFromOutputFile creates graphs from a completed output file
func GenerateGraphsFromOutputFile(outputPath string) error {
	dir := filepath.Dir(outputPath)
	base := filepath.Base(outputPath)
	sessionUUID := strings.TrimSuffix(base, filepath.Ext(base))

	// Load metrics data
	records, err := loadRecords(outputPath)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return fmt.Errorf("need at least 2 records")
	}

	// Load static info if available
	var staticInfo map[string]interface{}
	staticPath := filepath.Join(dir, sessionUUID+".json")
	if data, err := os.ReadFile(staticPath); err == nil {
		json.Unmarshal(data, &staticInfo)
	}

	return generateHTML(dir, sessionUUID, records, staticInfo)
}

// ExportGraphs generates graphs from the exporter's session data
func (e *Exporter) ExportGraphs() error {
	outputPath := filepath.Join(e.outputDir, fmt.Sprintf("%s.%s", e.sessionUUID, e.format))
	return GenerateGraphsFromOutputFile(outputPath)
}

// loadRecords reads records from a data file (jsonl, csv, tsv, or parquet)
func loadRecords(path string) ([]map[string]interface{}, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jsonl":
		return loadJSONL(path)
	case ".csv":
		return loadDelimited(path, ',')
	case ".tsv":
		return loadDelimited(path, '\t')
	case ".parquet":
		return loadParquet(path)
	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}
}

func loadJSONL(path string) ([]map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []map[string]interface{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var record map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, scanner.Err()
}

func loadDelimited(path string, delimiter rune) ([]map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = delimiter

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		record := make(map[string]interface{})
		for i, val := range row {
			if i < len(header) && val != "" {
				// Try to parse as number
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					record[header[i]] = f
				} else if val == "true" {
					record[header[i]] = true
				} else if val == "false" {
					record[header[i]] = false
				} else {
					record[header[i]] = val
				}
			}
		}
		records = append(records, record)
	}
	return records, nil
}

func loadParquet(path string) ([]map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	pf, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	reader := parquet.NewReader(pf)

	for {
		row := make(map[string]interface{})
		if err := reader.Read(&row); err != nil {
			break
		}
		records = append(records, row)
	}
	return records, nil
}

// generateHTML creates the HTML file with all charts
func generateHTML(outputDir, sessionUUID string, records []map[string]interface{}, staticInfo map[string]interface{}) error {
	// Sort by timestamp
	sort.Slice(records, func(i, j int) bool {
		return toFloat(records[i]["timestamp"]) < toFloat(records[j]["timestamp"])
	})

	// Build time series for each numeric column
	series := buildSeries(records)
	histograms := buildHistograms(records)

	// Create page with charts
	page := components.NewPage()
	page.PageTitle = "Profiler Metrics - " + sessionUUID

	// Add line charts (both raw and delta for each metric)
	for _, s := range series {
		if len(s.Values) < 2 {
			continue
		}
		page.AddCharts(createLineChart(s, false)) // Raw
		if len(s.Deltas) > 0 {
			page.AddCharts(createLineChart(s, true)) // Delta
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

	// Render and inject static info
	var buf strings.Builder
	if err := page.Render(&buf); err != nil {
		return err
	}

	html := buf.String()
	if staticInfo != nil {
		html = strings.Replace(html, "<body>", "<body>\n"+renderStaticInfo(sessionUUID, staticInfo), 1)
		html = strings.Replace(html, "</head>", cssAndScript()+"</head>", 1)
	}

	outputPath := filepath.Join(outputDir, sessionUUID+".html")
	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return err
	}

	log.Printf("Generated graphs: %s", outputPath)
	return nil
}

// Series holds time series data
type Series struct {
	Name       string
	Timestamps []int64
	Values     []float64
	Deltas     []float64
}

// Histogram holds histogram bucket data over time
type Histogram struct {
	Name       string
	Timestamps []int64
	Buckets    map[string][]float64
}

// buildSeries extracts numeric columns into time series with deltas
func buildSeries(records []map[string]interface{}) []*Series {
	// Find all numeric columns
	cols := make(map[string]bool)
	for _, r := range records {
		for k, v := range r {
			if k == "timestamp" || strings.HasSuffix(k, "T") || strings.HasSuffix(k, "Json") {
				continue
			}
			if _, ok := toFloatOk(v); ok {
				cols[k] = true
			}
		}
	}

	// Build series
	seriesMap := make(map[string]*Series)
	for col := range cols {
		seriesMap[col] = &Series{Name: col}
	}

	for _, r := range records {
		ts := int64(toFloat(r["timestamp"]))
		for col := range cols {
			s := seriesMap[col]
			if v, ok := toFloatOk(r[col]); ok {
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

	// Sort by name
	var result []*Series
	for _, s := range seriesMap {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// buildHistograms extracts vLLM histogram data
func buildHistograms(records []map[string]interface{}) []*Histogram {
	// First pass: collect all histogram names and their bucket labels
	histBuckets := make(map[string]map[string]bool) // histogram name -> set of bucket labels

	for _, r := range records {
		jsonStr, ok := r["vllmHistogramsJson"].(string)
		if !ok || jsonStr == "" {
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

	// Second pass: build histograms with all buckets aligned
	histMap := make(map[string]*Histogram)
	for name, labels := range histBuckets {
		h := &Histogram{Name: name, Buckets: make(map[string][]float64)}
		for label := range labels {
			h.Buckets[label] = []float64{}
		}
		histMap[name] = h
	}

	// Track last known values for carry-forward
	lastValues := make(map[string]map[string]float64) // histogram name -> bucket label -> last value

	for _, r := range records {
		ts := int64(toFloat(r["timestamp"]))
		jsonStr, ok := r["vllmHistogramsJson"].(string)

		// For each histogram, add a data point (using last known value if missing)
		for name, h := range histMap {
			if lastValues[name] == nil {
				lastValues[name] = make(map[string]float64)
			}

			var currentBuckets map[string]float64
			if ok && jsonStr != "" {
				var data map[string]map[string]float64
				if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
					currentBuckets = data[name]
				}
			}

			if currentBuckets != nil {
				// We have data for this timestamp
				h.Timestamps = append(h.Timestamps, ts)
				for label := range h.Buckets {
					var val float64
					if v, exists := currentBuckets[label]; exists {
						val = v
					} else {
						// Carry forward last known value for cumulative histograms
						val = lastValues[name][label]
					}
					h.Buckets[label] = append(h.Buckets[label], val)
					lastValues[name][label] = val
				}
			}
		}
	}

	var result []*Histogram
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

// createLineChart creates a line chart for raw or delta values
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

// createHeatmap creates a heatmap showing histogram evolution over time
func createHeatmap(h *Histogram) *charts.HeatMap {
	if len(h.Timestamps) < 2 || len(h.Buckets) == 0 {
		return nil
	}

	heatmap := charts.NewHeatMap()
	title := "vllm " + formatName(h.Name) + " (Heatmap)"

	// Sort bucket labels numerically
	var labels []string
	for l := range h.Buckets {
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

	// Use all time points (no sampling - let echarts handle zooming)
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
				cum = vals[len(vals)-1] // use last known value
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
			Calculable: opts.Bool(true), Min: 0, Max: float32(maxVal),
			InRange: &opts.VisualMapInRange{
				Color: []string{"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8", "#ffffbf", "#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026"},
			},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "400px"}),
	)

	heatmap.SetXAxis(timeLabels).AddSeries("", data)
	return heatmap
}

// createBarChart creates a bar chart showing final histogram distribution
func createBarChart(h *Histogram) *charts.Bar {
	bar := charts.NewBar()
	title := "vllm " + formatName(h.Name) + " (Distribution)"

	// Sort bucket labels
	var labels []string
	for l := range h.Buckets {
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

	lastIdx := len(h.Timestamps) - 1
	var xLabels []string
	var data []opts.BarData
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
	if vals, ok := h.Buckets["+Inf"]; ok && lastIdx < len(vals) {
		total = vals[lastIdx]
	} else if vals, ok := h.Buckets["inf"]; ok && lastIdx < len(vals) {
		total = vals[lastIdx]
	}

	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title, Subtitle: fmt.Sprintf("Total: %.0f", total)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", AxisLabel: &opts.AxisLabel{Rotate: 45}}),
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "400px"}),
	)

	bar.SetXAxis(xLabels).AddSeries("", data)
	return bar
}

// formatName converts camelCase to readable format
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

// toFloat converts to float64, returns 0 on failure
func toFloat(v interface{}) float64 {
	f, _ := toFloatOk(v)
	return f
}

// toFloatOk converts to float64 with success indicator
func toFloatOk(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

// renderStaticInfo generates HTML for static system info
func renderStaticInfo(sessionUUID string, info map[string]interface{}) string {
	var b strings.Builder

	b.WriteString(`<div class="static-info-container">`)
	b.WriteString(`<div class="static-info-header"><h1>System Profiler Report</h1>`)
	b.WriteString(`<div class="session-id">Session: ` + sessionUUID + `</div></div>`)

	// System info
	b.WriteString(`<div class="info-section"><h3>System</h3><table class="info-table">`)
	writeRow(&b, "UUID", info["uuid"])
	writeRow(&b, "VM ID", info["vId"])
	writeRow(&b, "Hostname", info["vHostname"])
	writeRow(&b, "Boot Time", formatTime(info["vBootTime"]))
	writeRow(&b, "Processors", info["vNumProcessors"])
	writeRow(&b, "CPU", info["vCpuType"])
	writeRow(&b, "CPU Cache", info["vCpuCache"])
	writeRow(&b, "Kernel", info["vKernelInfo"])
	writeRow(&b, "Time Synced", info["vTimeSynced"])
	writeRowf(&b, "Time Offset", info["vTimeOffsetSeconds"], "%.6f sec")
	writeRowf(&b, "Time Max Error", info["vTimeMaxErrorSeconds"], "%.6f sec")
	b.WriteString(`</table></div>`)

	// Memory
	b.WriteString(`<div class="info-section"><h3>Memory</h3><table class="info-table">`)
	writeRowBytes(&b, "Total", info["vMemoryTotalBytes"])
	writeRowBytes(&b, "Swap", info["vSwapTotalBytes"])
	b.WriteString(`</table></div>`)

	// Container
	if cid := info["cId"]; cid != nil && cid != "" {
		b.WriteString(`<div class="info-section"><h3>Container</h3><table class="info-table">`)
		writeRow(&b, "ID", cid)
		writeRow(&b, "CPUs", info["cNumProcessors"])
		writeRow(&b, "Cgroup Version", info["cCgroupVersion"])
		b.WriteString(`</table></div>`)
	}

	// Network
	if netJSON, ok := info["networkInterfaces"].(string); ok && netJSON != "" {
		var ifaces []map[string]interface{}
		if json.Unmarshal([]byte(netJSON), &ifaces) == nil && len(ifaces) > 0 {
			b.WriteString(`<div class="info-section"><h3>Network Interfaces</h3>`)
			for _, iface := range ifaces {
				b.WriteString(`<table class="info-table">`)
				writeRow(&b, "Name", iface["name"])
				writeRow(&b, "MAC", iface["mac"])
				writeRow(&b, "State", iface["state"])
				writeRow(&b, "MTU", iface["mtu"])
				writeRow(&b, "Speed (Mbps)", iface["speedMbps"])
				b.WriteString(`</table>`)
			}
			b.WriteString(`</div>`)
		}
	}

	// Disks
	if diskJSON, ok := info["disks"].(string); ok && diskJSON != "" {
		var disks []map[string]interface{}
		if json.Unmarshal([]byte(diskJSON), &disks) == nil && len(disks) > 0 {
			b.WriteString(`<div class="info-section"><h3>Disks</h3>`)
			for _, disk := range disks {
				b.WriteString(`<table class="info-table">`)
				writeRow(&b, "Name", disk["name"])
				writeRow(&b, "Model", disk["model"])
				writeRow(&b, "Vendor", disk["vendor"])
				writeRowBytes(&b, "Size", disk["sizeBytes"])
				b.WriteString(`</table>`)
			}
			b.WriteString(`</div>`)
		}
	}

	// GPU
	if count := toFloat(info["nvidiaGpuCount"]); count > 0 {
		b.WriteString(`<div class="info-section"><h3>NVIDIA GPU</h3><table class="info-table">`)
		writeRow(&b, "GPU Count", info["nvidiaGpuCount"])
		writeRow(&b, "Driver Version", info["nvidiaDriverVersion"])
		writeRow(&b, "CUDA Version", info["nvidiaCudaVersion"])
		writeRow(&b, "NVML Version", info["nvmlVersion"])
		b.WriteString(`</table>`)

		if gpuJSON, ok := info["nvidiaGpus"].(string); ok && gpuJSON != "" {
			var gpus []map[string]interface{}
			if json.Unmarshal([]byte(gpuJSON), &gpus) == nil {
				for i, gpu := range gpus {
					b.WriteString(fmt.Sprintf(`<div class="gpu-section"><strong>GPU %d: %v</strong><table class="info-table">`, i, gpu["name"]))
					writeRow(&b, "Index", gpu["index"])
					writeRow(&b, "UUID", gpu["uuid"])
					writeRow(&b, "Serial", gpu["serial"])
					writeRow(&b, "Board Part Number", gpu["boardPartNumber"])
					writeRow(&b, "Brand", gpu["brand"])
					writeRow(&b, "Architecture", gpu["architecture"])
					writeRowPair(&b, "CUDA Capability", gpu["cudaCapabilityMajor"], gpu["cudaCapabilityMinor"])
					writeRowBytes(&b, "Memory Total", gpu["memoryTotalBytes"])
					writeRowBytes(&b, "BAR1 Total", gpu["bar1TotalBytes"])
					writeRowUnit(&b, "Memory Bus Width", gpu["memoryBusWidthBits"], "bits")
					writeRow(&b, "CUDA Cores", gpu["numCores"])
					writeRowUnit(&b, "Max Clock Graphics", gpu["maxClockGraphicsMhz"], "MHz")
					writeRowUnit(&b, "Max Clock Memory", gpu["maxClockMemoryMhz"], "MHz")
					writeRowUnit(&b, "Max Clock SM", gpu["maxClockSmMhz"], "MHz")
					writeRowUnit(&b, "Max Clock Video", gpu["maxClockVideoMhz"], "MHz")
					writeRow(&b, "PCI Bus ID", gpu["pciBusId"])
					writeRow(&b, "PCI Device ID", gpu["pciDeviceId"])
					writeRow(&b, "PCI Subsystem ID", gpu["pciSubsystemId"])
					writeRow(&b, "PCIe Max Link Gen", gpu["pcieMaxLinkGen"])
					writeRow(&b, "PCIe Max Link Width", gpu["pcieMaxLinkWidth"])
					writeRowWatts(&b, "Power Default Limit", gpu["powerDefaultLimitMw"])
					writeRowWatts(&b, "Power Min Limit", gpu["powerMinLimitMw"])
					writeRowWatts(&b, "Power Max Limit", gpu["powerMaxLimitMw"])
					writeRow(&b, "VBIOS Version", gpu["vbiosVersion"])
					writeRow(&b, "Inforom Image", gpu["inforomImageVersion"])
					writeRow(&b, "Inforom OEM", gpu["inforomOemVersion"])
					writeRow(&b, "Num Fans", gpu["numFans"])
					writeRowUnit(&b, "Temp Shutdown", gpu["tempShutdownC"], "°C")
					writeRowUnit(&b, "Temp Slowdown", gpu["tempSlowdownC"], "°C")
					writeRowUnit(&b, "Temp Max Operating", gpu["tempMaxOperatingC"], "°C")
					writeRowUnit(&b, "Temp Target", gpu["tempTargetC"], "°C")
					writeRow(&b, "ECC Mode", gpu["eccModeEnabled"])
					writeRow(&b, "Persistence Mode", gpu["persistenceModeOn"])
					writeRow(&b, "Compute Mode", gpu["computeMode"])
					writeRow(&b, "Multi-GPU Board", gpu["isMultiGpuBoard"])
					writeRow(&b, "Display Mode", gpu["displayModeEnabled"])
					writeRow(&b, "Display Active", gpu["displayActive"])
					writeRow(&b, "MIG Mode", gpu["migModeEnabled"])
					writeRow(&b, "Encoder H264", gpu["encoderCapacityH264"])
					writeRow(&b, "Encoder HEVC", gpu["encoderCapacityHevc"])
					writeRow(&b, "Encoder AV1", gpu["encoderCapacityAv1"])
					writeRow(&b, "NVLink Count", gpu["nvlinkCount"])
					b.WriteString(`</table></div>`)
				}
			}
		}
		b.WriteString(`</div>`)
	}

	b.WriteString(`</div>`)
	return b.String()
}

func writeRow(b *strings.Builder, label string, val interface{}) {
	if val == nil {
		return
	}
	s := fmt.Sprintf("%v", val)
	if s == "" || s == "<nil>" {
		return
	}
	b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s</td></tr>`, label, s))
}

func writeRowUnit(b *strings.Builder, label string, val interface{}, unit string) {
	if val == nil {
		return
	}
	f := toFloat(val)
	if f == 0 {
		return
	}
	b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%v %s</td></tr>`, label, val, unit))
}

func writeRowPair(b *strings.Builder, label string, v1, v2 interface{}) {
	if v1 == nil || v2 == nil {
		return
	}
	b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%v.%v</td></tr>`, label, v1, v2))
}

func writeRowf(b *strings.Builder, label string, val interface{}, format string) {
	if val == nil {
		return
	}
	b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>`+format+`</td></tr>`, label, val))
}

func writeRowBytes(b *strings.Builder, label string, val interface{}) {
	bytes := toFloat(val)
	if bytes == 0 {
		b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>0 B</td></tr>`, label))
		return
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for bytes >= 1024 && i < len(units)-1 {
		bytes /= 1024
		i++
	}
	b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%.2f %s</td></tr>`, label, bytes, units[i]))
}

func writeRowWatts(b *strings.Builder, label string, val interface{}) {
	mw := toFloat(val)
	if mw == 0 {
		return
	}
	b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%.0f W</td></tr>`, label, mw/1000))
}

func formatTime(val interface{}) string {
	ts := int64(toFloat(val))
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func cssAndScript() string {
	return `
<style>
* { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif; }
body { max-width: 1400px; margin: 0 auto; padding: 20px; }
.static-info-container { margin-bottom: 20px; }
.static-info-header { border-bottom: 2px solid #333; padding-bottom: 10px; margin-bottom: 15px; }
.static-info-header h1 { margin: 0; font-size: 18px; }
.session-id { font-size: 11px; color: #666; font-family: monospace; }
.info-section { margin-bottom: 15px; padding: 15px; background: #f5f5f5; border: 1px solid #ddd; }
.info-section h3 { margin: 0 0 10px 0; font-size: 13px; }
.info-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.info-table td { padding: 3px 8px; border-bottom: 1px solid #eee; }
.info-table td:first-child { width: 150px; color: #666; }
.info-table td:last-child { font-family: monospace; font-size: 11px; word-break: break-all; }
.info-table tr:last-child td { border-bottom: none; }
.gpu-section { background: #fff; border: 1px solid #ddd; padding: 10px; margin-top: 10px; }
.container { display: block !important; margin: 0 0 10px 0 !important; padding: 15px !important; background: #f5f5f5 !important; border: 1px solid #ddd !important; overflow: hidden !important; }
.item { margin: 0 !important; }
</style>
<script>
window.addEventListener('resize', function() {
    document.querySelectorAll('[_echarts_instance_]').forEach(function(el) {
        var c = echarts.getInstanceByDom(el); if (c) c.resize();
    });
});
window.addEventListener('load', function() {
    setTimeout(function() {
        document.querySelectorAll('[_echarts_instance_]').forEach(function(el) {
            var c = echarts.getInstanceByDom(el); if (c) c.resize();
        });
    }, 100);
});
</script>
`
}
