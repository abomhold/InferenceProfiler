package container

import (
	"os"
	"path/filepath"
	"strings"

	"InferenceProfiler/pkg/utils"
)

var (
	cgroupDir      = "/sys/fs/cgroup"
	procSelfCgroup = "/proc/self/cgroup"
	procNetDev     = "/proc/net/dev"
)

var (
	cgVersion int
	v2Path    string
)

func detect() int {
	if !utils.IsDir(cgroupDir) {
		utils.Debugf("container: %s not found, not in container", cgroupDir)
		return 0
	}

	if utils.Exists(filepath.Join(cgroupDir, "cgroup.controllers")) {
		v2Path = findV2Path()
		if v2Path == "" {
			utils.Debugf("container: cgroup v2 detected but no valid path found")
			return 0
		}
		cgVersion = 2
		utils.Debugf("container: detected cgroup v2, path=%s", v2Path)
		return 2
	}

	cgVersion = 1
	utils.Debugf("container: detected cgroup v1")
	return 1
}

func version() int {
	return cgVersion
}

func findV2Path() string {
	lines, _, err := utils.FileLines(procSelfCgroup)
	if err != nil {
		utils.Debugf("container: failed to read %s: %v", procSelfCgroup, err)
		return cgroupDir
	}
	for _, line := range lines {
		parts := strings.SplitN(line, utils.FieldSeparatorColon, 3)
		if len(parts) == 3 && parts[0] == "0" {
			path := cgroupDir + parts[2]
			if utils.IsDir(path) {
				return path
			}
		}
	}
	return cgroupDir
}

func getContainerID() string {
	lines, _, err := utils.FileLines(procSelfCgroup)
	if err != nil {
		utils.Debugf("container: failed to read %s for container ID: %v", procSelfCgroup, err)
	}
	for _, line := range lines {
		if parts := strings.SplitN(line, utils.FieldSeparatorColon, 3); len(parts) >= 3 {
			path := parts[2]
			if segments := strings.Split(path, "/docker/"); len(segments) > 1 {
				return segments[len(segments)-1]
			}
		}
	}

	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return utils.UnavailableValue
}

func getNetStats() (recv, sent, ts int64) {
	lines, ts, err := utils.FileLines(procNetDev)
	if err != nil {
		utils.Debugf("container: failed to read %s: %v", procNetDev, err)
		return 0, 0, ts
	}
	for _, line := range lines {
		if !strings.Contains(line, utils.FieldSeparatorColon) || strings.Contains(line, "lo"+utils.FieldSeparatorColon) {
			continue
		}
		fields := strings.Fields(strings.SplitN(line, utils.FieldSeparatorColon, 2)[1])
		if len(fields) >= 9 {
			recv += utils.ParseInt64(fields[0])
			sent += utils.ParseInt64(fields[8])
		}
	}
	return recv, sent, ts
}
