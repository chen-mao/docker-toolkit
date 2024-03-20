package discover

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

type gdsDeviceDiscoverer struct {
	None
	logger  logger.Interface
	devices Discover
	mounts  Discover
}

// NewGDSDiscoverer creates a discoverer for GPUDirect Storage devices and mounts.
func NewGDSDiscoverer(logger logger.Interface, driverRoot string, devRoot string) (Discover, error) {
	devices := NewCharDeviceDiscoverer(
		logger,
		devRoot,
		[]string{"/dev/xdxct-fs*"},
	)

	udev := NewMounts(
		logger,
		lookup.NewDirectoryLocator(lookup.WithLogger(logger), lookup.WithRoot(driverRoot)),
		driverRoot,
		[]string{"/run/udev"},
	)

	cufile := NewMounts(
		logger,
		lookup.NewFileLocator(
			lookup.WithLogger(logger),
			lookup.WithRoot(driverRoot),
		),
		driverRoot,
		[]string{"/etc/cufile.json"},
	)

	d := gdsDeviceDiscoverer{
		logger:  logger,
		devices: devices,
		mounts:  Merge(udev, cufile),
	}

	return &d, nil
}

// Devices discovers the xdxct-fs device nodes for use with GPUDirect Storage
func (d *gdsDeviceDiscoverer) Devices() ([]Device, error) {
	return d.devices.Devices()
}

// Mounts discovers the required mounts for GPUDirect Storage.
// If no devices are discovered the discovered mounts are empty
func (d *gdsDeviceDiscoverer) Mounts() ([]Mount, error) {
	devices, err := d.Devices()
	if err != nil || len(devices) == 0 {
		d.logger.Debugf("No xdxct-fs devices detected; skipping detection of mounts")
		return nil, nil
	}

	return d.mounts.Mounts()
}
