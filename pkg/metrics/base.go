// Package metrics defines the data structures for system metrics.
package metrics

// Record is a generic map type for metric records.
type Record = map[string]interface{}

// BaseStatic contains base static system information.
type BaseStatic struct {
	UUID     string `json:"uuid"`
	VMID     string `json:"vId"`
	Hostname string `json:"vHostname"`
	BootTime int64  `json:"vBootTime"`
}

// BaseDynamic contains base dynamic metrics.
type BaseDynamic struct {
	Timestamp int64 `json:"timestamp"`
}

// Static combines all static metric types.
type Static struct {
	BaseStatic
	// Embedded static types from collectors will be added here
}

// Dynamic combines all dynamic metric types.
type Dynamic struct {
	BaseDynamic
	// Embedded dynamic types from collectors will be added here
}
