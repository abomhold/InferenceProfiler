package collecting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"InferenceProfiler/pkg/metrics/vm"
	"InferenceProfiler/pkg/probing"
)

type Network struct{}

func NewNetwork() *Network      { return &Network{} }
func (c *Network) Name() string { return "VM-Network" }
func (c *Network) Close() error { return nil }

func (c *Network) CollectStatic() any {
	m := &vm.NetworkStatic{}

	interfaces, _ := os.ReadDir("/sys/class/net/")
	var netInterfaces []vm.NetworkInterface

	for _, entry := range interfaces {
		iface := entry.Name()
		if iface == LoopbackInterface {
			continue
		}

		basePath := filepath.Join("/sys/class/net", iface)

		macVal, _ := probing.File(filepath.Join(basePath, "address"))
		stateVal, _ := probing.File(filepath.Join(basePath, "operstate"))
		mtu, _ := probing.FileInt(filepath.Join(basePath, "mtu"))
		speed, _ := probing.FileInt(filepath.Join(basePath, "speed"))

		ni := vm.NetworkInterface{
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
		m.NetworkInterfacesJSON = string(data)
	}

	return m
}

func (c *Network) CollectDynamic() any {
	m := &vm.NetworkDynamic{}

	lines, tNet := probing.FileLines("/proc/net/dev")
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
		if iface == LoopbackInterface {
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

	m.NetworkBytesRecvd, m.NetworkBytesRecvdT = bRecv, tNet
	m.NetworkBytesSent, m.NetworkBytesSentT = bSent, tNet
	m.NetworkPacketsRecvd, m.NetworkPacketsRecvdT = pRecv, tNet
	m.NetworkPacketsSent, m.NetworkPacketsSentT = pSent, tNet
	m.NetworkErrorsRecvd, m.NetworkErrorsRecvdT = eRecv, tNet
	m.NetworkErrorsSent, m.NetworkErrorsSentT = eSent, tNet
	m.NetworkDropsRecvd, m.NetworkDropsRecvdT = dRecv, tNet
	m.NetworkDropsSent, m.NetworkDropsSentT = dSent, tNet

	return m
}
