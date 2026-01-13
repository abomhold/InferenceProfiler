package collectors

// System constants
const (
	// JiffiesPerSecond is the kernel's timer interrupt frequency (USER_HZ)
	// Used to convert time values from /proc/stat and related files
	JiffiesPerSecond = 100

	// SectorSize is the standard disk sector size in bytes
	SectorSize = 512

	// CgroupDir is the standard cgroup filesystem mount point
	CgroupDir = "/sys/fs/cgroup"

	// LoopbackInterface is excluded from network metrics
	LoopbackInterface = "lo"

	// MaxNvLinks is the maximum NVLinks on current architectures (Hopper)
	MaxNvLinks = 18

	// DiskRegex matches physical disk devices (excludes partitions and loopback)
	DiskRegex = `^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`
)
