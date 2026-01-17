package network

import "InferenceProfiler/pkg/collectors/types"

// Interface contains static information for a network interface.
type Interface struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	State     string `json:"state"`
	MTU       int64  `json:"mtu"`
	SpeedMbps int64  `json:"speedMbps,omitempty"`
}

// Static contains static network information.
type Static struct {
	NetworkInterfacesJSON string
}

// ToRecord converts Static to a Record.
func (s *Static) ToRecord() types.Record {
	r := types.Record{}
	if s.NetworkInterfacesJSON != "" {
		r["networkInterfaces"] = s.NetworkInterfacesJSON
	}
	return r
}

// Dynamic contains dynamic network metrics.
type Dynamic struct {
	NetworkBytesRecvd    int64
	NetworkBytesRecvdT   int64
	NetworkBytesSent     int64
	NetworkBytesSentT    int64
	NetworkPacketsRecvd  int64
	NetworkPacketsRecvdT int64
	NetworkPacketsSent   int64
	NetworkPacketsSentT  int64
	NetworkErrorsRecvd   int64
	NetworkErrorsRecvdT  int64
	NetworkErrorsSent    int64
	NetworkErrorsSentT   int64
	NetworkDropsRecvd    int64
	NetworkDropsRecvdT   int64
	NetworkDropsSent     int64
	NetworkDropsSentT    int64
}

// ToRecord converts Dynamic to a Record.
func (d *Dynamic) ToRecord() types.Record {
	return types.Record{
		"vNetworkBytesRecvd":    d.NetworkBytesRecvd,
		"vNetworkBytesRecvdT":   d.NetworkBytesRecvdT,
		"vNetworkBytesSent":     d.NetworkBytesSent,
		"vNetworkBytesSentT":    d.NetworkBytesSentT,
		"vNetworkPacketsRecvd":  d.NetworkPacketsRecvd,
		"vNetworkPacketsRecvdT": d.NetworkPacketsRecvdT,
		"vNetworkPacketsSent":   d.NetworkPacketsSent,
		"vNetworkPacketsSentT":  d.NetworkPacketsSentT,
		"vNetworkErrorsRecvd":   d.NetworkErrorsRecvd,
		"vNetworkErrorsRecvdT":  d.NetworkErrorsRecvdT,
		"vNetworkErrorsSent":    d.NetworkErrorsSent,
		"vNetworkErrorsSentT":   d.NetworkErrorsSentT,
		"vNetworkDropsRecvd":    d.NetworkDropsRecvd,
		"vNetworkDropsRecvdT":   d.NetworkDropsRecvdT,
		"vNetworkDropsSent":     d.NetworkDropsSent,
		"vNetworkDropsSentT":    d.NetworkDropsSentT,
	}
}
