package vm

type CPUStatic struct {
	NumProcessors       int     `json:"vNumProcessors"`
	CPUType             string  `json:"vCpuType"`
	CPUCache            string  `json:"vCpuCache"`
	KernelInfo          string  `json:"vKernelInfo"`
	TimeSynced          bool    `json:"vTimeSynced"`
	TimeOffsetSeconds   float64 `json:"vTimeOffsetSeconds"`
	TimeMaxErrorSeconds float64 `json:"vTimeMaxErrorSeconds"`
}

type CPUDynamic struct {
	CPUTime             int64   `json:"vCpuTime"`
	CPUTimeT            int64   `json:"vCpuTimeT"`
	CPUTimeUserMode     int64   `json:"vCpuTimeUserMode"`
	CPUTimeUserModeT    int64   `json:"vCpuTimeUserModeT"`
	CPUTimeKernelMode   int64   `json:"vCpuTimeKernelMode"`
	CPUTimeKernelModeT  int64   `json:"vCpuTimeKernelModeT"`
	CPUIdleTime         int64   `json:"vCpuIdleTime"`
	CPUIdleTimeT        int64   `json:"vCpuIdleTimeT"`
	CPUTimeIOWait       int64   `json:"vCpuTimeIOWait"`
	CPUTimeIOWaitT      int64   `json:"vCpuTimeIOWaitT"`
	CPUTimeIntSrvc      int64   `json:"vCpuTimeIntSrvc"`
	CPUTimeIntSrvcT     int64   `json:"vCpuTimeIntSrvcT"`
	CPUTimeSoftIntSrvc  int64   `json:"vCpuTimeSoftIntSrvc"`
	CPUTimeSoftIntSrvcT int64   `json:"vCpuTimeSoftIntSrvcT"`
	CPUNice             int64   `json:"vCpuNice"`
	CPUNiceT            int64   `json:"vCpuNiceT"`
	CPUSteal            int64   `json:"vCpuSteal"`
	CPUStealT           int64   `json:"vCpuStealT"`
	CPUContextSwitches  int64   `json:"vCpuContextSwitches"`
	CPUContextSwitchesT int64   `json:"vCpuContextSwitchesT"`
	LoadAvg             float64 `json:"vLoadAvg"`
	LoadAvgT            int64   `json:"vLoadAvgT"`
	CPUMhz              float64 `json:"vCpuMhz"`
	CPUMhzT             int64   `json:"vCpuMhzT"`
}
