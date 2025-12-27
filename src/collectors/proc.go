package collectors

import (
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func CollectProcesses() (int64, []map[string]MetricValue) {
	dirs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return 0, nil
	}

	pageSize := int64(syscall.Getpagesize())
	var processes []map[string]MetricValue

	for _, pidPath := range dirs {
		pid, err := strconv.ParseInt(filepath.Base(pidPath), 10, 64)
		if err != nil {
			continue
		}

		statData, tStat := ProbeFile(filepath.Join(pidPath, "stat"))
		rparenIndex := strings.LastIndex(statData, ")")
		if rparenIndex == -1 || len(statData) < rparenIndex+3 {
			continue
		}

		statParts := strings.Fields(statData[rparenIndex+2:])
		if len(statParts) < 40 {
			continue
		}

		cmdline, _ := ProbeFile(filepath.Join(pidPath, "cmdline"))
		status, tStatus := ParseProcKV(filepath.Join(pidPath, "status"), ":")
		statm, tStatm := ProbeFile(filepath.Join(pidPath, "statm"))

		var rssBytes int64
		if parts := strings.Fields(statm); len(parts) >= 2 {
			rssPages, _ := strconv.ParseInt(parts[1], 10, 64)
			rssBytes = rssPages * pageSize
		}

		pName := status["Name"]
		if pName == "" {
			if lp := strings.Index(statData, "("); lp != -1 && rparenIndex > lp {
				pName = statData[lp+1 : rparenIndex]
			}
		}

		processes = append(processes, map[string]MetricValue{
			"pId":                          {Value: pid, Time: tStat},
			"pName":                        {Value: pName, Time: tStat},
			"pCmdline":                     {Value: strings.ReplaceAll(cmdline, "\x00", " "), Time: tStat},
			"pNumThreads":                  NewMetricWithTime(parseInt64(statParts[17]), tStat),
			"pCpuTimeUserMode":             NewMetricWithTime(parseInt64(statParts[11])*JiffiesPerSecond, tStat),
			"pCpuTimeKernelMode":           NewMetricWithTime(parseInt64(statParts[12])*JiffiesPerSecond, tStat),
			"pChildrenUserMode":            NewMetricWithTime(parseInt64(statParts[13])*JiffiesPerSecond, tStat),
			"pChildrenKernelMode":          NewMetricWithTime(parseInt64(statParts[14])*JiffiesPerSecond, tStat),
			"pVoluntaryContextSwitches":    NewMetricWithTime(parseInt64(status["voluntary_ctxt_switches"]), tStatus),
			"pNonvoluntaryContextSwitches": NewMetricWithTime(parseInt64(status["nonvoluntary_ctxt_switches"]), tStatus),
			"pBlockIODelays":               NewMetricWithTime(parseInt64(statParts[39])*JiffiesPerSecond, tStat),
			"pVirtualMemoryBytes":          NewMetricWithTime(parseInt64(statParts[20]), tStat),
			"pResidentSetSize":             NewMetricWithTime(rssBytes, tStatm),
		})
	}

	return int64(len(processes)), processes
}
