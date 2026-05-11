package vm

import (
	"InferenceProfiler/pkg/collecting/base"
	"InferenceProfiler/pkg/utils"
	"path/filepath"
	"strings"
)

type NetInfo struct {
	Name  string `json:"name"`
	MAC   string `json:"mac"`
	MTU   int64  `json:"mtu"`
	Speed int64  `json:"speed"`
}

type NetStatic struct {
	Interfaces []NetInfo `json:"networkInterfaces"`
}

type NetDynamic struct {
	BytesRecvd   base.MetricInt `json:"BytesRecvd"`
	BytesSent    base.MetricInt `json:"BytesSent"`
	PacketsRecvd base.MetricInt `json:"PacketsRecvd"`
	PacketsSent  base.MetricInt `json:"PacketsSent"`
	ErrorsRecvd  base.MetricInt `json:"ErrorsRecvd"`
	ErrorsSent   base.MetricInt `json:"ErrorsSent"`
	DropsRecvd   base.MetricInt `json:"DropsRecvd"`
	DropsSent    base.MetricInt `json:"DropsSent"`
}

var (
	procNetDev = "/proc/net/dev"
	sysNetBase = "/sys/class/net"
)

func collectNetStatic(s *NetStatic) {
	s.Interfaces = []NetInfo{}

	entries, err := filepath.Glob(filepath.Join(sysNetBase, "*"))
	if err != nil {
		utils.Debugf("net: glob %s/* error: %v", sysNetBase, err)
	}
	for _, entry := range entries {
		name := filepath.Base(entry)
		if name == "lo" {
			continue
		}
		info := NetInfo{Name: name}
		if mac, _, _ := utils.File(filepath.Join(entry, "address")); mac != "" && mac != "00:00:00:00:00:00" {
			info.MAC = mac
		}
		if mtu, _, err := utils.FileInt(filepath.Join(entry, "mtu")); err == nil {
			info.MTU = mtu
		}
		if speed, _, err := utils.FileInt(filepath.Join(entry, "speed")); err == nil && speed > 0 {
			info.Speed = speed
		}
		s.Interfaces = append(s.Interfaces, info)
	}
	utils.Debugf("net: found %d interfaces", len(s.Interfaces))
	for _, iface := range s.Interfaces {
		utils.Debugf("net: %s mac=%s mtu=%d speed=%d", iface.Name, iface.MAC, iface.MTU, iface.Speed)
	}
}

func collectNetDynamic(d *NetDynamic) {
	lines, ts, err := utils.FileLines(procNetDev)
	if err != nil {
		utils.Debugf("net: failed to read %s: %v", procNetDev, err)
		return
	}

	var brV, bsV, prV, psV, erV, esV, drV, dsV int64

	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		if name == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}
		brV += utils.ParseInt64(fields[0])
		prV += utils.ParseInt64(fields[1])
		erV += utils.ParseInt64(fields[2])
		drV += utils.ParseInt64(fields[3])
		bsV += utils.ParseInt64(fields[8])
		psV += utils.ParseInt64(fields[9])
		esV += utils.ParseInt64(fields[10])
		dsV += utils.ParseInt64(fields[11])
	}

	d.BytesRecvd = base.MetricInt{V: brV, T: ts}
	d.PacketsRecvd = base.MetricInt{V: prV, T: ts}
	d.ErrorsRecvd = base.MetricInt{V: erV, T: ts}
	d.DropsRecvd = base.MetricInt{V: drV, T: ts}
	d.BytesSent = base.MetricInt{V: bsV, T: ts}
	d.PacketsSent = base.MetricInt{V: psV, T: ts}
	d.ErrorsSent = base.MetricInt{V: esV, T: ts}
	d.DropsSent = base.MetricInt{V: dsV, T: ts}
}
