package collectors

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const LoopbackInterface = "lo"

// --- Static Metrics ---

func CollectNetworkStatic() StaticMetrics {
	results := make(StaticMetrics)
	interfaces, _ := os.ReadDir("/sys/class/net/")

	idx := 0
	for _, entry := range interfaces {
		iface := entry.Name()
		if iface == "lo" {
			continue
		}

		basePath := filepath.Join("/sys/class/net", iface)

		mac, _ := ProbeFile(filepath.Join(basePath, "address"))
		state, _ := ProbeFile(filepath.Join(basePath, "operstate"))
		mtu, _ := ProbeFileInt(filepath.Join(basePath, "mtu"))
		speed, _ := ProbeFileInt(filepath.Join(basePath, "speed")) // Speed in Mbps

		prefix := "vNetwork" + strconv.Itoa(idx)
		results[prefix+"Name"] = iface
		results[prefix+"MAC"] = strings.TrimSpace(mac)
		results[prefix+"State"] = strings.TrimSpace(state)
		results[prefix+"MTU"] = mtu

		// Speed can be -1 if the interface is down or doesn't support it
		if speed > 0 {
			results[prefix+"SpeedMbps"] = speed
		}

		idx++
	}

	return results
}

// --- Dynamic Metrics ---

func CollectNetworkDynamic() DynamicMetrics {
	lines, ts := ProbeFileLines("/proc/net/dev")
	var bRecv, pRecv, eRecv, dRecv int64
	var bSent, pSent, eSent, dSent int64

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

	return DynamicMetrics{
		"vNetworkBytesRecvd":   NewMetricWithTime(bRecv, ts),
		"vNetworkBytesSent":    NewMetricWithTime(bSent, ts),
		"vNetworkPacketsRecvd": NewMetricWithTime(pRecv, ts),
		"vNetworkPacketsSent":  NewMetricWithTime(pSent, ts),
		"vNetworkErrorsRecvd":  NewMetricWithTime(eRecv, ts),
		"vNetworkErrorsSent":   NewMetricWithTime(eSent, ts),
		"vNetworkDroppedRecvd": NewMetricWithTime(dRecv, ts),
		"vNetworkDroppedSent":  NewMetricWithTime(dSent, ts),
	}

}
