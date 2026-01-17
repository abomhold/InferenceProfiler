package graphing

import (
	"InferenceProfiler/pkg/formatting"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
)

// StaticInfoData holds parsed static information for template rendering.
type StaticInfoData struct {
	SessionID string

	// System
	UUID          string
	VMID          string
	Hostname      string
	BootTime      int64
	NumProcessors int
	CPUType       string
	CPUCache      string
	KernelInfo    string
	TimeSynced    bool
	TimeOffset    float64
	TimeMaxError  float64

	// Memory
	MemoryTotal int64
	SwapTotal   int64

	// Container
	ContainerID   string
	ContainerCPUs int64
	CgroupVersion int64

	// Network
	NetworkInterfaces []NetworkInterfaceData

	// Disks
	Disks []DiskData

	// GPU
	GPUCount      int
	DriverVersion string
	CUDAVersion   string
	NVMLVersion   string
	GPUs          []GPUData
}

// NetworkInterfaceData holds network interface information.
type NetworkInterfaceData struct {
	Name      string
	MAC       string
	State     string
	MTU       int64
	SpeedMbps int64
}

// DiskData holds disk information.
type DiskData struct {
	Name      string
	Model     string
	Vendor    string
	SizeBytes int64
}

// GPUData holds GPU information.
type GPUData struct {
	Index               int
	Name                string
	UUID                string
	Serial              string
	BoardPartNumber     string
	Brand               string
	Architecture        string
	CUDACapabilityMajor int
	CUDACapabilityMinor int
	MemoryTotalBytes    int64
	Bar1TotalBytes      int64
	MemoryBusWidthBits  int
	NumCores            int
	MaxClockGraphicsMhz int
	MaxClockMemoryMhz   int
	MaxClockSmMhz       int
	MaxClockVideoMhz    int
	PCIBusID            string
	PCIeMaxLinkGen      int
	PCIeMaxLinkWidth    int
	PowerDefaultLimitMw int
	PowerMinLimitMw     int
	PowerMaxLimitMw     int
	VBIOSVersion        string
	NumFans             int
	TempShutdownC       int
	TempSlowdownC       int
	TempMaxOperatingC   int
	ECCModeEnabled      bool
	PersistenceModeOn   bool
	ComputeMode         string
	MIGModeEnabled      bool
	NVLinkCount         int
}

// PageData holds all data for the full HTML page.
type PageData struct {
	Title        string
	StaticInfo   *StaticInfoData
	ChartContent template.HTML
}

// renderStaticInfoHTML renders static info using templates.
func renderStaticInfoHTML(sessionID string, info map[string]interface{}) (string, error) {
	data := parseStaticInfo(sessionID, info)

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "static_info", data); err != nil {
		return "", fmt.Errorf("failed to execute static_info template: %w", err)
	}

	return buf.String(), nil
}

// renderFullPage renders a complete HTML page with charts.
func renderFullPage(title string, staticInfo *StaticInfoData, chartHTML string) (string, error) {
	data := PageData{
		Title:        title,
		StaticInfo:   staticInfo,
		ChartContent: template.HTML(chartHTML),
	}

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "page", data); err != nil {
		return "", fmt.Errorf("failed to execute page template: %w", err)
	}

	return buf.String(), nil
}

// renderStylesAndScripts returns the CSS and JavaScript as a string.
func renderStylesAndScripts() (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "styles", nil); err != nil {
		return "", err
	}
	if err := templates.ExecuteTemplate(&buf, "scripts", nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// parseStaticInfo converts raw map data into structured StaticInfoData.
func parseStaticInfo(sessionID string, info map[string]interface{}) *StaticInfoData {
	if info == nil {
		return nil
	}

	data := &StaticInfoData{
		SessionID: sessionID,
	}

	// System info
	data.UUID = getString(info, "uuid")
	data.VMID = getString(info, "vId")
	data.Hostname = getString(info, "vHostname")
	data.BootTime = getInt64(info, "vBootTime")
	data.NumProcessors = int(getInt64(info, "vNumProcessors"))
	data.CPUType = getString(info, "vCpuType")
	data.CPUCache = getString(info, "vCpuCache")
	data.KernelInfo = getString(info, "vKernelInfo")
	data.TimeSynced = getBool(info, "vTimeSynced")
	data.TimeOffset = getFloat64(info, "vTimeOffsetSeconds")
	data.TimeMaxError = getFloat64(info, "vTimeMaxErrorSeconds")

	// Memory
	data.MemoryTotal = getInt64(info, "vMemoryTotalBytes")
	data.SwapTotal = getInt64(info, "vSwapTotalBytes")

	// Container
	data.ContainerID = getString(info, "cId")
	data.ContainerCPUs = getInt64(info, "cNumProcessors")
	data.CgroupVersion = getInt64(info, "cCgroupVersion")

	// Network interfaces
	if netJSON := getString(info, "networkInterfaces"); netJSON != "" {
		var ifaces []map[string]interface{}
		if json.Unmarshal([]byte(netJSON), &ifaces) == nil {
			for _, iface := range ifaces {
				data.NetworkInterfaces = append(data.NetworkInterfaces, NetworkInterfaceData{
					Name:      getString(iface, "name"),
					MAC:       getString(iface, "mac"),
					State:     getString(iface, "state"),
					MTU:       getInt64(iface, "mtu"),
					SpeedMbps: getInt64(iface, "speedMbps"),
				})
			}
		}
	}

	// Disks
	if diskJSON := getString(info, "disks"); diskJSON != "" {
		var disks []map[string]interface{}
		if json.Unmarshal([]byte(diskJSON), &disks) == nil {
			for _, disk := range disks {
				data.Disks = append(data.Disks, DiskData{
					Name:      getString(disk, "name"),
					Model:     getString(disk, "model"),
					Vendor:    getString(disk, "vendor"),
					SizeBytes: getInt64(disk, "sizeBytes"),
				})
			}
		}
	}

	// GPU
	data.GPUCount = int(getInt64(info, "nvidiaGpuCount"))
	data.DriverVersion = getString(info, "nvidiaDriverVersion")
	data.CUDAVersion = getString(info, "nvidiaCudaVersion")
	data.NVMLVersion = getString(info, "nvmlVersion")

	if gpuJSON := getString(info, "nvidiaGpus"); gpuJSON != "" {
		var gpus []map[string]interface{}
		if json.Unmarshal([]byte(gpuJSON), &gpus) == nil {
			for _, gpu := range gpus {
				data.GPUs = append(data.GPUs, GPUData{
					Index:               int(getInt64(gpu, "index")),
					Name:                getString(gpu, "name"),
					UUID:                getString(gpu, "uuid"),
					Serial:              getString(gpu, "serial"),
					BoardPartNumber:     getString(gpu, "boardPartNumber"),
					Brand:               getString(gpu, "brand"),
					Architecture:        getString(gpu, "architecture"),
					CUDACapabilityMajor: int(getInt64(gpu, "cudaCapabilityMajor")),
					CUDACapabilityMinor: int(getInt64(gpu, "cudaCapabilityMinor")),
					MemoryTotalBytes:    getInt64(gpu, "memoryTotalBytes"),
					Bar1TotalBytes:      getInt64(gpu, "bar1TotalBytes"),
					MemoryBusWidthBits:  int(getInt64(gpu, "memoryBusWidthBits")),
					NumCores:            int(getInt64(gpu, "numCores")),
					MaxClockGraphicsMhz: int(getInt64(gpu, "maxClockGraphicsMhz")),
					MaxClockMemoryMhz:   int(getInt64(gpu, "maxClockMemoryMhz")),
					MaxClockSmMhz:       int(getInt64(gpu, "maxClockSmMhz")),
					MaxClockVideoMhz:    int(getInt64(gpu, "maxClockVideoMhz")),
					PCIBusID:            getString(gpu, "pciBusId"),
					PCIeMaxLinkGen:      int(getInt64(gpu, "pcieMaxLinkGen")),
					PCIeMaxLinkWidth:    int(getInt64(gpu, "pcieMaxLinkWidth")),
					PowerDefaultLimitMw: int(getInt64(gpu, "powerDefaultLimitMw")),
					PowerMinLimitMw:     int(getInt64(gpu, "powerMinLimitMw")),
					PowerMaxLimitMw:     int(getInt64(gpu, "powerMaxLimitMw")),
					VBIOSVersion:        getString(gpu, "vbiosVersion"),
					NumFans:             int(getInt64(gpu, "numFans")),
					TempShutdownC:       int(getInt64(gpu, "tempShutdownC")),
					TempSlowdownC:       int(getInt64(gpu, "tempSlowdownC")),
					TempMaxOperatingC:   int(getInt64(gpu, "tempMaxOperatingC")),
					ECCModeEnabled:      getBool(gpu, "eccModeEnabled"),
					PersistenceModeOn:   getBool(gpu, "persistenceModeOn"),
					ComputeMode:         getString(gpu, "computeMode"),
					MIGModeEnabled:      getBool(gpu, "migModeEnabled"),
					NVLinkCount:         int(getInt64(gpu, "nvlinkCount")),
				})
			}
		}
	}

	return data
}

// Helper functions for extracting typed values from maps

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		return formatting.ToInt64(v)
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		return formatting.ToFloat(v)
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		return formatting.ToBool(v)
	}
	return false
}
