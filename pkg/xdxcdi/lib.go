package xdxcdi

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup/root"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/spec"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxml"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/device"
)

type wrapper struct {
	Interface

	vendor string
	class  string

	mergedDeviceOptions []transform.MergedDeviceOption
}

type xdxcdilib struct {
	logger             logger.Interface
	xdxmllib           xdxml.Interface
	mode               string
	devicelib          device.Interface
	deviceNamer        string
	driverRoot         string
	devRoot            string
	xdxctCTKPath       string
	librarySearchPaths []string

	csvFiles          []string
	csvIgnorePatterns []string

	vendor string
	class  string

	driver  *root.Driver

	mergedDeviceOptions []transform.MergedDeviceOption
}

// New creates a new xdxcdi library
func New(opts ...Option) (Interface, error) {
	l := &xdxcdilib{}
	for _, opt := range opts {
		opt(l)
	}
	if l.mode == "" {
		l.mode = ModeAuto
	}
	if l.logger == nil {
		l.logger = logger.New()
	}
	if l.deviceNamer == "" {
		l.deviceNamer = "GPU"
	}
	if l.driverRoot == "" {
		l.driverRoot = "/"
	}
	if l.devRoot == "" {
		l.devRoot = l.driverRoot
	}
	if l.xdxctCTKPath == "" {
		l.xdxctCTKPath = "/usr/bin/xdxct-ctk"
	}

	// TODO: We need to improve the construction of this driver root.
	l.driver = root.New(l.logger, l.driverRoot, l.librarySearchPaths)

	var lib Interface
	switch l.resolveMode() {
	case ModeCSV:
		// CSV is used to support tegra device.
		l.logger.Info("Now we not support CSV Mode.")
	case ModeManagement:
		l.logger.Info("Now we not support Management Mode.")
	case ModeXdxml:
		// TODO xdxml
		if l.xdxmllib == nil {
			l.xdxmllib = xdxml.New()
		}
		if l.devicelib == nil {
			l.devicelib = device.New(device.WithXdxml(l.xdxmllib))
		}

		lib = (*xdxmllib)(l)
	case ModeWsl:
		l.logger.Info("Now we not support WSL Mode.")
	case ModeMofed:
		// Mofed is used to support InfiniBand network.
		l.logger.Info("Now we not support WSL Mode.")
	default:
		return nil, fmt.Errorf("unknown mode %q", l.mode)
	}

	w := wrapper{
		Interface:           lib,
		vendor:              l.vendor,
		class:               l.class,
		mergedDeviceOptions: l.mergedDeviceOptions,
	}
	return &w, nil
}

// GetSpec combines the device specs and common edits from the wrapped Interface to a single spec.Interface.
func (l *wrapper) GetSpec() (spec.Interface, error) {
	deviceSpecs, err := l.GetAllDeviceSpecs()
	if err != nil {
		return nil, err
	}

	edits, err := l.GetCommonEdits()
	if err != nil {
		return nil, err
	}

	return spec.New(
		spec.WithDeviceSpecs(deviceSpecs),
		spec.WithEdits(*edits.ContainerEdits),
		spec.WithVendor(l.vendor),
		spec.WithClass(l.class),
		spec.WithMergedDeviceOptions(l.mergedDeviceOptions...),
	)
}

// resolveMode resolves the mode for CDI spec generation based on the current system.
func (l *xdxcdilib) resolveMode() (rmode string) {
	if l.mode != ModeAuto {
		return l.mode
	}
	defer func() {
		l.logger.Infof("Auto-detected mode as %q", rmode)
	}()

	return ModeXdxml
}
