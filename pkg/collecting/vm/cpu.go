package vm

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/beevik/ntp"
)

var (
	procStat        = "/proc/stat"
	procCpuinfo     = "/proc/cpuinfo"
	procLoadavg     = "/proc/loadavg"
	sysCpuCacheBase = "/sys/devices/system/cpu"
)

type CpuStatic struct {
	HostName      string `json:"HostName"`
	NumProcessors int    `json:"NumProcessors"`
	CPUType       string `json:"Type"`
	CPUCache      string `json:"Cache"`
	KernelInfo    string `json:"KernelInfo"`
	NTPInfo       string `json:"NTPInfo"`
}

type CpuDynamic struct {
	TimeUserMode    base.MetricInt   `json:"TimeUserMode"`
	TimeKernelMode  base.MetricInt   `json:"TimeKernelMode"`
	IdleTime        base.MetricInt   `json:"IdleTime"`
	TimeIOWait      base.MetricInt   `json:"TimeIOWait"`
	TimeIntSrvc     base.MetricInt   `json:"TimeIntSrvc"`
	TimeSoftIntSrvc base.MetricInt   `json:"TimeSoftIntSrvc"`
	Nice            base.MetricInt   `json:"Nice"`
	Steal           base.MetricInt   `json:"Steal"`
	ContextSwitches base.MetricInt   `json:"ContextSwitches"`
	LoadAvg         base.MetricFloat `json:"LoadAvg"`
	Mhz             base.MetricFloat `json:"Mhz"`
}

func collectCpuStatic(s *CpuStatic) {
	s.HostName = getHostname()
	s.NumProcessors = runtime.NumCPU()
	s.CPUType = getCPUType()
	s.CPUCache = getCPUCache()
	s.KernelInfo = getKernelInfo()

	utils.Debugf("cpu: static hostname=%s cpus=%d type=%s", s.HostName, s.NumProcessors, s.CPUType)
	utils.Debugf("cpu: cache=%s", s.CPUCache)

	t := utils.DebugTimer()
	s.NTPInfo = getNTPInfo()
	utils.DebugDuration("cpu", "NTP query", t)
}

func collectCpuDynamic(d *CpuDynamic) {
	lines, tStat, err := utils.FileLines(procStat)
	if err != nil {
		utils.Debugf("cpu: failed to read %s: %v", procStat, err)
		return
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "cpu" && len(fields) >= 9 {
			d.TimeUserMode = base.MetricInt{V: utils.ParseInt64(fields[1]), T: tStat}
			d.Nice = base.MetricInt{V: utils.ParseInt64(fields[2]), T: tStat}
			d.TimeKernelMode = base.MetricInt{V: utils.ParseInt64(fields[3]), T: tStat}
			d.IdleTime = base.MetricInt{V: utils.ParseInt64(fields[4]), T: tStat}
			d.TimeIOWait = base.MetricInt{V: utils.ParseInt64(fields[5]), T: tStat}
			d.TimeIntSrvc = base.MetricInt{V: utils.ParseInt64(fields[6]), T: tStat}
			d.TimeSoftIntSrvc = base.MetricInt{V: utils.ParseInt64(fields[7]), T: tStat}
			d.Steal = base.MetricInt{V: utils.ParseInt64(fields[8]), T: tStat}
		} else if fields[0] == "ctxt" && len(fields) >= 2 {
			d.ContextSwitches = base.MetricInt{V: utils.ParseInt64(fields[1]), T: tStat}
		}
	}
	d.LoadAvg = getLoadAvg()
	d.Mhz = getCPUFreq()
}

func getCPUType() string {
	lines, _, err := utils.FileLines(procCpuinfo)
	if err != nil {
		utils.Debugf("cpu: failed to read %s: %v", procCpuinfo, err)
		return utils.UnavailableValue
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			if parts := strings.SplitN(line, utils.FieldSeparatorColon, 2); len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return utils.UnavailableValue
}

func getCPUCache() string {
	result := make(map[string]string)
	seen := make(map[string]bool)
	dirs, err := filepath.Glob(filepath.Join(sysCpuCacheBase, "cpu*/cache/index*"))
	if err != nil {
		utils.Debugf("cpu: cache glob error: %v", err)
	}
	utils.Debugf("cpu: found %d cache index entries", len(dirs))
	for _, dir := range dirs {
		level, _, _ := utils.File(filepath.Join(dir, "level"))
		cType, _, _ := utils.File(filepath.Join(dir, "type"))
		sizeStr, _, _ := utils.File(filepath.Join(dir, "size"))
		shared, _, _ := utils.File(filepath.Join(dir, "shared_cpu_map"))
		cacheID := fmt.Sprintf("L%s-%s-%s", level, cType, shared)
		if seen[cacheID] || level == "" || sizeStr == "" {
			continue
		}
		seen[cacheID] = true
		suffix := ""
		if level == "1" {
			switch cType {
			case "Data":
				suffix = "d"
			case "Instruction":
				suffix = "i"
			}
		}
		result["L"+level+suffix] = sizeStr
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func getKernelInfo() string {
	var u syscall.Utsname
	if err := syscall.Uname(&u); err != nil {
		utils.Debugf("cpu: syscall.Uname failed: %v", err)
		return utils.UnavailableValue
	}
	data, err := json.Marshal(map[string]string{
		"Sysname":  utils.ByteSliceToString(u.Sysname[:]),
		"Nodename": utils.ByteSliceToString(u.Nodename[:]),
		"Release":  utils.ByteSliceToString(u.Release[:]),
		"Version":  utils.ByteSliceToString(u.Version[:]),
		"Machine":  utils.ByteSliceToString(u.Machine[:]),
	})
	if err != nil {
		return utils.UnavailableValue
	}
	return string(data)
}

func getNTPInfo() string {
	response, err := ntp.Query("pool.ntp.org")
	if err != nil {
		utils.Debugf("cpu: NTP query failed: %v", err)
		data, _ := json.Marshal(map[string]interface{}{
			"Status": err.Error(),
		})
		return string(data)
	}
	utils.Debugf("cpu: NTP offset=%v rtt=%v stratum=%d", response.ClockOffset, response.RTT, response.Stratum)
	data, err := json.Marshal(map[string]interface{}{
		"Status":         "success",
		"ClockOffset":    response.ClockOffset,
		"Time":           response.Time,
		"RTT":            response.RTT,
		"Precision":      response.Precision,
		"Stratum":        response.Stratum,
		"ReferenceID":    response.ReferenceID,
		"ReferenceTime":  response.ReferenceTime,
		"RootDelay":      response.RootDelay,
		"RootDispersion": response.RootDispersion,
		"RootDistance":   response.RootDistance,
		"Leap":           response.Leap,
		"MinError":       response.MinError,
		"KissCode":       response.KissCode,
		"Poll":           response.Poll,
	})
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func getLoadAvg() base.MetricFloat {
	val, ts, _ := utils.File(procLoadavg)
	if parts := strings.Fields(val); len(parts) > 0 {
		return base.MetricFloat{V: utils.ParseFloat64(parts[0]), T: ts}
	}
	return base.MetricFloat{T: ts}
}

func getCPUFreq() base.MetricFloat {
	if val, ts, err := utils.FileInt("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq"); err == nil {
		return base.MetricFloat{V: float64(val), T: ts}
	}
	return base.MetricFloat{T: utils.GetTimestamp()}
}

func getHostname() string {
	if h, err := os.Hostname(); err == nil {
		return h
	}
	return "unknown"
}
