package xdxcdi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
	"github.com/XDXCT/xdxct-container-toolkit/internal/edits"
	"github.com/XDXCT/xdxct-container-toolkit/internal/info/drm"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/device"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxml"
	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

// GetGPUDeviceSpecs returns the CDI device specs for the full GPU represented by 'device'.
func (l *xdxmllib) GetGPUDeviceSpecs(i int, d device.Device) (*specs.Device, error) {
	edits, err := l.GetGPUDeviceEdits(d)
	if err != nil {
		return nil, fmt.Errorf("failed to get edits for device: %v", err)
	}

	name := "gpu"
	if err != nil {
		return nil, fmt.Errorf("failed to get device name: %v", err)
	}

	spec := specs.Device{
		Name:           name,
		ContainerEdits: *edits.ContainerEdits,
	}

	return &spec, nil
}

// GetGPUDeviceEdits returns the CDI edits for the full GPU represented by 'device'.
func (l *xdxmllib) GetGPUDeviceEdits(d device.Device) (*cdi.ContainerEdits, error) {
	device, err := newFullGPUDiscoverer(l.logger, l.devRoot, l.xdxctCTKPath, d)
	if err != nil {
		return nil, fmt.Errorf("failed to create device discoverer: %v", err)
	}

	editsForDevice, err := edits.FromDiscoverer(device)
	if err != nil {
		return nil, fmt.Errorf("failed to create container edits for device: %v", err)
	}

	return editsForDevice, nil
}

// byPathHookDiscoverer discovers the entities required for injecting by-path DRM device links
type byPathHookDiscoverer struct {
	logger        logger.Interface
	devRoot       string
	xdxctCTKPath string
	pciBusID      string
	deviceNodes   discover.Discover
}

var _ discover.Discover = (*byPathHookDiscoverer)(nil)

// newFullGPUDiscoverer creates a discoverer for the full GPU defined by the specified device.
func newFullGPUDiscoverer(logger logger.Interface, devRoot string, xdxctCTKPath string, d device.Device) (discover.Discover, error) {
	// In xdxgpu driver, get deviceNodePaths by PciInfo
	pciInfo, ret := d.GetPciInfo()
	if ret != xdxml.SUCCESS {
		return nil, fmt.Errorf("error getting PCI info for device: %v", ret)
	}
	pciBusID := getBusID(pciInfo)

	drmDeviceNodes, err := drm.GetDeviceNodesByBusID(pciBusID)
	if err != nil {
		return nil, fmt.Errorf("failed to determine DRM devices for %v: %v", pciBusID, err)
	}

	// deviceNodePaths := append([]string{path}, drmDeviceNodes...)
	deviceNodePaths := drmDeviceNodes

	deviceNodes := discover.NewCharDeviceDiscoverer(
		logger,
		devRoot,
		deviceNodePaths,
	)

	byPathHooks := &byPathHookDiscoverer{
		logger:        logger,
		devRoot:       devRoot,
		xdxctCTKPath: xdxctCTKPath,
		pciBusID:      pciBusID,
		deviceNodes:   deviceNodes,
	}

	deviceFolderPermissionHooks := newDeviceFolderPermissionHookDiscoverer(
		logger,
		devRoot,
		xdxctCTKPath,
		deviceNodes,
	)

	dd := discover.Merge(
		deviceNodes,
		byPathHooks,
		deviceFolderPermissionHooks,
	)

	return dd, nil
}

// Devices returns the empty list for the by-path hook discoverer
func (d *byPathHookDiscoverer) Devices() ([]discover.Device, error) {
	return nil, nil
}

// Hooks returns the hooks for the GPU device.
// The following hooks are detected:
//  1. A hook to create /dev/dri/by-path symlinks
func (d *byPathHookDiscoverer) Hooks() ([]discover.Hook, error) {
	links, err := d.deviceNodeLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to discover DRA device links: %v", err)
	}
	if len(links) == 0 {
		return nil, nil
	}

	var args []string
	for _, l := range links {
		args = append(args, "--link", l)
	}

	hook := discover.CreateXdxctCTKHook(
		d.xdxctCTKPath,
		"create-symlinks",
		args...,
	)

	return []discover.Hook{hook}, nil
}

// Mounts returns an empty slice for a full GPU
func (d *byPathHookDiscoverer) Mounts() ([]discover.Mount, error) {
	return nil, nil
}

func (d *byPathHookDiscoverer) deviceNodeLinks() ([]string, error) {
	devices, err := d.deviceNodes.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to discover device nodes: %v", err)
	}

	if len(devices) == 0 {
		return nil, nil
	}

	selectedDevices := make(map[string]bool)
	for _, d := range devices {
		selectedDevices[d.HostPath] = true
	}

	candidates := []string{
		fmt.Sprintf("/dev/dri/by-path/pci-%s-card", d.pciBusID),
		fmt.Sprintf("/dev/dri/by-path/pci-%s-render", d.pciBusID),
	}

	var links []string
	for _, c := range candidates {
		linkPath := filepath.Join(d.devRoot, c)
		device, err := os.Readlink(linkPath)
		if err != nil {
			d.logger.Warningf("Failed to evaluate symlink %v; ignoring", linkPath)
			continue
		}

		deviceNode := device
		if !filepath.IsAbs(device) {
			deviceNode = filepath.Join(filepath.Dir(linkPath), device)
		}
		if !selectedDevices[deviceNode] {
			d.logger.Debugf("ignoring device symlink %v -> %v since %v is not mounted", linkPath, device, deviceNode)
			continue
		}
		d.logger.Debugf("adding device symlink %v -> %v", linkPath, device)
		links = append(links, fmt.Sprintf("%v::%v", device, linkPath))
	}

	return links, nil
}

// getBusID provides a utility function that returns the string representation of the bus ID.
func getBusID(p xdxml.PciInfo) string {
	busStr := fmt.Sprintf("%02x", p.Bus)
	deviceStr := fmt.Sprintf("%02x", p.Device)
	funcStr := fmt.Sprintf("%1x", p.Func)

	id := fmt.Sprintf("0000:%s:%s.%s", busStr, deviceStr, funcStr)
	id = strings.ToLower(id)
	return id
}
