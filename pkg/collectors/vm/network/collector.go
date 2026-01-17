package network

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"InferenceProfiler/pkg/collectors/types"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/probing"
)

// Collector collects network metrics.
type Collector struct{}

// New creates a new Network collector.
func New() *Collector {
	return &Collector{}
}

// Name returns the collector name.
func (c *Collector) Name() string {
	return "VM-Network"
}

// Close releases any resources.
func (c *Collector) Close() error {
	return nil
}

// CollectStatic collects static network information.
func (c *Collector) CollectStatic() types.Record {
	s := &Static{}

	interfaces, _ := os.ReadDir(config.SysClassNet)
	var netInterfaces []Interface

	for _, entry := range interfaces {
		iface := entry.Name()
		if iface == config.LoopbackInterface {
			continue
		}

		basePath := filepath.Join(config.SysClassNet, iface)

		macVal, _ := probing.File(filepath.Join(basePath, "address"))
		stateVal, _ := probing.File(filepath.Join(basePath, "operstate"))
		mtu, _ := probing.FileInt(filepath.Join(basePath, "mtu"))
		speed, _ := probing.FileInt(filepath.Join(basePath, "speed"))

		ni := Interface{
			Name:  iface,
			MAC:   strings.TrimSpace(macVal),
			State: strings.TrimSpace(stateVal),
			MTU:   mtu,
		}

		if speed > 0 {
			ni.SpeedMbps = speed
		}

		netInterfaces = append(netInterfaces, ni)
	}

	if len(netInterfaces) > 0 {
		data, _ := json.Marshal(netInterfaces)
		s.NetworkInterfacesJSON = string(data)
	}

	return s.ToRecord()
}

// CollectDynamic collects dynamic network metrics.
func (c *Collector) CollectDynamic() types.Record {
	d := &Dynamic{}

	lines, tNet := probing.FileLines(config.ProcNetDev)
	var bRecv, pRecv, eRecv, dRecv int64
	var bSent, pSent, eSent, dSent int64

	// Skip header lines
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		if iface == config.LoopbackInterface {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 12 {
			continue
		}

		bRecv += probing.ParseInt64(fields[0])
		pRecv += probing.ParseInt64(fields[1])
		eRecv += probing.ParseInt64(fields[2])
		dRecv += probing.ParseInt64(fields[3])
		bSent += probing.ParseInt64(fields[8])
		pSent += probing.ParseInt64(fields[9])
		eSent += probing.ParseInt64(fields[10])
		dSent += probing.ParseInt64(fields[11])
	}

	d.NetworkBytesRecvd, d.NetworkBytesRecvdT = bRecv, tNet
	d.NetworkBytesSent, d.NetworkBytesSentT = bSent, tNet
	d.NetworkPacketsRecvd, d.NetworkPacketsRecvdT = pRecv, tNet
	d.NetworkPacketsSent, d.NetworkPacketsSentT = pSent, tNet
	d.NetworkErrorsRecvd, d.NetworkErrorsRecvdT = eRecv, tNet
	d.NetworkErrorsSent, d.NetworkErrorsSentT = eSent, tNet
	d.NetworkDropsRecvd, d.NetworkDropsRecvdT = dRecv, tNet
	d.NetworkDropsSent, d.NetworkDropsSentT = dSent, tNet

	return d.ToRecord()
}
