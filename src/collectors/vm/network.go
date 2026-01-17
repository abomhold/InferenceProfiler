package vm

import (
	"InferenceProfiler/src/collectors"
	"InferenceProfiler/src/collectors/types"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// NetworkCollector collects network metrics from /proc/net/dev
type NetworkCollector struct {
	collectors.BaseCollector
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

func (c *NetworkCollector) Name() string {
	return "Network"
}

func (c *NetworkCollector) CollectStatic(m *types.StaticMetrics) {
	interfaces, _ := os.ReadDir("/sys/class/net/")
	var netInterfaces []types.NetworkInterfaceStatic

	for _, entry := range interfaces {
		iface := entry.Name()
		if iface == collectors.LoopbackInterface {
			continue
		}

		basePath := filepath.Join("/sys/class/net", iface)

		mac, _ := collectors.ProbeFile(filepath.Join(basePath, "address"))
		state, _ := collectors.ProbeFile(filepath.Join(basePath, "operstate"))
		mtu, _ := collectors.ProbeFileInt(filepath.Join(basePath, "mtu"))
		speed, _ := collectors.ProbeFileInt(filepath.Join(basePath, "speed"))

		ni := types.NetworkInterfaceStatic{
			Name:  iface,
			MAC:   strings.TrimSpace(mac),
			State: strings.TrimSpace(state),
			MTU:   mtu,
		}

		// Speed can be -1 if the interface is down or doesn't support it
		if speed > 0 {
			ni.SpeedMbps = speed
		}

		netInterfaces = append(netInterfaces, ni)
	}

	if len(netInterfaces) > 0 {
		if data, err := json.Marshal(netInterfaces); err == nil {
			m.NetworkInterfacesJSON = string(data)
		}
	}
}

func (c *NetworkCollector) CollectDynamic(m *types.DynamicMetrics) {
	lines, ts := collectors.ProbeFileLines("/proc/net/dev")
	var bRecv, pRecv, eRecv, dRecv int64
	var bSent, pSent, eSent, dSent int64

	// Skip header lines (first 2 lines)
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
		if iface == collectors.LoopbackInterface {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 12 {
			continue
		}

		bRecv += collectors.parseInt64(fields[0])
		pRecv += collectors.parseInt64(fields[1])
		eRecv += collectors.parseInt64(fields[2])
		dRecv += collectors.parseInt64(fields[3])
		bSent += collectors.parseInt64(fields[8])
		pSent += collectors.parseInt64(fields[9])
		eSent += collectors.parseInt64(fields[10])
		dSent += collectors.parseInt64(fields[11])
	}

	m.NetworkBytesRecvd = bRecv
	m.NetworkBytesRecvdT = ts
	m.NetworkBytesSent = bSent
	m.NetworkBytesSentT = ts
	m.NetworkPacketsRecvd = pRecv
	m.NetworkPacketsRecvdT = ts
	m.NetworkPacketsSent = pSent
	m.NetworkPacketsSentT = ts
	m.NetworkErrorsRecvd = eRecv
	m.NetworkErrorsRecvdT = ts
	m.NetworkErrorsSent = eSent
	m.NetworkErrorsSentT = ts
	m.NetworkDropsRecvd = dRecv
	m.NetworkDropsRecvdT = ts
	m.NetworkDropsSent = dSent
	m.NetworkDropsSentT = ts
}
