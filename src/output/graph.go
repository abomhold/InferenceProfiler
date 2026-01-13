package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/parquet-go/parquet-go"
)

// GraphGenerator creates visualization charts from metrics data
type GraphGenerator struct {
	outputDir   string
	sessionUUID string
	staticInfo  map[string]interface{} // Static system info to display at top
}

// NewGraphGenerator creates a new graph generator
func NewGraphGenerator(outputDir, sessionUUID string) *GraphGenerator {
	return &GraphGenerator{
		outputDir:   outputDir,
		sessionUUID: sessionUUID,
	}
}

// SetStaticInfo sets the static system information to display at the top of the HTML
func (g *GraphGenerator) SetStaticInfo(info map[string]interface{}) {
	g.staticInfo = info
}

// LoadStaticInfoFromFile loads static info from a JSON file
func (g *GraphGenerator) LoadStaticInfoFromFile(jsonPath string) error {
	file, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to open static info file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var data map[string]interface{}
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode static info: %w", err)
	}

	g.staticInfo = data
	return nil
}

// MetricSeries holds time series data for a single metric
type MetricSeries struct {
	Name       string
	Timestamps []int64
	Values     []float64
	Deltas     []float64
}

// HistogramSeries holds time series data for histogram buckets
type HistogramSeries struct {
	Name       string
	Timestamps []int64
	Buckets    map[string][]float64 // bucket label -> values over time
}

// GenerateFromRecords creates charts from a slice of flattened metric records
func (g *GraphGenerator) GenerateFromRecords(records []map[string]interface{}) error {
	if len(records) < 2 {
		return fmt.Errorf("need at least 2 records to calculate deltas")
	}

	// Sort records by timestamp
	sort.Slice(records, func(i, j int) bool {
		ts1, _ := toFloat64(records[i]["timestamp"])
		ts2, _ := toFloat64(records[j]["timestamp"])
		return ts1 < ts2
	})

	// Extract all numeric columns (excluding timestamp fields ending in 'T')
	columns := g.extractNumericColumns(records)

	// Build time series for each column
	seriesMap := make(map[string]*MetricSeries)
	for _, col := range columns {
		seriesMap[col] = &MetricSeries{
			Name:       col,
			Timestamps: make([]int64, 0, len(records)),
			Values:     make([]float64, 0, len(records)),
			Deltas:     make([]float64, 0, len(records)-1),
		}
	}

	// Populate values
	for _, record := range records {
		ts, ok := toFloat64(record["timestamp"])
		if !ok {
			continue
		}
		timestamp := int64(ts)

		for _, col := range columns {
			series := seriesMap[col]
			if val, ok := toFloat64(record[col]); ok {
				series.Timestamps = append(series.Timestamps, timestamp)
				series.Values = append(series.Values, val)
			}
		}
	}

	// Calculate deltas
	for _, series := range seriesMap {
		for i := 1; i < len(series.Values); i++ {
			delta := series.Values[i] - series.Values[i-1]
			series.Deltas = append(series.Deltas, delta)
		}
	}

	// Extract vLLM histogram data
	histogramMap := g.extractVLLMHistograms(records)

	// Group series by category for organized output
	categories := g.categorizeMetrics(seriesMap)

	// Generate combined HTML page
	return g.generateCombinedHTML(categories, seriesMap, histogramMap)
}

// extractVLLMHistograms parses vLLM histogram JSON fields from records
func (g *GraphGenerator) extractVLLMHistograms(records []map[string]interface{}) map[string]*HistogramSeries {
	histogramMap := make(map[string]*HistogramSeries)

	// Known histogram field names in vLLMHistogramsJson
	histogramNames := []string{
		"tokensPerStep", "latencyTtft", "latencyE2e", "latencyQueue",
		"latencyInference", "latencyPrefill", "latencyDecode", "latencyInterToken",
		"reqSizePromptTokens", "reqSizeGenerationTokens", "reqParamsMaxTokens", "reqParamsN",
	}

	for _, name := range histogramNames {
		histogramMap[name] = &HistogramSeries{
			Name:       name,
			Timestamps: make([]int64, 0, len(records)),
			Buckets:    make(map[string][]float64),
		}
	}

	for _, record := range records {
		ts, ok := toFloat64(record["timestamp"])
		if !ok {
			continue
		}
		timestamp := int64(ts)

		// Look for vllmHistogramsJson field
		histJSON, ok := record["vllmHistogramsJson"].(string)
		if !ok || histJSON == "" {
			continue
		}

		// Parse the JSON
		var histograms map[string]map[string]float64
		if err := json.Unmarshal([]byte(histJSON), &histograms); err != nil {
			continue
		}

		// Process each histogram type
		for histName, buckets := range histograms {
			series, exists := histogramMap[histName]
			if !exists {
				series = &HistogramSeries{
					Name:       histName,
					Timestamps: make([]int64, 0),
					Buckets:    make(map[string][]float64),
				}
				histogramMap[histName] = series
			}

			series.Timestamps = append(series.Timestamps, timestamp)

			// Add each bucket value
			for bucketLabel, value := range buckets {
				if _, exists := series.Buckets[bucketLabel]; !exists {
					series.Buckets[bucketLabel] = make([]float64, 0)
				}
				// Pad with zeros if this bucket wasn't present in earlier records
				for len(series.Buckets[bucketLabel]) < len(series.Timestamps)-1 {
					series.Buckets[bucketLabel] = append(series.Buckets[bucketLabel], 0)
				}
				series.Buckets[bucketLabel] = append(series.Buckets[bucketLabel], value)
			}

			// Ensure all buckets have the same length
			for bucketLabel := range series.Buckets {
				for len(series.Buckets[bucketLabel]) < len(series.Timestamps) {
					series.Buckets[bucketLabel] = append(series.Buckets[bucketLabel], 0)
				}
			}
		}
	}

	return histogramMap
}

// GenerateFromSnapshotFiles reads snapshot files and generates charts
func (g *GraphGenerator) GenerateFromSnapshotFiles(snapshotFiles []string) error {
	if len(snapshotFiles) < 2 {
		return fmt.Errorf("need at least 2 snapshot files to calculate deltas")
	}

	sort.Strings(snapshotFiles)

	var records []map[string]interface{}
	for _, filePath := range snapshotFiles {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		decoder := json.NewDecoder(file)
		decoder.UseNumber()

		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			file.Close()
			continue
		}
		file.Close()

		convertJSONNumbers(data)
		records = append(records, data)
	}

	return g.GenerateFromRecords(records)
}

// GenerateFromJSONL reads a JSONL file and generates charts
func (g *GraphGenerator) GenerateFromJSONL(jsonlPath string) error {
	file, err := os.Open(jsonlPath)
	if err != nil {
		return fmt.Errorf("failed to open JSONL file: %w", err)
	}
	defer file.Close()

	var records []map[string]interface{}
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var record map[string]interface{}
		decoder := json.NewDecoder(strings.NewReader(string(line)))
		decoder.UseNumber()

		if err := decoder.Decode(&record); err != nil {
			log.Printf("Warning: skipping malformed JSONL line: %v", err)
			continue
		}

		convertJSONNumbers(record)
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading JSONL file: %w", err)
	}

	return g.GenerateFromRecords(records)
}

// GenerateFromParquet reads a Parquet file and generates charts
func (g *GraphGenerator) GenerateFromParquet(parquetPath string) error {
	file, err := os.Open(parquetPath)
	if err != nil {
		return fmt.Errorf("failed to open Parquet file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat Parquet file: %w", err)
	}

	pf, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		return fmt.Errorf("failed to open Parquet file: %w", err)
	}

	var records []map[string]interface{}
	reader := parquet.NewReader(pf)

	for {
		row := make(map[string]interface{})
		err := reader.Read(&row)
		if err != nil {
			break // End of file or error
		}
		records = append(records, row)
	}

	if len(records) == 0 {
		return fmt.Errorf("no records found in Parquet file")
	}

	return g.GenerateFromRecords(records)
}

// GenerateFromFile auto-detects file format and generates charts
func (g *GraphGenerator) GenerateFromFile(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".jsonl":
		return g.GenerateFromJSONL(filePath)
	case ".parquet":
		return g.GenerateFromParquet(filePath)
	case ".json":
		// Could be a single JSON file or we need to look for pattern
		return g.GenerateFromJSONL(filePath) // Try JSONL format
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}

// extractNumericColumns finds all numeric columns (excluding timestamp suffix 'T' fields)
func (g *GraphGenerator) extractNumericColumns(records []map[string]interface{}) []string {
	columnSet := make(map[string]bool)

	for _, record := range records {
		for key, val := range record {
			// Skip timestamp fields (ending in 'T') and the main timestamp
			if key == "timestamp" || strings.HasSuffix(key, "T") {
				continue
			}
			// Skip JSON string fields
			if strings.HasSuffix(key, "Json") || strings.HasSuffix(key, "JSON") {
				continue
			}
			// Skip non-numeric fields
			if _, ok := toFloat64(val); ok {
				columnSet[key] = true
			}
		}
	}

	var columns []string
	for col := range columnSet {
		columns = append(columns, col)
	}
	sort.Strings(columns)
	return columns
}

// categorizeMetrics groups metrics by their prefix/category
// Order matters: check more specific prefixes before generic ones
func (g *GraphGenerator) categorizeMetrics(seriesMap map[string]*MetricSeries) map[string][]string {
	categories := map[string][]string{
		"VM CPU (v*)":     {},
		"VM Memory (v*)":  {},
		"VM Disk (v*)":    {},
		"VM Network (v*)": {},
		"Container (c*)":  {},
		"NVIDIA GPU":      {},
		"vLLM":            {},
		"Process (p*)":    {},
		"Other":           {},
	}

	for name := range seriesMap {
		category := g.getMetricCategory(name)
		categories[category] = append(categories[category], name)
	}

	// Sort each category
	for cat := range categories {
		sort.Strings(categories[cat])
	}

	return categories
}

// getMetricCategory determines the category for a metric name
// Order matters: check more specific prefixes before generic ones
func (g *GraphGenerator) getMetricCategory(name string) string {
	// Check vLLM FIRST before any v* patterns
	if strings.HasPrefix(name, "vllm") {
		return "vLLM"
	}

	// Check NVIDIA
	if strings.HasPrefix(name, "nvidia") {
		return "NVIDIA GPU"
	}

	// Check process metrics
	if strings.HasPrefix(name, "process") {
		return "Process (p*)"
	}

	// Check container metrics (c* but not clocksEventReasons which is nvidia)
	if strings.HasPrefix(name, "c") && !strings.HasPrefix(name, "clocksEventReasons") {
		return "Container (c*)"
	}

	// Check VM metrics (v* prefixes)
	if strings.HasPrefix(name, "vCpu") || strings.HasPrefix(name, "vLoad") {
		return "VM CPU (v*)"
	}
	if strings.HasPrefix(name, "vMemory") {
		return "VM Memory (v*)"
	}
	if strings.HasPrefix(name, "vDisk") {
		return "VM Disk (v*)"
	}
	if strings.HasPrefix(name, "vNetwork") {
		return "VM Network (v*)"
	}

	return "Other"
}

// generateCombinedHTML creates a single HTML file with all charts
func (g *GraphGenerator) generateCombinedHTML(categories map[string][]string, seriesMap map[string]*MetricSeries, histogramMap map[string]*HistogramSeries) error {
	page := components.NewPage()
	page.PageTitle = fmt.Sprintf("Profiler Metrics - %s", g.sessionUUID)

	// Define category order
	categoryOrder := []string{
		"VM CPU (v*)", "VM Memory (v*)", "VM Disk (v*)", "VM Network (v*)",
		"Container (c*)", "NVIDIA GPU", "vLLM", "Process (p*)", "Other",
	}

	chartsAdded := 0
	for _, category := range categoryOrder {
		metrics, ok := categories[category]
		if !ok || len(metrics) == 0 {
			continue
		}

		for _, metricName := range metrics {
			series, ok := seriesMap[metricName]
			if !ok || len(series.Values) < 2 {
				continue
			}

			// Create both raw and delta charts
			rawChart := g.createLineChart(series, category, false)
			page.AddCharts(rawChart)
			chartsAdded++

			if len(series.Deltas) > 0 {
				deltaChart := g.createLineChart(series, category, true)
				page.AddCharts(deltaChart)
				chartsAdded++
			}
		}
	}

	// Add vLLM histogram charts
	histChartsAdded := g.addHistogramCharts(page, histogramMap)
	chartsAdded += histChartsAdded

	if chartsAdded == 0 {
		return fmt.Errorf("no charts generated - all metrics had zero deltas or insufficient data")
	}

	// Render to buffer first
	var buf strings.Builder
	if err := page.Render(&buf); err != nil {
		return fmt.Errorf("failed to render charts: %w", err)
	}

	// Inject static info after <body> tag
	htmlContent := buf.String()
	if g.staticInfo != nil && len(g.staticInfo) > 0 {
		staticHTML := g.generateStaticInfoHTML()
		htmlContent = strings.Replace(htmlContent, "<body>", "<body>\n"+staticHTML, 1)

		// Also inject custom CSS into <head>
		customCSS := g.generateCustomCSS()
		htmlContent = strings.Replace(htmlContent, "</head>", customCSS+"</head>", 1)
	}

	// Write to file
	outputPath := filepath.Join(g.outputDir, fmt.Sprintf("%s-graphs.html", g.sessionUUID))
	if err := os.WriteFile(outputPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	log.Printf("Generated graphs: %s (%d charts)", outputPath, chartsAdded)
	return nil
}

// generateCustomCSS returns CSS styles for the static info section
func (g *GraphGenerator) generateCustomCSS() string {
	return `
    <style>
        * {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
        }
        body {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
            font-size: 14px;
        }
        .static-info-container {
            margin: 0 0 20px 0;
        }
        .static-info-header {
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #333;
        }
        .static-info-header h1 {
            margin: 0;
            font-size: 18px;
            font-weight: bold;
        }
        .static-info-header .session-id {
            font-size: 11px;
            color: #666;
            font-family: monospace;
        }
        .info-section {
            margin-bottom: 15px;
            padding: 15px;
            background: #f5f5f5;
            border: 1px solid #ddd;
        }
        .info-section h3 {
            margin: 0 0 10px 0;
            font-size: 13px;
            font-weight: bold;
            color: #333;
        }
        .info-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 12px;
        }
        .info-table td {
            padding: 3px 8px;
            vertical-align: top;
            border-bottom: 1px solid #eee;
        }
        .info-table tr:last-child td {
            border-bottom: none;
        }
        .info-table td:first-child {
            width: 200px;
            color: #666;
        }
        .info-table td:last-child {
            color: #000;
            word-break: break-all;
            font-family: monospace;
            font-size: 11px;
        }
        .info-grid {
            display: flex;
            flex-direction: column;
            gap: 10px;
        }
        .gpu-section, .disk-section, .net-section {
            background: #fff;
            border: 1px solid #ddd;
            padding: 10px;
            margin-top: 10px;
        }
        /* Override go-echarts inline styles that appear in body */
        .container {
            display: block !important;
            justify-content: unset !important;
            align-items: unset !important;
            margin: 0 0 10px 0 !important;
            padding: 15px !important;
            background: #f5f5f5 !important;
            border: 1px solid #ddd !important;
            box-sizing: border-box !important;
            overflow: hidden !important;
        }
        .item {
            margin: 0 !important;
        }
        .item > div[_echarts_instance_] {
            width: 100% !important;
        }
    </style>
    <script>
        window.addEventListener('resize', function() {
            // Resize all echarts instances on window resize
            var charts = document.querySelectorAll('[_echarts_instance_]');
            charts.forEach(function(el) {
                var chart = echarts.getInstanceByDom(el);
                if (chart) {
                    chart.resize();
                }
            });
        });
        // Also resize on load to ensure proper initial sizing
        window.addEventListener('load', function() {
            setTimeout(function() {
                var charts = document.querySelectorAll('[_echarts_instance_]');
                charts.forEach(function(el) {
                    var chart = echarts.getInstanceByDom(el);
                    if (chart) {
                        chart.resize();
                    }
                });
            }, 100);
        });
    </script>
`
}

// generateStaticInfoHTML creates a plain HTML section displaying static system info
func (g *GraphGenerator) generateStaticInfoHTML() string {
	if g.staticInfo == nil || len(g.staticInfo) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(`<div class="static-info-container">
    <div class="static-info-header">
        <h1>System Profiler Report</h1>
        <div class="session-id">Session: ` + g.sessionUUID + `</div>
    </div>
    <div class="info-grid">
`)

	// System Info Section
	sb.WriteString(`        <div class="info-section">
            <h3>System Information</h3>
            <table class="info-table">
`)
	g.writeTableRow(&sb, "UUID", g.staticInfo["uuid"])
	g.writeTableRow(&sb, "VM ID", g.staticInfo["vId"])
	g.writeTableRow(&sb, "Hostname", g.staticInfo["vHostname"])
	g.writeTableRow(&sb, "Boot Time", g.formatBootTime(g.staticInfo["vBootTime"]))
	g.writeTableRow(&sb, "Processors", g.staticInfo["vNumProcessors"])
	g.writeTableRow(&sb, "CPU Type", g.staticInfo["vCpuType"])
	g.writeTableRow(&sb, "CPU Cache", g.staticInfo["vCpuCache"])
	g.writeTableRow(&sb, "Kernel", g.staticInfo["vKernelInfo"])
	g.writeTableRow(&sb, "Time Synced", g.staticInfo["vTimeSynced"])
	g.writeTableRowFloat(&sb, "Time Offset", g.staticInfo["vTimeOffsetSeconds"], " sec")
	g.writeTableRowFloat(&sb, "Time Max Error", g.staticInfo["vTimeMaxErrorSeconds"], " sec")
	sb.WriteString(`            </table>
        </div>
`)

	// Memory Section
	sb.WriteString(`        <div class="info-section">
            <h3>Memory</h3>
            <table class="info-table">
`)
	g.writeTableRowBytes(&sb, "Total Memory", g.staticInfo["vMemoryTotalBytes"])
	g.writeTableRowBytes(&sb, "Total Swap", g.staticInfo["vSwapTotalBytes"])
	sb.WriteString(`            </table>
        </div>
`)

	// Container Section
	if cid, ok := g.staticInfo["cId"]; ok && cid != nil && cid != "" {
		sb.WriteString(`        <div class="info-section">
            <h3>Container</h3>
            <table class="info-table">
`)
		g.writeTableRow(&sb, "Container ID", cid)
		g.writeTableRow(&sb, "CPUs", g.staticInfo["cNumProcessors"])
		g.writeTableRow(&sb, "Cgroup Version", g.staticInfo["cCgroupVersion"])
		sb.WriteString(`            </table>
        </div>
`)
	}

	// Network Section
	if netJSON, ok := g.staticInfo["networkInterfaces"].(string); ok && netJSON != "" {
		var interfaces []map[string]interface{}
		if err := json.Unmarshal([]byte(netJSON), &interfaces); err == nil && len(interfaces) > 0 {
			sb.WriteString(`        <div class="info-section">
            <h3>Network Interfaces</h3>
`)
			for _, iface := range interfaces {
				sb.WriteString(`            <div class="net-section">
                <table class="info-table">
`)
				g.writeTableRow(&sb, "Name", iface["name"])
				g.writeTableRow(&sb, "MAC", iface["mac"])
				g.writeTableRow(&sb, "State", iface["state"])
				g.writeTableRow(&sb, "MTU", iface["mtu"])
				g.writeTableRow(&sb, "Speed (Mbps)", iface["speedMbps"])
				sb.WriteString(`                </table>
            </div>
`)
			}
			sb.WriteString(`        </div>
`)
		}
	}

	// Disks Section
	if disksJSON, ok := g.staticInfo["disks"].(string); ok && disksJSON != "" {
		var disks []map[string]interface{}
		if err := json.Unmarshal([]byte(disksJSON), &disks); err == nil && len(disks) > 0 {
			sb.WriteString(`        <div class="info-section">
            <h3>Disks</h3>
`)
			for _, disk := range disks {
				sb.WriteString(`            <div class="disk-section">
                <table class="info-table">
`)
				g.writeTableRow(&sb, "Name", disk["name"])
				g.writeTableRow(&sb, "Model", disk["model"])
				g.writeTableRow(&sb, "Vendor", disk["vendor"])
				g.writeTableRowBytes(&sb, "Size", disk["sizeBytes"])
				sb.WriteString(`                </table>
            </div>
`)
			}
			sb.WriteString(`        </div>
`)
		}
	}

	sb.WriteString(`    </div>
`)

	// NVIDIA GPU Section (full width, detailed)
	if gpuCount, ok := g.staticInfo["nvidiaGpuCount"]; ok && gpuCount != nil {
		count := 0
		switch v := gpuCount.(type) {
		case float64:
			count = int(v)
		case int:
			count = v
		case int64:
			count = int(v)
		}

		if count > 0 {
			sb.WriteString(`    <div class="info-section">
        <h3>NVIDIA GPU</h3>
        <table class="info-table">
`)
			g.writeTableRow(&sb, "GPU Count", gpuCount)
			g.writeTableRow(&sb, "Driver Version", g.staticInfo["nvidiaDriverVersion"])
			g.writeTableRow(&sb, "CUDA Version", g.staticInfo["nvidiaCudaVersion"])
			g.writeTableRow(&sb, "NVML Version", g.staticInfo["nvmlVersion"])
			sb.WriteString(`        </table>
`)

			// Parse and display each GPU's full details
			if gpusJSON, ok := g.staticInfo["nvidiaGpus"].(string); ok && gpusJSON != "" {
				var gpus []map[string]interface{}
				if err := json.Unmarshal([]byte(gpusJSON), &gpus); err == nil {
					for i, gpu := range gpus {
						sb.WriteString(fmt.Sprintf(`        <div class="gpu-section">
            <strong>GPU %d: %v</strong>
            <table class="info-table">
`, i, gpu["name"]))
						g.writeTableRow(&sb, "Index", gpu["index"])
						g.writeTableRow(&sb, "UUID", gpu["uuid"])
						g.writeTableRow(&sb, "Serial", gpu["serial"])
						g.writeTableRow(&sb, "Board Part Number", gpu["boardPartNumber"])
						g.writeTableRow(&sb, "Brand", gpu["brand"])
						g.writeTableRow(&sb, "Architecture", gpu["architecture"])
						g.writeGpuTableRow(&sb, "CUDA Capability", gpu["cudaCapabilityMajor"], gpu["cudaCapabilityMinor"], "%v.%v")
						g.writeTableRowBytes(&sb, "Memory Total", gpu["memoryTotalBytes"])
						g.writeTableRowBytes(&sb, "BAR1 Total", gpu["bar1TotalBytes"])
						g.writeGpuTableRowUnit(&sb, "Memory Bus Width", gpu["memoryBusWidthBits"], "bits")
						g.writeTableRow(&sb, "CUDA Cores", gpu["numCores"])
						g.writeGpuTableRowUnit(&sb, "Max Clock Graphics", gpu["maxClockGraphicsMhz"], "MHz")
						g.writeGpuTableRowUnit(&sb, "Max Clock Memory", gpu["maxClockMemoryMhz"], "MHz")
						g.writeGpuTableRowUnit(&sb, "Max Clock SM", gpu["maxClockSmMhz"], "MHz")
						g.writeGpuTableRowUnit(&sb, "Max Clock Video", gpu["maxClockVideoMhz"], "MHz")
						g.writeTableRow(&sb, "PCI Bus ID", gpu["pciBusId"])
						g.writeTableRow(&sb, "PCI Device ID", gpu["pciDeviceId"])
						g.writeTableRow(&sb, "PCI Subsystem ID", gpu["pciSubsystemId"])
						g.writeTableRow(&sb, "PCIe Max Link Gen", gpu["pcieMaxLinkGen"])
						g.writeTableRow(&sb, "PCIe Max Link Width", gpu["pcieMaxLinkWidth"])
						g.writeTableRowMw(&sb, "Power Default Limit", gpu["powerDefaultLimitMw"])
						g.writeTableRowMw(&sb, "Power Min Limit", gpu["powerMinLimitMw"])
						g.writeTableRowMw(&sb, "Power Max Limit", gpu["powerMaxLimitMw"])
						g.writeTableRow(&sb, "VBIOS Version", gpu["vbiosVersion"])
						g.writeTableRow(&sb, "Inforom Image Version", gpu["inforomImageVersion"])
						g.writeTableRow(&sb, "Inforom OEM Version", gpu["inforomOemVersion"])
						g.writeTableRow(&sb, "Num Fans", gpu["numFans"])
						g.writeGpuTableRowUnit(&sb, "Temp Shutdown", gpu["tempShutdownC"], "°C")
						g.writeGpuTableRowUnit(&sb, "Temp Slowdown", gpu["tempSlowdownC"], "°C")
						g.writeGpuTableRowUnit(&sb, "Temp Max Operating", gpu["tempMaxOperatingC"], "°C")
						g.writeGpuTableRowUnit(&sb, "Temp Target", gpu["tempTargetC"], "°C")
						g.writeTableRow(&sb, "ECC Mode", gpu["eccModeEnabled"])
						g.writeTableRow(&sb, "Persistence Mode", gpu["persistenceModeOn"])
						g.writeTableRow(&sb, "Compute Mode", gpu["computeMode"])
						g.writeTableRow(&sb, "Multi-GPU Board", gpu["isMultiGpuBoard"])
						g.writeTableRow(&sb, "Display Mode", gpu["displayModeEnabled"])
						g.writeTableRow(&sb, "Display Active", gpu["displayActive"])
						g.writeTableRow(&sb, "MIG Mode", gpu["migModeEnabled"])
						g.writeTableRow(&sb, "Encoder Capacity H264", gpu["encoderCapacityH264"])
						g.writeTableRow(&sb, "Encoder Capacity HEVC", gpu["encoderCapacityHevc"])
						g.writeTableRow(&sb, "Encoder Capacity AV1", gpu["encoderCapacityAv1"])
						g.writeTableRow(&sb, "NVLink Count", gpu["nvlinkCount"])
						sb.WriteString(`            </table>
        </div>
`)
					}
				}
			}
			sb.WriteString(`    </div>
`)
		}
	}

	sb.WriteString(`</div>
`)

	return sb.String()
}

// Helper methods for writing table rows

func (g *GraphGenerator) writeTableRow(sb *strings.Builder, label string, value interface{}) {
	if value == nil {
		return
	}
	// Format the value - show empty strings as "(none)" and include zeros
	displayValue := fmt.Sprintf("%v", value)
	if displayValue == "" {
		displayValue = "(none)"
	}
	sb.WriteString(fmt.Sprintf(`                <tr><td>%s</td><td>%s</td></tr>
`, label, displayValue))
}

func (g *GraphGenerator) writeTableRowBytes(sb *strings.Builder, label string, value interface{}) {
	if value == nil {
		return
	}
	formatted := g.formatBytes(value)
	sb.WriteString(fmt.Sprintf(`                <tr><td>%s</td><td>%s</td></tr>
`, label, formatted))
}

func (g *GraphGenerator) writeTableRowFloat(sb *strings.Builder, label string, value interface{}, unit string) {
	if value == nil {
		return
	}
	var floatVal float64
	switch v := value.(type) {
	case float64:
		floatVal = v
	case int64:
		floatVal = float64(v)
	case int:
		floatVal = float64(v)
	default:
		return
	}
	sb.WriteString(fmt.Sprintf(`                <tr><td>%s</td><td>%.6f%s</td></tr>
`, label, floatVal, unit))
}

func (g *GraphGenerator) writeTableRowMw(sb *strings.Builder, label string, value interface{}) {
	if value == nil {
		return
	}
	var mw float64
	switch v := value.(type) {
	case float64:
		mw = v
	case int64:
		mw = float64(v)
	case int:
		mw = float64(v)
	default:
		return
	}
	if mw == 0 {
		return
	}
	watts := mw / 1000.0
	sb.WriteString(fmt.Sprintf(`                <tr><td>%s</td><td>%.1f W</td></tr>
`, label, watts))
}

// writeGpuTableRow writes a GPU table row with two values (e.g., CUDA capability)
func (g *GraphGenerator) writeGpuTableRow(sb *strings.Builder, label string, val1, val2 interface{}, format string) {
	if val1 == nil || val2 == nil {
		return
	}
	sb.WriteString(fmt.Sprintf(`                <tr><td>%s</td><td>`+format+`</td></tr>
`, label, val1, val2))
}

// writeGpuTableRowUnit writes a GPU table row with a value and unit suffix
func (g *GraphGenerator) writeGpuTableRowUnit(sb *strings.Builder, label string, value interface{}, unit string) {
	if value == nil {
		return
	}
	sb.WriteString(fmt.Sprintf(`                <tr><td>%s</td><td>%v %s</td></tr>
`, label, value, unit))
}

func (g *GraphGenerator) formatBytes(value interface{}) string {
	if value == nil {
		return ""
	}
	var bytes float64
	switch v := value.(type) {
	case float64:
		bytes = v
	case int64:
		bytes = float64(v)
	case int:
		bytes = float64(v)
	default:
		return ""
	}
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	unitIndex := 0
	for bytes >= 1024 && unitIndex < len(units)-1 {
		bytes /= 1024
		unitIndex++
	}
	if unitIndex == 0 {
		return fmt.Sprintf("%.0f %s", bytes, units[unitIndex])
	}
	return fmt.Sprintf("%.2f %s", bytes, units[unitIndex])
}

func (g *GraphGenerator) formatBootTime(value interface{}) string {
	if value == nil {
		return ""
	}
	var bootTime int64
	switch v := value.(type) {
	case float64:
		bootTime = int64(v)
	case int64:
		bootTime = v
	case int:
		bootTime = int64(v)
	default:
		return ""
	}
	if bootTime == 0 {
		return ""
	}
	t := time.Unix(bootTime, 0)
	return t.Format("2006-01-02 15:04:05 MST")
}

// addHistogramCharts adds vLLM histogram charts to the page
func (g *GraphGenerator) addHistogramCharts(page *components.Page, histogramMap map[string]*HistogramSeries) int {
	chartsAdded := 0

	// Sort histogram names for consistent output
	var histNames []string
	for name, series := range histogramMap {
		if len(series.Timestamps) > 0 && len(series.Buckets) > 0 {
			histNames = append(histNames, name)
		}
	}
	sort.Strings(histNames)

	for _, name := range histNames {
		series := histogramMap[name]

		// Create a heatmap showing bucket evolution over time
		heatmap := g.createHistogramHeatmap(series)
		if heatmap != nil {
			page.AddCharts(heatmap)
			chartsAdded++
		}

		// Also create a summary bar chart showing final distribution
		bar := g.createHistogramBarChart(series)
		if bar != nil {
			page.AddCharts(bar)
			chartsAdded++
		}
	}

	return chartsAdded
}

// createHistogramHeatmap creates a heatmap showing bucket counts over time
func (g *GraphGenerator) createHistogramHeatmap(series *HistogramSeries) *charts.HeatMap {
	if len(series.Timestamps) < 2 || len(series.Buckets) == 0 {
		return nil
	}

	heatmap := charts.NewHeatMap()

	title := fmt.Sprintf("vLLM Histogram: %s (Time Evolution)", formatMetricName(series.Name))

	// Sort bucket labels
	var bucketLabels []string
	for label := range series.Buckets {
		bucketLabels = append(bucketLabels, label)
	}
	sortBucketLabels(bucketLabels)

	// Create time labels (sample if too many)
	maxTimePoints := 100
	step := 1
	if len(series.Timestamps) > maxTimePoints {
		step = len(series.Timestamps) / maxTimePoints
	}

	var timeLabels []string
	var sampledIndices []int
	for i := 0; i < len(series.Timestamps); i += step {
		t := time.Unix(0, series.Timestamps[i])
		timeLabels = append(timeLabels, t.Format("15:04:05"))
		sampledIndices = append(sampledIndices, i)
	}

	// Format bucket labels for Y-axis
	yLabels := make([]string, len(bucketLabels))
	for i, label := range bucketLabels {
		if label == "inf" || label == "+Inf" {
			yLabels[i] = "+Inf"
		} else {
			yLabels[i] = label
		}
	}

	heatmap.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: opts.Bool(true),
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(false),
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name:      "Time",
			Type:      "category",
			SplitArea: &opts.SplitArea{Show: opts.Bool(true)},
			AxisLabel: &opts.AxisLabel{
				Rotate: 45,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name:      "Bucket (≤)",
			Type:      "category",
			Data:      yLabels,
			SplitArea: &opts.SplitArea{Show: opts.Bool(true)},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:   "15%",
			Right:  "15%",
			Bottom: "20%",
			Top:    "80",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "400px",
		}),
	)

	// Build heatmap data: [x_index, y_index, value]
	// Value is the per-bucket count (non-cumulative)
	var heatmapData []opts.HeatMapData
	maxVal := float64(0)

	for xi, ti := range sampledIndices {
		var prevCumulative float64 = 0
		for yi, label := range bucketLabels {
			values := series.Buckets[label]
			var cumulative float64 = 0
			if ti < len(values) {
				cumulative = values[ti]
			}

			// Per-bucket count
			bucketCount := cumulative - prevCumulative
			if bucketCount < 0 {
				bucketCount = 0
			}

			if bucketCount > maxVal {
				maxVal = bucketCount
			}

			heatmapData = append(heatmapData, opts.HeatMapData{
				Value: [3]interface{}{xi, yi, bucketCount},
			})

			prevCumulative = cumulative
		}
	}

	// Set visual map with calculated max (Calculable: true adds the interactive slider)
	if maxVal == 0 {
		maxVal = 100
	}
	heatmap.SetGlobalOptions(
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: opts.Bool(true),
			Min:        0,
			Max:        float32(maxVal),
			InRange: &opts.VisualMapInRange{
				Color: []string{"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8", "#ffffbf", "#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026"},
			},
		}),
	)

	heatmap.SetXAxis(timeLabels).
		AddSeries("Count", heatmapData)

	return heatmap
}

// createHistogramBarChart creates a bar chart showing the final histogram distribution
func (g *GraphGenerator) createHistogramBarChart(series *HistogramSeries) *charts.Bar {
	if len(series.Timestamps) == 0 || len(series.Buckets) == 0 {
		return nil
	}

	bar := charts.NewBar()

	// Calculate total observations from the final +Inf bucket
	lastIdx := len(series.Timestamps) - 1
	var totalObs float64 = 0
	if infVals, ok := series.Buckets["+Inf"]; ok && lastIdx < len(infVals) {
		totalObs = infVals[lastIdx]
	} else if infVals, ok := series.Buckets["inf"]; ok && lastIdx < len(infVals) {
		totalObs = infVals[lastIdx]
	}

	title := fmt.Sprintf("vLLM Histogram: %s (Final Distribution)", formatMetricName(series.Name))
	subtitle := fmt.Sprintf("Total observations: %.0f", totalObs)

	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(false),
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Bucket (≤)",
			Type: "category",
			AxisLabel: &opts.AxisLabel{
				Rotate: 45,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Count",
			Type: "value",
		}),
		charts.WithGridOpts(opts.Grid{
			Left:   "10%",
			Right:  "10%",
			Bottom: "20%",
			Top:    "80",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "400px",
		}),
	)

	// Sort bucket labels numerically
	var bucketLabels []string
	for label := range series.Buckets {
		bucketLabels = append(bucketLabels, label)
	}
	sortBucketLabels(bucketLabels)

	// Get the final values for each bucket and compute per-bucket counts
	xLabels := make([]string, len(bucketLabels))
	barData := make([]opts.BarData, len(bucketLabels))

	var prevValue float64 = 0

	for i, label := range bucketLabels {
		// Format label for display
		displayLabel := label
		if label == "inf" || label == "+Inf" {
			displayLabel = "+Inf"
		}
		xLabels[i] = displayLabel

		// Get the cumulative count for this bucket at the final timestamp
		values := series.Buckets[label]
		var cumulative float64 = 0
		if lastIdx < len(values) {
			cumulative = values[lastIdx]
		}

		// Compute per-bucket count (difference from previous bucket)
		bucketCount := cumulative - prevValue
		if bucketCount < 0 {
			bucketCount = 0
		}
		barData[i] = opts.BarData{Value: bucketCount}
		prevValue = cumulative
	}

	bar.SetXAxis(xLabels)
	bar.AddSeries("Count", barData,
		charts.WithBarChartOpts(opts.BarChart{
			BarGap: "10%",
		}),
		charts.WithItemStyleOpts(opts.ItemStyle{
			Color: "#5470c6",
		}),
	)

	return bar
}

// sortBucketLabels sorts histogram bucket labels numerically
func sortBucketLabels(labels []string) {
	sort.Slice(labels, func(i, j int) bool {
		// Handle "inf" specially - always last
		if labels[i] == "inf" || labels[i] == "+Inf" {
			return false
		}
		if labels[j] == "inf" || labels[j] == "+Inf" {
			return true
		}

		// Try to parse as floats
		vi, errI := strconv.ParseFloat(labels[i], 64)
		vj, errJ := strconv.ParseFloat(labels[j], 64)

		if errI == nil && errJ == nil {
			return vi < vj
		}

		// Fallback to string comparison
		return labels[i] < labels[j]
	})
}

// createLineChart creates a line chart for a single metric series
func (g *GraphGenerator) createLineChart(series *MetricSeries, category string, isDelta bool) *charts.Line {
	line := charts.NewLine()

	// Format title
	title := formatMetricName(series.Name)
	var subtitle string
	if isDelta {
		title = title + " (Delta)"
		subtitle = fmt.Sprintf("Category: %s | Rate of Change", category)
	} else {
		title = title + " (Raw)"
		subtitle = fmt.Sprintf("Category: %s | Raw Values", category)
	}

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(false),
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Time",
			Type: "category",
			AxisLabel: &opts.AxisLabel{
				Rotate: 45,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Value",
			Type: "value",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithGridOpts(opts.Grid{
			Left:   "10%",
			Right:  "10%",
			Bottom: "20%",
			Top:    "80",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "450px",
		}),
	)

	var xLabels []string
	var lineData []opts.LineData

	if isDelta {
		// For delta, use timestamps starting from second point
		xLabels = make([]string, len(series.Timestamps)-1)
		for i := 1; i < len(series.Timestamps); i++ {
			ts := series.Timestamps[i]
			t := time.Unix(0, ts)
			xLabels[i-1] = t.Format("15:04:05.000")
		}

		lineData = make([]opts.LineData, len(series.Deltas))
		for i, delta := range series.Deltas {
			lineData[i] = opts.LineData{Value: delta}
		}
	} else {
		// For raw, use all timestamps
		xLabels = make([]string, len(series.Timestamps))
		for i, ts := range series.Timestamps {
			t := time.Unix(0, ts)
			xLabels[i] = t.Format("15:04:05.000")
		}

		lineData = make([]opts.LineData, len(series.Values))
		for i, val := range series.Values {
			lineData[i] = opts.LineData{Value: val}
		}
	}

	line.SetXAxis(xLabels)

	seriesName := "Raw"
	if isDelta {
		seriesName = "Delta"
	}

	lineOpts := []charts.SeriesOpts{
		charts.WithLineChartOpts(opts.LineChart{
			Smooth:     opts.Bool(true),
			ShowSymbol: opts.Bool(true),
		}),
	}

	if isDelta {
		lineOpts = append(lineOpts, charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: opts.Float(0.2),
		}))
	}

	line.AddSeries(seriesName, lineData, lineOpts...)

	return line
}

// formatMetricName converts camelCase to readable format
func formatMetricName(name string) string {
	// Handle common prefixes
	name = strings.TrimPrefix(name, "v")
	name = strings.TrimPrefix(name, "c")

	// Insert spaces before capitals
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return result.String()
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}

	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f, true
		}
		if i, err := val.Int64(); err == nil {
			return float64(i), true
		}
		return 0, false
	default:
		// Try reflection for other numeric types
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rv.Int()), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint()), true
		case reflect.Float32, reflect.Float64:
			return rv.Float(), true
		}
		return 0, false
	}
}

// ExportGraphs generates graphs from the session data
func (e *Exporter) ExportGraphs() error {
	gen := NewGraphGenerator(e.outputDir, e.sessionUUID.String())

	// Try to load static info from the session's JSON file
	staticPath := filepath.Join(e.outputDir, fmt.Sprintf("%s.json", e.sessionUUID))
	if err := gen.LoadStaticInfoFromFile(staticPath); err != nil {
		log.Printf("Note: Could not load static info: %v", err)
	}

	// If streaming mode, try to read from the output file
	if e.streaming {
		ext := e.format
		outputPath := filepath.Join(e.outputDir, fmt.Sprintf("%s.%s", e.sessionUUID, ext))

		if _, err := os.Stat(outputPath); err == nil {
			return gen.GenerateFromFile(outputPath)
		}
		return fmt.Errorf("streaming output file not found: %s", outputPath)
	}

	// Batch mode - use snapshot files
	if len(e.snapshotFiles) < 2 {
		return fmt.Errorf("need at least 2 snapshots to generate graphs")
	}

	return gen.GenerateFromSnapshotFiles(e.snapshotFiles)
}

// GenerateGraphsFromOutputFile creates graphs from a completed output file
// Useful for generating graphs from previously collected data
func GenerateGraphsFromOutputFile(outputPath string) error {
	dir := filepath.Dir(outputPath)
	base := filepath.Base(outputPath)
	ext := filepath.Ext(base)
	sessionUUID := strings.TrimSuffix(base, ext)

	gen := NewGraphGenerator(dir, sessionUUID)

	// Try to load static info from the session's JSON file
	staticPath := filepath.Join(dir, fmt.Sprintf("%s.json", sessionUUID))
	if err := gen.LoadStaticInfoFromFile(staticPath); err != nil {
		log.Printf("Note: Could not load static info: %v", err)
	}

	return gen.GenerateFromFile(outputPath)
}
