package utils

const (
	JiffiesPerSecond    = 100
	NanosecondsPerSec   = 1_000_000_000
	CgroupDir           = "/sys/fs/cgroup"
	ProcNetDev          = "/proc/net/dev"
	ProcSelfCgroup      = "/proc/self/cgroup"
	MaxNvLinks          = 18
	VLLMEnvVar          = "VLLM_METRICS_ENDPOINT"
	DefaultVLLMEndpoint = "http://localhost:8000/metrics"
	CMDSeparator        = "--"
	//procStat      = "/proc/stat"
	//procCPUInfo   = "/proc/cpuinfo"
	//procLoadavg   = "/proc/loadavg"
	//procMeminfo   = "/proc/meminfo"
	//procVmstat    = "/proc/vmstat"
	//procDiskstats = "/proc/diskstats"
	//procNetDev    = "/proc/net/dev"
	//sysCPUPath    = "/sys/devices/system/cpu"
	//sysBlockPath  = "/sys/class/block"
	//sysNetPath    = "/sys/class/net"
	//
	//nanosecondsPerSec = 1_000_000_000
	//jiffiesPerSecond  = 100
	//diskPattern = regexp.MustCompile(`^(sd[a-z]+|nvme\d+n\d+|vd[a-z]+|xvd[a-z]+|hd[a-z]+)$`)
)
