package collectors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// NetworkCollector collects network metrics from /proc/net/dev
type NetworkCollector struct {
	BaseCollector
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

func (c *NetworkCollector) Name() string {
	return "Network"
}

func (c *NetworkCollector) CollectStatic(m *StaticMetrics) {
	interfaces, _ := os.ReadDir("/sys/class/net/")
	var netInterfaces []NetworkInterfaceStatic

	for _, entry := range interfaces {
		iface := entry.Name()
		if iface == LoopbackInterface {
			continue
		}

		basePath := filepath.Join("/sys/class/net", iface)

		mac, _ := ProbeFile(filepath.Join(basePath, "address"))
		state, _ := ProbeFile(filepath.Join(basePath, "operstate"))
		mtu, _ := ProbeFileInt(filepath.Join(basePath, "mtu"))
		speed, _ := ProbeFileInt(filepath.Join(basePath, "speed"))

		ni := NetworkInterfaceStatic{
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

func (c *NetworkCollector) CollectDynamic(m *DynamicMetrics) {
	lines, ts := ProbeFileLines("/proc/net/dev")
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
		if iface == LoopbackInterface {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 12 {
			continue
		}

		bRecv += parseInt64(fields[0])
		pRecv += parseInt64(fields[1])
		eRecv += parseInt64(fields[2])
		dRecv += parseInt64(fields[3])
		bSent += parseInt64(fields[8])
		pSent += parseInt64(fields[9])
		eSent += parseInt64(fields[10])
		dSent += parseInt64(fields[11])
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
