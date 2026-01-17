package graphing

import (
	"fmt"
	"html/template"
	"time"
)

// HTML templates for report generation
var templates = template.Must(template.New("").Funcs(templateFuncs).Parse(`
{{define "page"}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    {{template "styles" .}}
    {{template "scripts" .}}
</head>
<body>
    {{template "static_info" .StaticInfo}}
    {{.ChartContent}}
</body>
</html>
{{end}}

{{define "styles"}}
<style>
* {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
}
body {
    max-width: 1400px;
    margin: 0 auto;
    padding: 20px;
}
.static-info-container {
    margin-bottom: 20px;
}
.static-info-header {
    border-bottom: 2px solid #333;
    padding-bottom: 10px;
    margin-bottom: 15px;
}
.static-info-header h1 {
    margin: 0;
    font-size: 18px;
}
.session-id {
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
}
.info-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
}
.info-table td {
    padding: 3px 8px;
    border-bottom: 1px solid #eee;
}
.info-table td:first-child {
    width: 150px;
    color: #666;
}
.info-table td:last-child {
    font-family: monospace;
    font-size: 11px;
    word-break: break-all;
}
.info-table tr:last-child td {
    border-bottom: none;
}
.gpu-section {
    background: #fff;
    border: 1px solid #ddd;
    padding: 10px;
    margin-top: 10px;
}
.container {
    display: block !important;
    margin: 0 0 10px 0 !important;
    padding: 15px !important;
    background: #f5f5f5 !important;
    border: 1px solid #ddd !important;
    overflow: hidden !important;
}
.item {
    margin: 0 !important;
}
</style>
{{end}}

{{define "scripts"}}
<script>
window.addEventListener('resize', function() {
    document.querySelectorAll('[_echarts_instance_]').forEach(function(el) {
        var c = echarts.getInstanceByDom(el);
        if (c) c.resize();
    });
});
window.addEventListener('load', function() {
    setTimeout(function() {
        document.querySelectorAll('[_echarts_instance_]').forEach(function(el) {
            var c = echarts.getInstanceByDom(el);
            if (c) c.resize();
        });
    }, 100);
});
</script>
{{end}}

{{define "static_info"}}
{{if .}}
<div class="static-info-container">
    <div class="static-info-header">
        <h1>Profiler Report</h1>
        <div class="session-id">Session: {{.SessionID}}</div>
    </div>

    {{template "system_section" .}}
    {{template "memory_section" .}}
    {{template "container_section" .}}
    {{template "network_section" .}}
    {{template "disk_section" .}}
    {{template "gpu_section" .}}
</div>
{{end}}
{{end}}

{{define "system_section"}}
<div class="info-section">
    <h3>System</h3>
    <table class="info-table">
        {{template "row" dict "Label" "UUID" "Value" .UUID}}
        {{template "row" dict "Label" "VM ID" "Value" .VMID}}
        {{template "row" dict "Label" "Hostname" "Value" .Hostname}}
        {{template "row" dict "Label" "Boot Time" "Value" (.BootTime | formatTime)}}
        {{template "row" dict "Label" "Processors" "Value" .NumProcessors}}
        {{template "row" dict "Label" "CPU" "Value" .CPUType}}
        {{template "row" dict "Label" "CPU Cache" "Value" .CPUCache}}
        {{template "row" dict "Label" "Kernel" "Value" .KernelInfo}}
        {{template "row" dict "Label" "Time Synced" "Value" .TimeSynced}}
        {{template "row_float" dict "Label" "Time Offset" "Value" .TimeOffset "Format" "%.6f sec"}}
        {{template "row_float" dict "Label" "Time Max Error" "Value" .TimeMaxError "Format" "%.6f sec"}}
    </table>
</div>
{{end}}

{{define "memory_section"}}
<div class="info-section">
    <h3>Memory</h3>
    <table class="info-table">
        {{template "row_bytes" dict "Label" "Total" "Value" .MemoryTotal}}
        {{template "row_bytes" dict "Label" "Swap" "Value" .SwapTotal}}
    </table>
</div>
{{end}}

{{define "container_section"}}
{{if .ContainerID}}
<div class="info-section">
    <h3>Container</h3>
    <table class="info-table">
        {{template "row" dict "Label" "ID" "Value" .ContainerID}}
        {{template "row" dict "Label" "CPUs" "Value" .ContainerCPUs}}
        {{template "row" dict "Label" "Cgroup Version" "Value" .CgroupVersion}}
    </table>
</div>
{{end}}
{{end}}

{{define "network_section"}}
{{if .NetworkInterfaces}}
<div class="info-section">
    <h3>Network Interfaces</h3>
    {{range .NetworkInterfaces}}
    <table class="info-table">
        {{template "row" dict "Label" "Name" "Value" .Name}}
        {{template "row" dict "Label" "MAC" "Value" .MAC}}
        {{template "row" dict "Label" "State" "Value" .State}}
        {{template "row" dict "Label" "MTU" "Value" .MTU}}
        {{if .SpeedMbps}}{{template "row" dict "Label" "Speed (Mbps)" "Value" .SpeedMbps}}{{end}}
    </table>
    {{end}}
</div>
{{end}}
{{end}}

{{define "disk_section"}}
{{if .Disks}}
<div class="info-section">
    <h3>Disks</h3>
    {{range .Disks}}
    <table class="info-table">
        {{template "row" dict "Label" "Name" "Value" .Name}}
        {{template "row" dict "Label" "Model" "Value" .Model}}
        {{template "row" dict "Label" "Vendor" "Value" .Vendor}}
        {{template "row_bytes" dict "Label" "Size" "Value" .SizeBytes}}
    </table>
    {{end}}
</div>
{{end}}
{{end}}

{{define "gpu_section"}}
{{if .GPUCount}}
<div class="info-section">
    <h3>NVIDIA GPU</h3>
    <table class="info-table">
        {{template "row" dict "Label" "GPU Count" "Value" .GPUCount}}
        {{template "row" dict "Label" "Driver Version" "Value" .DriverVersion}}
        {{template "row" dict "Label" "CUDA Version" "Value" .CUDAVersion}}
        {{template "row" dict "Label" "NVML Version" "Value" .NVMLVersion}}
    </table>
    {{range $i, $gpu := .GPUs}}
    <div class="gpu-section">
        <strong>GPU {{$i}}: {{$gpu.Name}}</strong>
        <table class="info-table">
            {{template "row" dict "Label" "Index" "Value" $gpu.Index}}
            {{template "row" dict "Label" "UUID" "Value" $gpu.UUID}}
            {{template "row" dict "Label" "Serial" "Value" $gpu.Serial}}
            {{template "row" dict "Label" "Board Part Number" "Value" $gpu.BoardPartNumber}}
            {{template "row" dict "Label" "Brand" "Value" $gpu.Brand}}
            {{template "row" dict "Label" "Architecture" "Value" $gpu.Architecture}}
            {{if $gpu.CUDACapabilityMajor}}
            <tr><td>CUDA Capability</td><td>{{$gpu.CUDACapabilityMajor}}.{{$gpu.CUDACapabilityMinor}}</td></tr>
            {{end}}
            {{template "row_bytes" dict "Label" "Memory Total" "Value" $gpu.MemoryTotalBytes}}
            {{template "row_bytes" dict "Label" "BAR1 Total" "Value" $gpu.Bar1TotalBytes}}
            {{template "row_unit" dict "Label" "Memory Bus Width" "Value" $gpu.MemoryBusWidthBits "Unit" "bits"}}
            {{template "row" dict "Label" "CUDA Cores" "Value" $gpu.NumCores}}
            {{template "row_unit" dict "Label" "Max Clock Graphics" "Value" $gpu.MaxClockGraphicsMhz "Unit" "MHz"}}
            {{template "row_unit" dict "Label" "Max Clock Memory" "Value" $gpu.MaxClockMemoryMhz "Unit" "MHz"}}
            {{template "row_unit" dict "Label" "Max Clock SM" "Value" $gpu.MaxClockSmMhz "Unit" "MHz"}}
            {{template "row_unit" dict "Label" "Max Clock Video" "Value" $gpu.MaxClockVideoMhz "Unit" "MHz"}}
            {{template "row" dict "Label" "PCI Bus ID" "Value" $gpu.PCIBusID}}
            {{template "row" dict "Label" "PCIe Max Link Gen" "Value" $gpu.PCIeMaxLinkGen}}
            {{template "row" dict "Label" "PCIe Max Link Width" "Value" $gpu.PCIeMaxLinkWidth}}
            {{template "row_watts" dict "Label" "Power Default Limit" "Value" $gpu.PowerDefaultLimitMw}}
            {{template "row_watts" dict "Label" "Power Min Limit" "Value" $gpu.PowerMinLimitMw}}
            {{template "row_watts" dict "Label" "Power Max Limit" "Value" $gpu.PowerMaxLimitMw}}
            {{template "row" dict "Label" "VBIOS Version" "Value" $gpu.VBIOSVersion}}
            {{template "row" dict "Label" "Num Fans" "Value" $gpu.NumFans}}
            {{template "row_unit" dict "Label" "Temp Shutdown" "Value" $gpu.TempShutdownC "Unit" "°C"}}
            {{template "row_unit" dict "Label" "Temp Slowdown" "Value" $gpu.TempSlowdownC "Unit" "°C"}}
            {{template "row_unit" dict "Label" "Temp Max Operating" "Value" $gpu.TempMaxOperatingC "Unit" "°C"}}
            {{template "row" dict "Label" "ECC Mode" "Value" $gpu.ECCModeEnabled}}
            {{template "row" dict "Label" "Persistence Mode" "Value" $gpu.PersistenceModeOn}}
            {{template "row" dict "Label" "Compute Mode" "Value" $gpu.ComputeMode}}
            {{template "row" dict "Label" "MIG Mode" "Value" $gpu.MIGModeEnabled}}
            {{template "row" dict "Label" "NVLink Count" "Value" $gpu.NVLinkCount}}
        </table>
    </div>
    {{end}}
</div>
{{end}}
{{end}}

{{define "row"}}
{{if and .Value (ne (printf "%v" .Value) "") (ne (printf "%v" .Value) "<nil>") (ne (printf "%v" .Value) "0")}}
<tr><td>{{.Label}}</td><td>{{.Value}}</td></tr>
{{end}}
{{end}}

{{define "row_unit"}}
{{if and .Value (ne (printf "%v" .Value) "0")}}
<tr><td>{{.Label}}</td><td>{{.Value}} {{.Unit}}</td></tr>
{{end}}
{{end}}

{{define "row_float"}}
{{if .Value}}
<tr><td>{{.Label}}</td><td>{{printf .Format .Value}}</td></tr>
{{end}}
{{end}}

{{define "row_bytes"}}
{{if .Value}}
<tr><td>{{.Label}}</td><td>{{.Value | formatBytes}}</td></tr>
{{end}}
{{end}}

{{define "row_watts"}}
{{if and .Value (ne (printf "%v" .Value) "0")}}
<tr><td>{{.Label}}</td><td>{{.Value | formatWatts}}</td></tr>
{{end}}
{{end}}
`))

// Template helper functions
var templateFuncs = template.FuncMap{
	"dict":        dictFunc,
	"formatBytes": formatBytesFunc,
	"formatWatts": formatWattsFunc,
	"formatTime":  formatTimeFunc,
}

// dictFunc creates a map from key-value pairs for template use.
func dictFunc(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		return nil
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			continue
		}
		dict[key] = values[i+1]
	}
	return dict
}

// formatBytesFunc formats bytes into human-readable format.
func formatBytesFunc(v interface{}) string {
	bytes := toFloat64(v)
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for bytes >= 1024 && i < len(units)-1 {
		bytes /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", bytes, units[i])
}

// formatWattsFunc formats milliwatts into watts.
func formatWattsFunc(v interface{}) string {
	mw := toFloat64(v)
	if mw == 0 {
		return ""
	}
	return fmt.Sprintf("%.0f W", mw/1000)
}

// formatTimeFunc formats a Unix timestamp.
func formatTimeFunc(v interface{}) string {
	ts := int64(toFloat64(v))
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

// toFloat64 safely converts interface to float64.
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	case uint64:
		return float64(n)
	}
	return 0
}
