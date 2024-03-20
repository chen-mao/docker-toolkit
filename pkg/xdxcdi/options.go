package xdxcdi

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/device"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxml"
)

// Option is a function that configures the xdxcdilib
type Option func(*xdxcdilib)

// WithDeviceLib sets the device library for the library
func WithDeviceLib(devicelib device.Interface) Option {
	return func(l *xdxcdilib) {
		l.devicelib = devicelib
	}
}

// WithDriverRoot sets the driver root for the library
func WithDriverRoot(root string) Option {
	return func(l *xdxcdilib) {
		l.driverRoot = root
	}
}

// WithDevRoot sets the root where /dev is located.
func WithDevRoot(root string) Option {
	return func(l *xdxcdilib) {
		l.devRoot = root
	}
}

// WithLogger sets the logger for the library
func WithLogger(logger logger.Interface) Option {
	return func(l *xdxcdilib) {
		l.logger = logger
	}
}

// WithXDXCTCTKPath sets the path to the XDXCT Container Toolkit CLI path for the library
func WithXDXCTCTKPath(path string) Option {
	return func(l *xdxcdilib) {
		l.xdxctCTKPath = path
	}
}

// WithCSVIgnorePatterns sets the ignore patterns for entries in the CSV files.
func WithCSVIgnorePatterns(csvIgnorePatterns []string) Option {
	return func(o *xdxcdilib) {
		o.csvIgnorePatterns = csvIgnorePatterns
	}
}

func WithXdxmlLib(xdxmllib xdxml.Interface) Option {
	return func(l *xdxcdilib) {
		l.xdxmllib = xdxmllib
	}
}

// WithMode sets the discovery mode for the library
func WithMode(mode string) Option {
	return func(l *xdxcdilib) {
		l.mode = mode
	}
}

// WithVendor sets the vendor for the library
func WithVendor(vendor string) Option {
	return func(o *xdxcdilib) {
		o.vendor = vendor
	}
}

// WithClass sets the class for the library
func WithClass(class string) Option {
	return func(o *xdxcdilib) {
		o.class = class
	}
}

// WithMergedDeviceOptions sets the merged device options for the library
// If these are not set, no merged device will be generated.
func WithMergedDeviceOptions(opts ...transform.MergedDeviceOption) Option {
	return func(o *xdxcdilib) {
		o.mergedDeviceOptions = opts
	}
}

// WithCSVFiles sets the CSV files for the library
func WithCSVFiles(csvFiles []string) Option {
	return func(o *xdxcdilib) {
		o.csvFiles = csvFiles
	}
}

// WithLibrarySearchPaths sets the library search paths.
// This is currently only used for CSV-mode.
func WithLibrarySearchPaths(paths []string) Option {
	return func(o *xdxcdilib) {
		o.librarySearchPaths = paths
	}
}
