package collectors

import "InferenceProfiler/pkg/collectors/types"

// BaseStatic contains base static system information.
type BaseStatic struct {
	UUID     string
	VMID     string
	Hostname string
	BootTime int64
}

// ToRecord converts BaseStatic to a Record.
func (b *BaseStatic) ToRecord() types.Record {
	return types.Record{
		"uuid":      b.UUID,
		"vId":       b.VMID,
		"vHostname": b.Hostname,
		"vBootTime": b.BootTime,
	}
}

// BaseDynamic contains base dynamic metrics.
type BaseDynamic struct {
	Timestamp int64
}

// ToRecord converts BaseDynamic to a Record.
func (b *BaseDynamic) ToRecord() types.Record {
	return types.Record{
		"timestamp": b.Timestamp,
	}
}
