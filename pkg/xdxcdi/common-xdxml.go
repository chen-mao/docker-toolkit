package xdxcdi

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
)

// newCommonNVMLDiscoverer returns a discoverer for entities that are not associated with a specific CDI device.
// This includes driver libraries and meta devices, for example.
func (l *xdxmllib) newCommonXDXMLDiscoverer() (discover.Discover, error) {
	// pyMounts := discover.NewCharDeviceDiscoverer(
	// 	l.logger,
	// 	"/usr/lib/python3/dist-packages/xdxsmi",
	// 	[]string{
	// 		"/usr/lib/python3/dist-packages/xdxsmi",
	// 	},
	// )

	graphicsMounts, err := discover.NewGraphicsMountsDiscoverer(l.logger, l.driver, l.xdxctCTKPath)
	if err != nil {
		l.logger.Warningf("failed to create discoverer for graphics mounts: %v", err)
	}

	driverFiles, err := NewDriverDiscoverer(l.logger, l.driver, l.xdxctCTKPath, l.xdxmllib)
	if err != nil {
		return nil, fmt.Errorf("failed to create discoverer for driver files: %v", err)
	}

	d := discover.Merge(
		// pyMounts,
		graphicsMounts,
		driverFiles,
	)

	return d, nil
}
