package collectors

import (
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// CollectProcesses collects per-process metrics
func CollectProcesses() (int64, []map[string]MetricValue) {
	var processes []map[string]MetricValue
	var processCount int64

	dirs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return 0, nil
	}

	pageSize := int64(syscall.Getpagesize())

	for _, pidPath := range dirs {
		pid, err := strconv.ParseInt(filepath.Base(pidPath), 10, 64)
		if err != nil {
			continue
		}

		statData, tStatData := ProbeFile(filepath.Join(pidPath, "stat"))
		cmdline, _ := ProbeFile(filepath.Join(pidPath, "cmdline"))
		statusData, tStatusData := ParseProcKV(filepath.Join(pidPath, "status"), ":")
		statmData, tStatm := ProbeFile(filepath.Join(pidPath, "statm"))

		if statData == "" {
			continue
		}

		rparenIndex := strings.LastIndex(statData, ")")
		if rparenIndex == -1 {
			continue
		}

		statParts := strings.Fields(statData[rparenIndex+2:])
		if len(statParts) < 40 {
			continue
		}

		processCount++

		var rssPages int64
		if statmData != "" {
			statmParts := strings.Fields(statmData)
			if len(statmParts) >= 2 {
				rssPages, _ = strconv.ParseInt(statmParts[1], 10, 64)
			}
		}
		rssBytes := rssPages * pageSize

		pName := statusData["Name"]
		if pName == "" {
			lparenIndex := strings.Index(statData, "(")
			if lparenIndex != -1 && rparenIndex > lparenIndex {
				pName = statData[lparenIndex+1 : rparenIndex]
			}
		}

		proc := make(map[string]MetricValue)
		proc["pId"] = MetricValue{Value: pid, Time: tStatData}
		proc["pName"] = MetricValue{Value: pName, Time: tStatData}
		proc["pCmdline"] = MetricValue{Value: strings.ReplaceAll(cmdline, "\x00", " "), Time: tStatData}

		proc["pNumThreads"] = NewMetricWithTime(parseInt64(statParts[17]), tStatData)

		proc["pCpuTimeUserMode"] = NewMetricWithTime(parseInt64(statParts[11])*JiffiesPerSecond, tStatData)
		proc["pCpuTimeKernelMode"] = NewMetricWithTime(parseInt64(statParts[12])*JiffiesPerSecond, tStatData)
		proc["pChildrenUserMode"] = NewMetricWithTime(parseInt64(statParts[13])*JiffiesPerSecond, tStatData)
		proc["pChildrenKernelMode"] = NewMetricWithTime(parseInt64(statParts[14])*JiffiesPerSecond, tStatData)

		volCtxSwitch, _ := strconv.ParseInt(statusData["voluntary_ctxt_switches"], 10, 64)
		nonvolCtxSwitch, _ := strconv.ParseInt(statusData["nonvoluntary_ctxt_switches"], 10, 64)
		proc["pVoluntaryContextSwitches"] = NewMetricWithTime(volCtxSwitch, tStatusData)
		proc["pNonvoluntaryContextSwitches"] = NewMetricWithTime(nonvolCtxSwitch, tStatusData)

		proc["pBlockIODelays"] = NewMetricWithTime(parseInt64(statParts[39])*JiffiesPerSecond, tStatData)

		proc["pVirtualMemoryBytes"] = NewMetricWithTime(parseInt64(statParts[20]), tStatData)
		proc["pResidentSetSize"] = NewMetricWithTime(rssBytes, tStatm)

		processes = append(processes, proc)
	}

	return processCount, processes
}
