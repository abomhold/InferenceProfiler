package cpu

import "InferenceProfiler/pkg/collectors/types"

// Static contains static CPU information.
type Static struct {
	NumProcessors       int
	CPUType             string
	CPUCache            string
	KernelInfo          string
	TimeSynced          bool
	TimeOffsetSeconds   float64
	TimeMaxErrorSeconds float64
}

// ToRecord converts Static to a Record.
func (s *Static) ToRecord() types.Record {
	return types.Record{
		"vNumProcessors":       s.NumProcessors,
		"vCpuType":             s.CPUType,
		"vCpuCache":            s.CPUCache,
		"vKernelInfo":          s.KernelInfo,
		"vTimeSynced":          s.TimeSynced,
		"vTimeOffsetSeconds":   s.TimeOffsetSeconds,
		"vTimeMaxErrorSeconds": s.TimeMaxErrorSeconds,
	}
}

// Dynamic contains dynamic CPU metrics.
type Dynamic struct {
	CPUTime             int64
	CPUTimeT            int64
	CPUTimeUserMode     int64
	CPUTimeUserModeT    int64
	CPUTimeKernelMode   int64
	CPUTimeKernelModeT  int64
	CPUIdleTime         int64
	CPUIdleTimeT        int64
	CPUTimeIOWait       int64
	CPUTimeIOWaitT      int64
	CPUTimeIntSrvc      int64
	CPUTimeIntSrvcT     int64
	CPUTimeSoftIntSrvc  int64
	CPUTimeSoftIntSrvcT int64
	CPUNice             int64
	CPUNiceT            int64
	CPUSteal            int64
	CPUStealT           int64
	CPUContextSwitches  int64
	CPUContextSwitchesT int64
	LoadAvg             float64
	LoadAvgT            int64
	CPUMhz              float64
	CPUMhzT             int64
}

// ToRecord converts Dynamic to a Record.
func (d *Dynamic) ToRecord() types.Record {
	return types.Record{
		"vCpuTime":             d.CPUTime,
		"vCpuTimeT":            d.CPUTimeT,
		"vCpuTimeUserMode":     d.CPUTimeUserMode,
		"vCpuTimeUserModeT":    d.CPUTimeUserModeT,
		"vCpuTimeKernelMode":   d.CPUTimeKernelMode,
		"vCpuTimeKernelModeT":  d.CPUTimeKernelModeT,
		"vCpuIdleTime":         d.CPUIdleTime,
		"vCpuIdleTimeT":        d.CPUIdleTimeT,
		"vCpuTimeIOWait":       d.CPUTimeIOWait,
		"vCpuTimeIOWaitT":      d.CPUTimeIOWaitT,
		"vCpuTimeIntSrvc":      d.CPUTimeIntSrvc,
		"vCpuTimeIntSrvcT":     d.CPUTimeIntSrvcT,
		"vCpuTimeSoftIntSrvc":  d.CPUTimeSoftIntSrvc,
		"vCpuTimeSoftIntSrvcT": d.CPUTimeSoftIntSrvcT,
		"vCpuNice":             d.CPUNice,
		"vCpuNiceT":            d.CPUNiceT,
		"vCpuSteal":            d.CPUSteal,
		"vCpuStealT":           d.CPUStealT,
		"vCpuContextSwitches":  d.CPUContextSwitches,
		"vCpuContextSwitchesT": d.CPUContextSwitchesT,
		"vLoadAvg":             d.LoadAvg,
		"vLoadAvgT":            d.LoadAvgT,
		"vCpuMhz":              d.CPUMhz,
		"vCpuMhzT":             d.CPUMhzT,
	}
}
