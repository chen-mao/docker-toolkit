package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config/image"
	"github.com/XDXCT/xdxct-container-toolkit/internal/info/drm"
	"github.com/XDXCT/xdxct-container-toolkit/internal/info/proc"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup/root"
)

// NewDRMNodesDiscoverer returns a discoverrer for the DRM device nodes associated with the specified visible devices.
//
// TODO: The logic for creating DRM devices should be consolidated between this
// and the logic for generating CDI specs for a single device. This is only used
// when applying OCI spec modifications to an incoming spec in "legacy" mode.
func NewDRMNodesDiscoverer(logger logger.Interface, devices image.VisibleDevices, devRoot string, xdxctCTKPath string) (Discover, error) {
	drmDeviceNodes, err := newDRMDeviceDiscoverer(logger, devices, devRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create DRM device discoverer: %v", err)
	}

	drmByPathSymlinks := newCreateDRMByPathSymlinks(logger, drmDeviceNodes, devRoot, xdxctCTKPath)

	discover := Merge(drmDeviceNodes, drmByPathSymlinks)
	return discover, nil
}

// NewGraphicsMountsDiscoverer creates a discoverer for the mounts required by graphics tools such as vulkan.
func NewGraphicsMountsDiscoverer(logger logger.Interface, driver *root.Driver, xdxctCTKPath string) (Discover, error) {
	// pciMount := newMounts(
	// 	logger,
	// 	lookup.NewFileLocator(
	// 		lookup.WithLogger(logger),
	// 		lookup.WithRoot(driver.Root),
	// 		lookup.WithSearchPaths(
	// 			"/usr/lib/x86_64-linux-gnu",
	// 			"/usr/lib64",
	// 		),
	// 		lookup.WithCount(1),
	// 	),
	// 	driver.Root,
	// 	[]string{
	// 		"libpciaccess.so.0",
	// 	},
	// )

	xorg := optionalXorgDiscoverer(logger, driver, xdxctCTKPath)

	discover := Merge(
		// pciMount,
		xorg,
	)

	return discover, nil
}

type drmDevicesByPath struct {
	None
	logger        logger.Interface
	xdxctCTKPath string
	devRoot       string
	devicesFrom   Discover
}

// newCreateDRMByPathSymlinks creates a discoverer for a hook to create the by-path symlinks for DRM devices discovered by the specified devices discoverer
func newCreateDRMByPathSymlinks(logger logger.Interface, devices Discover, devRoot string, xdxctCTKPath string) Discover {
	d := drmDevicesByPath{
		logger:        logger,
		xdxctCTKPath: xdxctCTKPath,
		devRoot:       devRoot,
		devicesFrom:   devices,
	}

	return &d
}

// Hooks returns a hook to create the symlinks from the required CSV files
func (d drmDevicesByPath) Hooks() ([]Hook, error) {
	devices, err := d.devicesFrom.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to discover devices for by-path symlinks: %v", err)
	}
	if len(devices) == 0 {
		return nil, nil
	}
	links, err := d.getSpecificLinkArgs(devices)
	if err != nil {
		return nil, fmt.Errorf("failed to determine specific links: %v", err)
	}
	if len(links) == 0 {
		return nil, nil
	}

	var args []string
	for _, l := range links {
		args = append(args, "--link", l)
	}

	hook := CreateXdxctCTKHook(
		d.xdxctCTKPath,
		"create-symlinks",
		args...,
	)

	return []Hook{hook}, nil
}

// getSpecificLinkArgs returns the required specic links that need to be created
func (d drmDevicesByPath) getSpecificLinkArgs(devices []Device) ([]string, error) {
	selectedDevices := make(map[string]bool)
	for _, d := range devices {
		selectedDevices[filepath.Base(d.HostPath)] = true
	}

	linkLocator := lookup.NewFileLocator(
		lookup.WithLogger(d.logger),
		lookup.WithRoot(d.devRoot),
	)
	candidates, err := linkLocator.Locate("/dev/dri/by-path/pci-*-*")
	if err != nil {
		d.logger.Warningf("Failed to locate by-path links: %v; ignoring", err)
		return nil, nil
	}

	var links []string
	for _, c := range candidates {
		device, err := os.Readlink(c)
		if err != nil {
			d.logger.Warningf("Failed to evaluate symlink %v; ignoring", c)
			continue
		}

		if selectedDevices[filepath.Base(device)] {
			d.logger.Debugf("adding device symlink %v -> %v", c, device)
			links = append(links, fmt.Sprintf("%v::%v", device, c))
		}
	}

	return links, nil
}

// newDRMDeviceDiscoverer creates a discoverer for the DRM devices associated with the requested devices.
func newDRMDeviceDiscoverer(logger logger.Interface, devices image.VisibleDevices, devRoot string) (Discover, error) {
	allDevices := NewCharDeviceDiscoverer(
		logger,
		devRoot,
		[]string{
			"/dev/dri/card*",
			"/dev/dri/renderD*",
		},
	)

	filter, err := newDRMDeviceFilter(logger, devices, devRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to construct DRM device filter: %v", err)
	}

	// We return a discoverer that applies the DRM device filter created above to all discovered DRM device nodes.
	d := newFilteredDisoverer(
		logger,
		allDevices,
		filter,
	)

	return d, err
}

// newDRMDeviceFilter creates a filter that matches DRM devices nodes for the visible devices.
func newDRMDeviceFilter(logger logger.Interface, devices image.VisibleDevices, devRoot string) (Filter, error) {
	gpuInformationPaths, err := proc.GetInformationFilePaths(devRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read GPU information: %v", err)
	}

	var selectedBusIds []string
	for _, f := range gpuInformationPaths {
		info, err := proc.ParseGPUInformationFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %v: %v", f, err)
		}
		uuid := info[proc.GPUInfoGPUUUID]
		busID := info[proc.GPUInfoBusLocation]
		minor := info[proc.GPUInfoDeviceMinor]

		if devices.Has(minor) || devices.Has(uuid) || devices.Has(busID) {
			selectedBusIds = append(selectedBusIds, busID)
		}
	}

	filter := make(selectDeviceByPath)
	for _, busID := range selectedBusIds {
		drmDeviceNodes, err := drm.GetDeviceNodesByBusID(busID)
		if err != nil {
			return nil, fmt.Errorf("failed to determine DRM devices for %v: %v", busID, err)
		}
		for _, drmDeviceNode := range drmDeviceNodes {
			filter[drmDeviceNode] = true
		}
	}

	return filter, nil
}

type xorgHooks struct {
	libraries     Discover
	driverVersion string
	xdxctCTKPath string
}

var _ Discover = (*xorgHooks)(nil)

// optionalXorgDiscoverer creates a discoverer for Xorg libraries.
// If the creation of the discoverer fails, a None discoverer is returned.
func optionalXorgDiscoverer(logger logger.Interface, driver *root.Driver, xdxctCTKPath string) Discover {
	xorg, err := newXorgDiscoverer(logger, driver, xdxctCTKPath)
	if err != nil {
		logger.Warningf("Failed to create Xorg discoverer: %v; skipping xorg libraries", err)
		return None{}
	}
	return xorg
}

func newXorgDiscoverer(logger logger.Interface, driver *root.Driver, xdxctCTKPath string) (Discover, error) {
	xorgLibs := NewMounts(
		logger,
		lookup.NewFileLocator(
			lookup.WithLogger(logger),
			lookup.WithRoot(driver.Root),
			lookup.WithSearchPaths(
				"/opt/xdxgpu/lib/xorg/modules/drivers",
				"/usr/lib/x86_64-linux-gnu/dri",
				"/usr/lib/aarch64-linux-gnu/dri",
				"/usr/lib64/xorg/modules/drivers",
				"/usr/lib/xorg/modules/drivers",
				"/usr/lib64/dri",
			),
		),
		driver.Root,
		[]string{
			"xdxgpu_dri.so",
			"xdxgpu_drv.so",
			"xdxgpu_drv_*.so",
		},
	)
	version := "155"
	xorgHooks := xorgHooks{
		libraries:     xorgLibs,
		driverVersion: version,
		xdxctCTKPath: xdxctCTKPath,
	}

	xorgConfg := NewMounts(
		logger,
		lookup.NewFileLocator(
			lookup.WithLogger(logger),
			lookup.WithRoot(driver.Root),
			lookup.WithSearchPaths("/usr/share"),
		),
		driver.Root,
		[]string{"X11/xorg.conf.d/10-xdxgpu.conf"},
	)

	d := Merge(
		xorgLibs,
		xorgConfg,
		xorgHooks,
	)

	return d, nil
}

// Devices returns no devices for Xorg
func (m xorgHooks) Devices() ([]Device, error) {
	return nil, nil
}

// Hooks returns a hook to create symlinks for Xorg libraries
func (m xorgHooks) Hooks() ([]Hook, error) {
	mounts, err := m.libraries.Mounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get mounts: %v", err)
	}
	if len(mounts) == 0 {
		return nil, nil
	}

	var target string
	for _, mount := range mounts {
		filename := filepath.Base(mount.HostPath)
		if filename == "libglxserver_xdxct.so."+m.driverVersion {
			target = mount.Path
		}
	}

	if target == "" {
		return nil, nil
	}

	link := strings.TrimSuffix(target, "."+m.driverVersion)
	links := []string{fmt.Sprintf("%s::%s", filepath.Base(target), link)}
	symlinkHook := CreateCreateSymlinkHook(
		m.xdxctCTKPath,
		links,
	)

	return symlinkHook.Hooks()
}

// Mounts returns the libraries required for Xorg
func (m xorgHooks) Mounts() ([]Mount, error) {
	return nil, nil
}

// selectDeviceByPath is a filter that allows devices to be selected by the path
type selectDeviceByPath map[string]bool

var _ Filter = (*selectDeviceByPath)(nil)

// DeviceIsSelected determines whether the device's path has been selected
func (s selectDeviceByPath) DeviceIsSelected(device Device) bool {
	return s[device.Path]
}

// MountIsSelected is always true
func (s selectDeviceByPath) MountIsSelected(Mount) bool {
	return true
}

// HookIsSelected is always true
func (s selectDeviceByPath) HookIsSelected(Hook) bool {
	return true
}
