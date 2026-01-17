package vm

type NetworkInterface struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	State     string `json:"state"`
	MTU       int64  `json:"mtu"`
	SpeedMbps int64  `json:"speedMbps,omitempty"`
}

type NetworkStatic struct {
	NetworkInterfacesJSON string `json:"networkInterfaces,omitempty"`
}

type NetworkDynamic struct {
	NetworkBytesRecvd    int64 `json:"vNetworkBytesRecvd"`
	NetworkBytesRecvdT   int64 `json:"vNetworkBytesRecvdT"`
	NetworkBytesSent     int64 `json:"vNetworkBytesSent"`
	NetworkBytesSentT    int64 `json:"vNetworkBytesSentT"`
	NetworkPacketsRecvd  int64 `json:"vNetworkPacketsRecvd"`
	NetworkPacketsRecvdT int64 `json:"vNetworkPacketsRecvdT"`
	NetworkPacketsSent   int64 `json:"vNetworkPacketsSent"`
	NetworkPacketsSentT  int64 `json:"vNetworkPacketsSentT"`
	NetworkErrorsRecvd   int64 `json:"vNetworkErrorsRecvd"`
	NetworkErrorsRecvdT  int64 `json:"vNetworkErrorsRecvdT"`
	NetworkErrorsSent    int64 `json:"vNetworkErrorsSent"`
	NetworkErrorsSentT   int64 `json:"vNetworkErrorsSentT"`
	NetworkDropsRecvd    int64 `json:"vNetworkDropsRecvd"`
	NetworkDropsRecvdT   int64 `json:"vNetworkDropsRecvdT"`
	NetworkDropsSent     int64 `json:"vNetworkDropsSent"`
	NetworkDropsSentT    int64 `json:"vNetworkDropsSentT"`
}
