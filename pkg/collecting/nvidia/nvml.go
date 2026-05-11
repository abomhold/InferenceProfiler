package nvidia

import (
	"InferenceProfiler/pkg/utils"
	"errors"
	"log"
	"sync"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

var (
	nvmlMu       sync.Mutex
	nvmlRefCount int
)

type NVML struct {
	devices []nvml.Device
}

func NewNVML() (*NVML, error) {
	nvmlMu.Lock()
	defer nvmlMu.Unlock()

	utils.Debugf("nvml: initializing, refcount=%d", nvmlRefCount)

	if nvmlRefCount == 0 {
		if ret := nvml.Init(); ret != nvml.SUCCESS {
			utils.Debugf("nvml: Init() failed: %v", ret)
			return nil, ret
		}
	}
	nvmlRefCount++

	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) || count == 0 {
		nvmlRefCount--
		if nvmlRefCount == 0 {
			nvml.Shutdown()
		}
		utils.Debugf("nvml: no GPUs found (count=%d ret=%v)", count, ret)
		return nil, errors.New("no GPUs found")
	}

	devices := make([]nvml.Device, 0, count)
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret == nvml.SUCCESS {
			devices = append(devices, device)
		} else {
			utils.Debugf("nvml: device %d handle failed: %v", i, ret)
		}
	}

	log.Printf("nvidia: found %d GPU(s)", len(devices))

	return &NVML{
		devices: devices,
	}, nil
}

func (n *NVML) Devices() []nvml.Device {
	return n.devices
}

func (n *NVML) Count() int {
	return len(n.devices)
}

func (n *NVML) Close() error {
	nvmlMu.Lock()
	defer nvmlMu.Unlock()

	utils.Debugf("nvml: closing (%d devices, refcount=%d)", len(n.devices), nvmlRefCount)
	nvmlRefCount--
	if nvmlRefCount == 0 {
		if ret := nvml.Shutdown(); ret != nvml.SUCCESS {
			utils.Debugf("nvml: Shutdown() failed: %v", ret)
			return ret
		}
		utils.Debugf("nvml: shutdown complete")
	}
	return nil
}
