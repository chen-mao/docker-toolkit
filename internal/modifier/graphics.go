package modifier

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config"
	"github.com/XDXCT/xdxct-container-toolkit/internal/config/image"
	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup/root"
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
)

// NewGraphicsModifier constructs a modifier that injects graphics-related modifications into an OCI runtime specification.
// The value of the XDXCT_DRIVER_CAPABILITIES environment variable is checked to determine if this modification should be made.
func NewGraphicsModifier(logger logger.Interface, cfg *config.Config, image image.GPU) (oci.SpecModifier, error) {
	if required, reason := requiresGraphicsModifier(image); !required {
		logger.Infof("No graphics modifier required: %v", reason)
		return nil, nil
	}

	// TODO: We should not just pass `nil` as the search path here.
	driver := root.New(logger, cfg.XDXCTContainerCLIConfig.Root, nil)
	xdxctCTKPath := cfg.XDXCTCTKConfig.Path

	mounts, err := discover.NewGraphicsMountsDiscoverer(
		logger,
		driver,
		xdxctCTKPath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mounts discoverer: %v", err)
	}

	// In standard usage, the devRoot is the same as the driver.Root.
	devRoot := driver.Root
	drmNodes, err := discover.NewDRMNodesDiscoverer(
		logger,
		image.DevicesFromEnvvars(visibleDevicesEnvvar),
		devRoot,
		xdxctCTKPath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct discoverer: %v", err)
	}

	d := discover.Merge(
		drmNodes,
		mounts,
	)
	return NewModifierFromDiscoverer(logger, d)
}

// requiresGraphicsModifier determines whether a graphics modifier is required.
func requiresGraphicsModifier(gpuImage image.GPU) (bool, string) {
	if devices := gpuImage.DevicesFromEnvvars(visibleDevicesEnvvar); len(devices.List()) == 0 {
		return false, "no devices requested"
	}

	if !gpuImage.GetDriverCapabilities().Any(image.DriverCapabilityGraphics, image.DriverCapabilityDisplay) {
		return false, "no required capabilities requested"
	}

	return true, ""
}
