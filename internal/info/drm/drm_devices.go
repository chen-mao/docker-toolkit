package drm

import (
	"fmt"
	"path/filepath"
)

// GetDeviceNodesByBusID returns the DRM devices associated with the specified PCI bus ID
func GetDeviceNodesByBusID(busID string) ([]string, error) {
	drmRoot := filepath.Join("/sys/bus/pci/devices", busID, "drm")
	matches_card, err := filepath.Glob(fmt.Sprintf("%s/card*", drmRoot))
	if err != nil {
		return nil, err
	}
	matches_renderD, err := filepath.Glob(fmt.Sprintf("%s/renderD*", drmRoot))
	if err != nil {
		return nil, err
	}

	matches := append(matches_card, matches_renderD...)
	var drmDeviceNodes []string
	for _, m := range matches {
		drmDeviceNode := filepath.Join("/dev/dri", filepath.Base(m))

		drmDeviceNodes = append(drmDeviceNodes, drmDeviceNode)
	}

	return drmDeviceNodes, nil
}
