package config

const (
	// System constants
	JiffiesPerSecond  = 100
	SectorSize        = 512
	NanosecondsPerSec = 1_000_000_000

	// Filesystem paths
	CgroupDir      = "/sys/fs/cgroup"
	ProcStat       = "/proc/stat"
	ProcMeminfo    = "/proc/meminfo"
	ProcVMStat     = "/proc/vmstat"
	ProcDiskstats  = "/proc/diskstats"
	ProcNetDev     = "/proc/net/dev"
	ProcLoadavg    = "/proc/loadavg"
	ProcCPUInfo    = "/proc/cpuinfo"
	ProcSelfCgroup = "/proc/self/cgroup"
	SysClassBlock  = "/sys/class/block"
	SysClassNet    = "/sys/class/net"
	SysCPUPath     = "/sys/devices/system/cpu"

	// Network
	LoopbackInterface = "lo"

	// NVIDIA
	MaxNvLinks = 18

	// vLLM
	DefaultVLLMEndpoint = "http://localhost:8000/metrics"
	VLLMEnvVar          = "VLLM_METRICS_URL"

	// Regex patterns
	DiskRegex = `^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`
)
