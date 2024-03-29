

package devices

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/internal/info/proc/devices"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

var errInvalidDeviceNode = errors.New("invalid device node")

// Interface provides a set of utilities for interacting with XDXCT devices on the system.
type Interface struct {
	devices.Devices

	logger logger.Interface

	dryRun bool
	// devRoot is the root directory where device nodes are expected to exist.
	devRoot string

	mknoder
}

// New constructs a new Interface struct with the specified options.
func New(opts ...Option) (*Interface, error) {
	i := &Interface{}
	for _, opt := range opts {
		opt(i)
	}

	if i.logger == nil {
		i.logger = logger.New()
	}
	if i.devRoot == "" {
		i.devRoot = "/"
	}
	if i.Devices == nil {
		devices, err := devices.GetXDXCTDevices()
		if err != nil {
			return nil, fmt.Errorf("failed to create devices info: %v", err)
		}
		i.Devices = devices
	}

	if i.dryRun {
		i.mknoder = &mknodLogger{i.logger}
	} else {
		i.mknoder = &mknodUnix{}
	}
	return i, nil
}

// createDeviceNode creates the specified device node with the require major and minor numbers.
// If a devRoot is configured, this is prepended to the path.
func (m *Interface) createDeviceNode(path string, major int, minor int) error {
	path = filepath.Join(m.devRoot, path)
	if _, err := os.Stat(path); err == nil {
		m.logger.Infof("Skipping: %s already exists", path)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat %s: %v", path, err)
	}

	return m.Mknode(path, major, minor)
}

// Major returns the major number for the specified XDXCT device node.
// If the device node is not supported, an error is returned.
func (m *Interface) Major(node string) (int64, error) {
	var valid bool
	var major devices.Major
	major, valid = m.Get(devices.XDXCTGPU)

	if valid {
		return int64(major), nil
	}

	return 0, errInvalidDeviceNode
}
