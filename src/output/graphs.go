package output

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"InferenceProfiler/src/collectors"
)

//go:embed templates/report.html
var reportTemplate string

// ReportData holds all data needed to render the HTML report
type ReportData struct {
	SessionUUID  string
	Hostname     string
	Duration     string
	SampleCount  int
	StaticInfo   *StaticInfoData
	Summary      []SummaryItem
	SystemCharts []ChartConfig
	GPUCharts    []GPUChartGroup
	ChartDataScript template.JS
}

// StaticInfoData organizes static info into categories
type StaticInfoData struct {
	System []InfoRow
	CPU    []InfoRow
	Memory []InfoRow
	GPU    []GPUInfo
}

// GPUInfo holds info for a single GPU
type GPUInfo struct {
	Name string
	Info []InfoRow
}

// InfoRow represents a label-value pair for display
type InfoRow struct {
	Label string
	Value string
}

// SummaryItem is a summary statistic
type SummaryItem struct {
	Label string
	Value string
}

// ChartConfig defines a chart to render
type ChartConfig struct {
	ID    string
	Title string
}

// GPUChartGroup holds charts for a single GPU
type GPUChartGroup struct {
	Name   string
	Charts []ChartConfig
}

// GenerateHTML creates an HTML report from profiler data
func GenerateHTML(outputPath string, sessionUUID string, staticMetrics *collectors.StaticMetrics, records []map[string]interface{}) error {
	if len(records) == 0 {
		return fmt.Errorf("no records to generate report from")
	}

	// Build report data
	data := &ReportData{
		SessionUUID: sessionUUID,
		SampleCount: len(records),
	}

	// Extract hostname and timing info
	if staticMetrics != nil {
		data.Hostname = staticMetrics.Hostname
	}

	// Calculate duration from timestamps
	if len(records) > 1 {
		first := getFloat64(records[0], "timestamp")
		last := getFloat64(records[len(records)-1], "timestamp")
		durationNs := last - first
		data.Duration = formatDuration(time.Duration(durationNs))
	}

	// Build static info sections
	if staticMetrics != nil {
		data.StaticInfo = buildStaticInfo(staticMetrics)
	}

	// Build summary statistics
	data.Summary = buildSummary(records)

	// Build system charts config
	data.SystemCharts = []ChartConfig{
		{ID: "chart-cpu-usage", Title: "CPU Usage"},
		{ID: "chart-memory", Title: "Memory Usage"},
		{ID: "chart-disk-io", Title: "Disk I/O"},
		{ID: "chart-network", Title: "Network Traffic"},
	}

	// Check for GPU data and build GPU charts
	numGPUs := detectNumGPUs(records)
	for i := 0; i < numGPUs; i++ {
		gpuName := fmt.Sprintf("GPU %d", i)
		if staticMetrics != nil && len(staticMetrics.NvidiaGPUsJSON) > 0 {
			var gpus []map[string]interface{}
			if err := json.Unmarshal([]byte(staticMetrics.NvidiaGPUsJSON), &gpus); err == nil {
				if i < len(gpus) {
					if name, ok := gpus[i]["name"].(string); ok {
						gpuName = name
					}
				}
			}
		}

		data.GPUCharts = append(data.GPUCharts, GPUChartGroup{
			Name: gpuName,
			Charts: []ChartConfig{
				{ID: fmt.Sprintf("chart-gpu%d-util", i), Title: "GPU Utilization"},
				{ID: fmt.Sprintf("chart-gpu%d-memory", i), Title: "GPU Memory"},
				{ID: fmt.Sprintf("chart-gpu%d-power", i), Title: "Power & Temperature"},
				{ID: fmt.Sprintf("chart-gpu%d-clocks", i), Title: "Clock Frequencies"},
			},
		})
	}

	// Generate chart data JavaScript
	data.ChartDataScript = template.JS(generateChartDataScript(records, numGPUs))

	// Parse and execute template
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

// GenerateHTMLFromFile generates an HTML report from an existing data file
func GenerateHTMLFromFile(dataPath string, outputDir string) (string, error) {
	// Load data
	loader, err := NewRecordLoader(dataPath)
	if err != nil {
		return "", err
	}

	records, err := loader.Load(dataPath)
	if err != nil {
		return "", fmt.Errorf("failed to load data: %w", err)
	}

	// Determine session UUID from filename
	sessionUUID := extractSessionUUID(dataPath)

	// Try to load static metrics
	var staticMetrics *collectors.StaticMetrics
	staticPath := filepath.Join(filepath.Dir(dataPath), fmt.Sprintf("static_%s.json", sessionUUID))
	if data, err := os.ReadFile(staticPath); err == nil {
		staticMetrics = &collectors.StaticMetrics{}
		json.Unmarshal(data, staticMetrics)
	}

	// Generate output path
	outputPath := filepath.Join(outputDir, fmt.Sprintf("report_%s.html", sessionUUID))

	if err := GenerateHTML(outputPath, sessionUUID, staticMetrics, records); err != nil {
		return "", err
	}

	return outputPath, nil
}

// buildStaticInfo organizes static metrics into display categories
func buildStaticInfo(m *collectors.StaticMetrics) *StaticInfoData {
	info := &StaticInfoData{}

	// System info
	info.System = []InfoRow{
		{Label: "Hostname", Value: m.Hostname},
		{Label: "OS", Value: fmt.Sprintf("%s %s", m.OSName, m.OSVersion)},
		{Label: "Kernel", Value: m.KernelVersion},
		{Label: "Architecture", Value: m.Architecture},
		{Label: "VM ID", Value: m.VMID},
	}
	if m.ContainerID != "" {
		info.System = append(info.System, InfoRow{Label: "Container ID", Value: truncate(m.ContainerID, 12)})
	}

	// CPU info
	info.CPU = []InfoRow{
		{Label: "Model", Value: m.CPUType},
		{Label: "Cores", Value: fmt.Sprintf("%d", m.CPUCount)},
		{Label: "Cache", Value: m.CPUCache},
	}

	// Memory info
	info.Memory = []InfoRow{
		{Label: "Total RAM", Value: formatBytes(m.TotalMemory)},
		{Label: "Total Swap", Value: formatBytes(m.TotalSwap)},
	}

	// GPU info
	if m.NvidiaGPUsJSON != "" {
		var gpus []map[string]interface{}
		if err := json.Unmarshal([]byte(m.NvidiaGPUsJSON), &gpus); err == nil {
			for _, gpu := range gpus {
				gpuInfo := GPUInfo{
					Name: getString(gpu, "name"),
				}

				gpuInfo.Info = []InfoRow{
					{Label: "UUID", Value: truncate(getString(gpu, "uuid"), 20)},
					{Label: "Memory", Value: formatBytes(int64(getFloat64Val(gpu, "memoryTotal")))},
					{Label: "Architecture", Value: getString(gpu, "architecture")},
					{Label: "Compute Capability", Value: fmt.Sprintf("%d.%d", 
						int(getFloat64Val(gpu, "cudaComputeCapabilityMajor")),
						int(getFloat64Val(gpu, "cudaComputeCapabilityMinor")))},
					{Label: "Driver Version", Value: getString(gpu, "driverVersion")},
					{Label: "CUDA Version", Value: getString(gpu, "cudaVersion")},
				}

				info.GPU = append(info.GPU, gpuInfo)
			}
		}
	}

	return info
}

// buildSummary creates summary statistics from records
func buildSummary(records []map[string]interface{}) []SummaryItem {
	if len(records) == 0 {
		return nil
	}

	var items []SummaryItem

	// Average CPU usage
	var totalCPU float64
	for _, r := range records {
		totalCPU += getFloat64(r, "vMemoryPercent")
	}
	items = append(items, SummaryItem{
		Label: "Avg Memory Usage",
		Value: fmt.Sprintf("%.1f%%", totalCPU/float64(len(records))),
	})

	// Peak memory
	var maxMem int64
	for _, r := range records {
		if mem := int64(getFloat64(r, "vMemoryUsed")); mem > maxMem {
			maxMem = mem
		}
	}
	items = append(items, SummaryItem{
		Label: "Peak Memory",
		Value: formatBytes(maxMem),
	})

	// Check for GPU metrics
	if getFloat64(records[0], "gpuUtilizationGpu") > 0 || hasKey(records[0], "gpuUtilizationGpu") {
		var maxGPU float64
		for _, r := range records {
			if util := getFloat64(r, "gpuUtilizationGpu"); util > maxGPU {
				maxGPU = util
			}
		}
		items = append(items, SummaryItem{
			Label: "Peak GPU Utilization",
			Value: fmt.Sprintf("%.0f%%", maxGPU),
		})
	}

	return items
}

// generateChartDataScript creates JavaScript for Plotly charts
func generateChartDataScript(records []map[string]interface{}, numGPUs int) string {
	var js strings.Builder

	// Extract timestamps
	timestamps := make([]float64, len(records))
	for i, r := range records {
		timestamps[i] = getFloat64(r, "timestamp")
	}

	// Convert to ISO format for Plotly
	js.WriteString("const timestamps = [")
	for i, ts := range timestamps {
		if i > 0 {
			js.WriteString(",")
		}
		t := time.Unix(0, int64(ts))
		js.WriteString(fmt.Sprintf("'%s'", t.Format(time.RFC3339)))
	}
	js.WriteString("];\n\n")

	// CPU Usage chart
	js.WriteString(generateLineChart("chart-cpu-usage", records, timestamps,
		[]seriesConfig{
			{key: "vCpuTimeUserMode", label: "User", color: 0, diff: true},
			{key: "vCpuTimeKernelMode", label: "System", color: 1, diff: true},
			{key: "vCpuTimeIOWait", label: "I/O Wait", color: 2, diff: true},
		}, "CPU Time (delta)", false))

	// Memory chart
	js.WriteString(generateLineChart("chart-memory", records, timestamps,
		[]seriesConfig{
			{key: "vMemoryUsed", label: "Used", color: 0, scale: 1e-9},
			{key: "vMemoryCached", label: "Cached", color: 1, scale: 1e-9},
			{key: "vMemoryBuffers", label: "Buffers", color: 2, scale: 1e-9},
		}, "Memory (GB)", false))

	// Disk I/O chart
	js.WriteString(generateLineChart("chart-disk-io", records, timestamps,
		[]seriesConfig{
			{key: "vDiskReadBytes", label: "Read", color: 0, diff: true, scale: 1e-6},
			{key: "vDiskWriteBytes", label: "Write", color: 1, diff: true, scale: 1e-6},
		}, "Throughput (MB/s)", false))

	// Network chart
	js.WriteString(generateLineChart("chart-network", records, timestamps,
		[]seriesConfig{
			{key: "vNetworkBytesRecvd", label: "Received", color: 0, diff: true, scale: 1e-6},
			{key: "vNetworkBytesSent", label: "Sent", color: 1, diff: true, scale: 1e-6},
		}, "Throughput (MB/s)", false))

	// GPU charts
	for i := 0; i < numGPUs; i++ {
		prefix := "gpu"
		if numGPUs > 1 {
			prefix = fmt.Sprintf("gpu%d", i)
		}

		// GPU Utilization
		js.WriteString(generateLineChart(fmt.Sprintf("chart-gpu%d-util", i), records, timestamps,
			[]seriesConfig{
				{key: prefix + "UtilizationGpu", label: "GPU", color: 0},
				{key: prefix + "UtilizationMemory", label: "Memory", color: 1},
			}, "Utilization (%)", false))

		// GPU Memory
		js.WriteString(generateLineChart(fmt.Sprintf("chart-gpu%d-memory", i), records, timestamps,
			[]seriesConfig{
				{key: prefix + "MemoryUsed", label: "Used", color: 0, scale: 1e-9},
				{key: prefix + "MemoryFree", label: "Free", color: 1, scale: 1e-9},
			}, "Memory (GB)", false))

		// Power & Temperature
		js.WriteString(generateLineChart(fmt.Sprintf("chart-gpu%d-power", i), records, timestamps,
			[]seriesConfig{
				{key: prefix + "PowerDraw", label: "Power (W)", color: 0, scale: 0.001},
				{key: prefix + "Temperature", label: "Temp (°C)", color: 2, yaxis: "y2"},
			}, "Power (W)", true))

		// Clock frequencies
		js.WriteString(generateLineChart(fmt.Sprintf("chart-gpu%d-clocks", i), records, timestamps,
			[]seriesConfig{
				{key: prefix + "ClocksSM", label: "SM", color: 0},
				{key: prefix + "ClocksMemory", label: "Memory", color: 1},
			}, "Frequency (MHz)", false))
	}

	return js.String()
}

type seriesConfig struct {
	key    string
	label  string
	color  int
	diff   bool    // calculate delta between samples
	scale  float64 // multiply values by this
	yaxis  string  // for dual-axis charts
}

func generateLineChart(id string, records []map[string]interface{}, timestamps []float64, series []seriesConfig, yLabel string, dualAxis bool) string {
	var js strings.Builder
	
	js.WriteString(fmt.Sprintf("// %s\n", id))
	js.WriteString("{\n")
	js.WriteString("  const traces = [\n")

	colors := []string{"#e94560", "#4ecdc4", "#ffeaa7", "#74b9ff", "#ff6b6b", "#45b7d1"}

	for i, s := range series {
		if i > 0 {
			js.WriteString(",\n")
		}

		// Extract values
		values := make([]float64, len(records))
		for j, r := range records {
			v := getFloat64(r, s.key)
			if s.scale != 0 {
				v *= s.scale
			}
			values[j] = v
		}

		// Calculate deltas if needed
		if s.diff && len(values) > 1 {
			deltas := make([]float64, len(values))
			for j := 1; j < len(values); j++ {
				dt := (timestamps[j] - timestamps[j-1]) / 1e9 // to seconds
				if dt > 0 {
					deltas[j] = (values[j] - values[j-1]) / dt
				}
			}
			values = deltas
		}

		// Generate trace
		colorIdx := s.color % len(colors)
		js.WriteString(fmt.Sprintf("    {x: timestamps, y: %s, name: '%s', type: 'scatter', mode: 'lines', line: {color: '%s'}",
			formatJSArray(values), s.label, colors[colorIdx]))
		
		if s.yaxis != "" {
			js.WriteString(fmt.Sprintf(", yaxis: '%s'", s.yaxis))
		}
		js.WriteString("}")
	}

	js.WriteString("\n  ];\n")

	// Layout
	js.WriteString("  const chartLayout = {...layout, ")
	js.WriteString(fmt.Sprintf("yaxis: {...layout.yaxis, title: '%s'}", yLabel))
	
	if dualAxis {
		js.WriteString(", yaxis2: {title: 'Temperature (°C)', overlaying: 'y', side: 'right', gridcolor: 'rgba(255,255,255,0.05)'}")
	}
	js.WriteString("};\n")

	js.WriteString(fmt.Sprintf("  Plotly.newPlot('%s', traces, chartLayout, config);\n", id))
	js.WriteString("}\n\n")

	return js.String()
}

func formatJSArray(values []float64) string {
	parts := make([]string, len(values))
	for i, v := range values {
		if v != v { // NaN check
			parts[i] = "null"
		} else {
			parts[i] = fmt.Sprintf("%.4f", v)
		}
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// Helper functions

func detectNumGPUs(records []map[string]interface{}) int {
	if len(records) == 0 {
		return 0
	}
	
	// Check for gpuIndex field
	if hasKey(records[0], "gpuIndex") {
		maxIdx := 0
		for _, r := range records {
			if idx := int(getFloat64(r, "gpuIndex")); idx > maxIdx {
				maxIdx = idx
			}
		}
		return maxIdx + 1
	}
	
	// Check for GPU fields without index
	if hasKey(records[0], "gpuUtilizationGpu") {
		return 1
	}
	
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int64:
			return float64(val)
		case int:
			return float64(val)
		}
	}
	return 0
}

func getFloat64Val(m map[string]interface{}, key string) float64 {
	return getFloat64(m, key)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func hasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func extractSessionUUID(path string) string {
	base := filepath.Base(path)
	// Try to extract UUID from filename like "metrics_UUID.jsonl"
	parts := strings.Split(base, "_")
	if len(parts) >= 2 {
		uuid := strings.TrimSuffix(parts[1], filepath.Ext(parts[1]))
		return uuid
	}
	return "unknown"
}

// GetAvailableMetrics returns a sorted list of all metric keys in the records
func GetAvailableMetrics(records []map[string]interface{}) []string {
	keySet := make(map[string]struct{})
	for _, r := range records {
		for k := range r {
			keySet[k] = struct{}{}
		}
	}
	
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
