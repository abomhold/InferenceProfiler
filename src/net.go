package src

import (
	"strconv"
	"strings"

	"github.com/inference-profiler/utils"
)

// NetMetrics contains VM-level network measurements.
// Aggregated across all non-loopback interfaces.
type NetMetrics struct {
	// Byte counters
	BytesRecvd exporter.Timed[int64] `json:"vNetworkBytesRecvd"`
	BytesSent  exporter.Timed[int64] `json:"vNetworkBytesSent"`

	// Packet counters
	PacketsRecvd exporter.Timed[int64] `json:"vNetworkPacketsRecvd"`
	PacketsSent  exporter.Timed[int64] `json:"vNetworkPacketsSent"`

	// Error counters
	ErrorsRecvd exporter.Timed[int64] `json:"vNetworkErrorsRecvd"`
	ErrorsSent  exporter.Timed[int64] `json:"vNetworkErrorsSent"`

	// Drop counters
	DropsRecvd exporter.Timed[int64] `json:"vNetworkDropsRecvd"`
	DropsSent  exporter.Timed[int64] `json:"vNetworkDropsSent"`
}

// CollectNet gathers network metrics from /proc/net/dev.
// Skips loopback interface.
func CollectNet() NetMetrics {
	lines, ts := utils.readLines("/proc/net/dev")

	var bytesRecv, packetsRecv, errsRecv, dropsRecv int64
	var bytesSent, packetsSent, errsSent, dropsSent int64

	// Skip first 2 header lines
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}

		iface := strings.TrimSpace(line[:colonIdx])
		data := line[colonIdx+1:]

		// Skip loopback
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(data)
		if len(fields) < 16 {
			continue
		}

		// /proc/net/dev columns after interface name:
		// Receive: bytes packets errs drop fifo frame compressed multicast
		// Transmit: bytes packets errs drop fifo colls carrier compressed
		bytesRecv += parseNetField(fields[0])
		packetsRecv += parseNetField(fields[1])
		errsRecv += parseNetField(fields[2])
		dropsRecv += parseNetField(fields[3])

		bytesSent += parseNetField(fields[8])
		packetsSent += parseNetField(fields[9])
		errsSent += parseNetField(fields[10])
		dropsSent += parseNetField(fields[11])
	}

	return NetMetrics{
		BytesRecvd:   exporter.TimedAt(bytesRecv, ts),
		BytesSent:    exporter.TimedAt(bytesSent, ts),
		PacketsRecvd: exporter.TimedAt(packetsRecv, ts),
		PacketsSent:  exporter.TimedAt(packetsSent, ts),
		ErrorsRecvd:  exporter.TimedAt(errsRecv, ts),
		ErrorsSent:   exporter.TimedAt(errsSent, ts),
		DropsRecvd:   exporter.TimedAt(dropsRecv, ts),
		DropsSent:    exporter.TimedAt(dropsSent, ts),
	}
}

func parseNetField(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
