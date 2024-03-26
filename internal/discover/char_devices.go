package discover

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
)

// charDevices is a discover for a list of character devices
type charDevices mounts

var _ Discover = (*charDevices)(nil)

// NewCharDeviceDiscoverer creates a discoverer which locates the specified set of device nodes.
func NewCharDeviceDiscoverer(logger logger.Interface, devRoot string, devices []string) Discover {
	locator := lookup.NewCharDeviceLocator(
		lookup.WithLogger(logger),
		lookup.WithRoot(devRoot),
	)

	return (*charDevices)(newMounts(logger, locator, devRoot, devices))
}

// Mounts returns the discovered mounts for the charDevices.
// Since this explicitly specifies a device list, the mounts are nil.
func (d *charDevices) Mounts() ([]Mount, error) {
	return nil, nil
}

// Devices returns the discovered devices for the charDevices.
// Here the device nodes are first discovered as mounts and these are converted to devices.
func (d *charDevices) Devices() ([]Device, error) {
	devicesAsMounts, err := (*mounts)(d).Mounts()
	if err != nil {
		return nil, err
	}
	var devices []Device
	for _, mount := range devicesAsMounts {
		device := Device{
			HostPath: mount.HostPath,
			Path:     mount.Path,
		}
		devices = append(devices, device)
	}

	return devices, nil
}
