package domains

import (
	"InferenceProfiler/pkg/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type DeviceStatic struct {
	Index                  int
	Name                   string
	UUID                   string
	Serial                 string
	ComputeCapabilityMajor int
	ComputeCapabilityMinor int
	DriverVersion          string
	CudaVersion            int
}

func CollectDeviceStatic(d nvml.Device, index int, s *DeviceStatic) {
	s.Index = index

	if name, ret := d.GetName(); ret == nvml.SUCCESS {
		s.Name = name
	} else {
		utils.Debugf("nvidia/device[%d]: GetName failed: %v", index, ret)
	}

	if uuid, ret := d.GetUUID(); ret == nvml.SUCCESS {
		s.UUID = uuid
	} else {
		utils.Debugf("nvidia/device[%d]: GetUUID failed: %v", index, ret)
	}

	if serial, ret := d.GetSerial(); ret == nvml.SUCCESS {
		s.Serial = serial
	} else {
		utils.Debugf("nvidia/device[%d]: GetSerial failed: %v", index, ret)
	}

	if major, minor, ret := d.GetCudaComputeCapability(); ret == nvml.SUCCESS {
		s.ComputeCapabilityMajor = major
		s.ComputeCapabilityMinor = minor
	} else {
		utils.Debugf("nvidia/device[%d]: GetCudaComputeCapability failed: %v", index, ret)
	}

	if version, ret := nvml.SystemGetDriverVersion(); ret == nvml.SUCCESS {
		s.DriverVersion = version
	} else {
		utils.Debugf("nvidia/device[%d]: SystemGetDriverVersion failed: %v", index, ret)
	}

	if cuda, ret := nvml.SystemGetCudaDriverVersion(); ret == nvml.SUCCESS {
		s.CudaVersion = cuda
	} else {
		utils.Debugf("nvidia/device[%d]: SystemGetCudaDriverVersion failed: %v", index, ret)
	}

	utils.Debugf("nvidia/device[%d]: name=%s uuid=%s cc=%d.%d driver=%s",
		index, s.Name, s.UUID, s.ComputeCapabilityMajor, s.ComputeCapabilityMinor, s.DriverVersion)
}
