package collectors

import (
	"strings"
)

type NetCollector struct {
	BaseCollector
}

func (c *NetCollector) Collect() NetMetrics {
	lines, ts := c.ReadLines("/proc/net/dev")

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

		bytesRecv += c.ParseInt(fields[0])
		packetsRecv += c.ParseInt(fields[1])
		errsRecv += c.ParseInt(fields[2])
		dropsRecv += c.ParseInt(fields[3])

		bytesSent += c.ParseInt(fields[8])
		packetsSent += c.ParseInt(fields[9])
		errsSent += c.ParseInt(fields[10])
		dropsSent += c.ParseInt(fields[11])
	}

	return NetMetrics{
		BytesRecvd:   TimedAt(bytesRecv, ts),
		BytesSent:    TimedAt(bytesSent, ts),
		PacketsRecvd: TimedAt(packetsRecv, ts),
		PacketsSent:  TimedAt(packetsSent, ts),
		ErrorsRecvd:  TimedAt(errsRecv, ts),
		ErrorsSent:   TimedAt(errsSent, ts),
		DropsRecvd:   TimedAt(dropsRecv, ts),
		DropsSent:    TimedAt(dropsSent, ts),
	}
}
