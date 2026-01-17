package config

const (
	JiffiesPerSecond    = 100
	NanosecondsPerSec   = 1_000_000_000
	CgroupDir           = "/sys/fs/cgroup"
	ProcNetDev          = "/proc/net/dev"
	ProcSelfCgroup      = "/proc/self/cgroup"
	MaxNvLinks          = 18
	VLLMEnvVar          = "VLLM_METRICS_ENDPOINT"
	DefaultVLLMEndpoint = "http://localhost:8000/metrics"
)
