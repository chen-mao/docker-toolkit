package devchar

import (
	"fmt"
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/internal/info/proc/devices"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxpci"
)

type allPossible struct {
	logger       logger.Interface
	devRoot      string
	deviceMajors devices.Devices
}

// newAllPossible returns a new allPossible device node lister.
// This lister lists all possible device nodes for XDXCT GPUs, control devices, and capability devices.
func newAllPossible(logger logger.Interface, devRoot string) (nodeLister, error) {
	deviceMajors, err := devices.GetXDXCTDevices()
	if err != nil {
		return nil, fmt.Errorf("failed reading device majors: %v", err)
	}

	var requiredMajors []devices.Name

	for _, name := range requiredMajors {
		if !deviceMajors.Exists(name) {
			return nil, fmt.Errorf("missing required device major %s", name)
		}
	}

	l := allPossible{
		logger:       logger,
		devRoot:      devRoot,
		deviceMajors: deviceMajors,
	}

	return l, nil
}

// DeviceNodes returns a list of all possible device nodes for XDXCT GPUs, control devices, and capability devices.
func (m allPossible) DeviceNodes() ([]deviceNode, error) {
	gpus, err := xdxpci.New(
		xdxpci.WithPCIDevicesRoot(filepath.Join(m.devRoot, xdxpci.PCIDevicesRoot)),
		xdxpci.WithLogger(m.logger),
	).GetGPUs()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU information: %v", err)
	}

	count := len(gpus)
	if count == 0 {
		m.logger.Infof("No XDXCT devices found in %s", m.devRoot)
		return nil, nil
	}

	var deviceNodes []deviceNode

	if err != nil {
		return nil, fmt.Errorf("failed to get control device nodes: %v", err)
	}

	for gpu := 0; gpu < count; gpu++ {
		deviceNodes = append(deviceNodes, m.getGPUDeviceNodes(gpu)...)
	}

	return deviceNodes, nil
}

// getGPUDeviceNodes generates a list of device nodes for a given GPU.
func (m allPossible) getGPUDeviceNodes(gpu int) []deviceNode {
	d := m.newDeviceNode(
		devices.XDXCTGPU,
		fmt.Sprintf("/dev/dri/xdxct%d", gpu),
		gpu,
	)

	return []deviceNode{d}
}

// newDeviceNode creates a new device node with the specified path and major/minor numbers.
// The path is adjusted for the specified driver root.
func (m allPossible) newDeviceNode(deviceName devices.Name, path string, minor int) deviceNode {
	major, _ := m.deviceMajors.Get(deviceName)

	return deviceNode{
		path:  filepath.Join(m.devRoot, path),
		major: uint32(major),
		minor: uint32(minor),
	}
}
