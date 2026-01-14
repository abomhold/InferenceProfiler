package collectors

// System constants
const (
	JiffiesPerSecond  = 100
	SectorSize        = 512
	CgroupDir         = "/sys/fs/cgroup"
	LoopbackInterface = "lo"
	MaxNvLinks        = 18
	DiskRegex         = `^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`
)
