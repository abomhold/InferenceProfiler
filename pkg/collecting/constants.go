package collecting

const (
	jiffiesPerSecond          = 100
	maxNvlinks                = 16
	procDir                   = "/proc"
	cgroupDir                 = "/sys/fs/cgroup"
	sectorBytes               = 512
	pageSize                  = 4096
	bytesPerKilobyte          = 1024
	bytesPerMegaByte          = 1048576
	microsecondsToJiffiesDiv  = 10000
	microsecondsToNanoseconds = 1000
	unknownValue              = "unknown"
	unavailableValue          = "unavailable"
	fieldSeparatorColon       = ":"
	fieldSeparatorSpace       = " "
	fieldSeparatorNull        = "\x00"
	vllmBucketSuffix          = "_bucket"
	vllmBucketPrefix          = "{le=\""
	vllmBucketEnd             = "\"}"
	vllmCommentPrefix         = "#"
)
