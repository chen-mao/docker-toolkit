package devices

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/info/proc/devices"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

// Option is a function that sets an option on the Interface struct.
type Option func(*Interface)

// WithDryRun sets the dry run option for the Interface struct.
func WithDryRun(dryRun bool) Option {
	return func(i *Interface) {
		i.dryRun = dryRun
	}
}

// WithLogger sets the logger for the Interface struct.
func WithLogger(logger logger.Interface) Option {
	return func(i *Interface) {
		i.logger = logger
	}
}

// WithDevRoot sets the root directory for the XDXCT device nodes.
func WithDevRoot(devRoot string) Option {
	return func(i *Interface) {
		i.devRoot = devRoot
	}
}

// WithDevices sets the devices for the Interface struct.
func WithDevices(devices devices.Devices) Option {
	return func(i *Interface) {
		i.Devices = devices
	}
}
