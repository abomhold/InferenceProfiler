package collectors

import (
	"strings"
)

// CollectNetwork collects network metrics
func CollectNetwork() map[string]MetricValue {
	metrics := make(map[string]MetricValue)
	stats, ts := getNetworkStats()

	metrics["vNetworkBytesRecvd"] = NewMetricWithTime(stats["bytesRecv"], ts)
	metrics["vNetworkBytesSent"] = NewMetricWithTime(stats["bytesSent"], ts)
	metrics["vNetworkPacketsRecvd"] = NewMetricWithTime(stats["packetsRecv"], ts)
	metrics["vNetworkPacketsSent"] = NewMetricWithTime(stats["packetsSent"], ts)
	metrics["vNetworkErrorsRecvd"] = NewMetricWithTime(stats["errsRecv"], ts)
	metrics["vNetworkErrorsSent"] = NewMetricWithTime(stats["errsSent"], ts)
	metrics["vNetworkDropsRecvd"] = NewMetricWithTime(stats["dropsRecv"], ts)
	metrics["vNetworkDropsSent"] = NewMetricWithTime(stats["dropsSent"], ts)

	return metrics
}

func getNetworkStats() (map[string]int64, int64) {
	stats := map[string]int64{
		"bytesRecv":   0,
		"packetsRecv": 0,
		"errsRecv":    0,
		"dropsRecv":   0,
		"bytesSent":   0,
		"packetsSent": 0,
		"errsSent":    0,
		"dropsSent":   0,
	}

	lines, ts := ProbeFileLines("/proc/net/dev")

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
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}

		stats["bytesRecv"] += parseInt64(fields[0])
		stats["packetsRecv"] += parseInt64(fields[1])
		stats["errsRecv"] += parseInt64(fields[2])
		stats["dropsRecv"] += parseInt64(fields[3])
		stats["bytesSent"] += parseInt64(fields[8])
		stats["packetsSent"] += parseInt64(fields[9])
		stats["errsSent"] += parseInt64(fields[10])
		stats["dropsSent"] += parseInt64(fields[11])
	}

	return stats, ts
}
