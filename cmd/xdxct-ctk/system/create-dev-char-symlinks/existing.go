package devchar

import (
	"path/filepath"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
	"golang.org/x/sys/unix"
)

type nodeLister interface {
	DeviceNodes() ([]deviceNode, error)
}

type existing struct {
	logger  logger.Interface
	devRoot string
}

func (m existing) DeviceNodes() ([]deviceNode, error) {
	locator := lookup.NewCharDeviceLocator(
		lookup.WithLogger(m.logger),
		lookup.WithRoot(m.devRoot),
		lookup.WithOptional(true),
	)

	devices, err := locator.Locate("/dev/dri/xdxct*")
	if err != nil {
		m.logger.Warningf("Error while locating device: %v", err)
	}

	// It's for cap devices
	capDevices, err := locator.Locate("/dev/xdxct-caps/xdxct-*")
	if err != nil {
		m.logger.Warningf("Error while locating caps device: %v", err)
	}

	if len(devices) == 0 && len(capDevices) == 0 {
		m.logger.Infof("No devices found in %s", m.devRoot)
		return nil, nil
	}

	var deviceNodes []deviceNode
	for _, d := range append(devices, capDevices...) {
		if m.nodeIsBlocked(d) {
			continue
		}
		var stat unix.Stat_t
		err := unix.Stat(d, &stat)
		if err != nil {
			m.logger.Warningf("Could not stat device: %v", err)
			continue
		}
		deviceNodes = append(deviceNodes, newDeviceNode(d, stat))
	}

	return deviceNodes, nil
}

// nodeIsBlocked returns true if the specified device node should be ignored.
func (m existing) nodeIsBlocked(path string) bool {
	blockedPrefixes := []string{"xdxct-null"}
	nodeName := filepath.Base(path)
	for _, prefix := range blockedPrefixes {
		if strings.HasPrefix(nodeName, prefix) {
			return true
		}
	}
	return false
}
