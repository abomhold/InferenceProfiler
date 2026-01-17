package config

const (
	// System constants
	JiffiesPerSecond  = 100
	SectorSize        = 512
	NanosecondsPerSec = 1_000_000_000
	LoopbackInterface = "lo"

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

	// Disk regex pattern for filtering disks
	DiskRegex = `^(sd[a-z]+|nvme\d+n\d+|vd[a-z]+|xvd[a-z]+|hd[a-z]+)$`

	// NVIDIA GPU constants
	MaxNvLinks = 18

	// vLLM constants
	VLLMEnvVar          = "VLLM_METRICS_ENDPOINT"
	DefaultVLLMEndpoint = "http://localhost:8000/metrics"
)
