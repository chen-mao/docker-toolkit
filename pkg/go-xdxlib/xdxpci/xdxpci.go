/*
 * Copyright (c) 2024, NVIDIA CORPORATION and XDXCT CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xdxpci

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/pciids"
)

const (
	// PCIDevicesRoot represents base path for all pci devices under sysfs
	PCIDevicesRoot = "/sys/bus/pci/devices"
	// PCIXdxctVendorID represents PCI vendor id for XDXCT
	PCIXdxctVendorID uint16 = 0x1eed
	// UnknownDeviceString is the device name to set for devices not found in the PCI database
	// PCIVgaControllerClass represents the PCI class for VGA Controllers
	PCIVgaControllerClass uint32 = 0x030000
	// PCI3dControllerClass represents the PCI class for 3D Graphics accellerators
	PCI3dControllerClass uint32 = 0x030200
	// UnknownDeviceString is the device name to set for devices not found in the PCI database
	UnknownDeviceString = "UNKNOWN_DEVICE"
	// UnknownClassString is the class name to set for devices not found in the PCI database
	UnknownClassString = "UNKNOWN_CLASS"
)

// Interface allows us to get a list of all XDXCT PCI devices
type Interface interface {
	GetAllDevices() ([]*XdxctPCIDevice, error)
	GetGPUs() ([]*XdxctPCIDevice, error)
	GetGPUByIndex(int) (*XdxctPCIDevice, error)
	GetGPUByPciBusID(string) (*XdxctPCIDevice, error)
}

// MemoryResources a more human readable handle
type MemoryResources map[int]*MemoryResource

// ResourceInterface exposes some higher level functions of resources
type ResourceInterface interface {
	GetTotalAddressableMemory(bool) (uint64, uint64)
}

type xdxpci struct {
	logger         logger
	pciDevicesRoot string
	pcidbPath      string
}

var _ Interface = (*xdxpci)(nil)
var _ ResourceInterface = (*MemoryResources)(nil)

// XdxctPCIDevice represents a PCI device for an XDXCT product
type XdxctPCIDevice struct {
	Path       string
	Address    string
	Vendor     uint16
	Class      uint32
	ClassName  string
	Device     uint16
	DeviceName string
	Driver     string
	IommuGroup int
	NumaNode   int
	Config     *ConfigSpace
	Resources  MemoryResources
	IsVF       bool
}

// IsResetAvailable some devices can be reset without rebooting,
// check if applicable
func (d *XdxctPCIDevice) IsResetAvailable() bool {
	_, err := os.Stat(path.Join(d.Path, "reset"))
	return err == nil
}

// Reset perform a reset to apply a new configuration at HW level
func (d *XdxctPCIDevice) Reset() error {
	err := os.WriteFile(path.Join(d.Path, "reset"), []byte("1"), 0)
	if err != nil {
		return fmt.Errorf("unable to write to reset file: %v", err)
	}
	return nil
}

// New interface that allows us to get a list of all XDXCT PCI devices
func New(opts ...Option) Interface {
	n := &xdxpci{}
	for _, opt := range opts {
		opt(n)
	}
	if n.logger == nil {
		n.logger = &simpleLogger{}
	}
	if n.pciDevicesRoot == "" {
		n.pciDevicesRoot = PCIDevicesRoot
	}
	return n
}

// Option defines a function for passing options to the New() call
type Option func(*xdxpci)

// WithLogger provides an Option to set the logger for the library
func WithLogger(logger logger) Option {
	return func(n *xdxpci) {
		n.logger = logger
	}
}

// WithPCIDevicesRoot provides an Option to set the root path
// for PCI devices on the system.
func WithPCIDevicesRoot(root string) Option {
	return func(n *xdxpci) {
		n.pciDevicesRoot = root
	}
}

// WithPCIDatabasePath provides an Option to set the path
// to the pciids database file.
func WithPCIDatabasePath(path string) Option {
	return func(n *xdxpci) {
		n.pcidbPath = path
	}
}

// IsGPU either VGA for older cards or 3D for newer
func (d *XdxctPCIDevice) IsGPU() bool {
	return d.IsVGAController() || d.Is3DController()
}

// IsVGAController if class == 0x300
func (d *XdxctPCIDevice) IsVGAController() bool {
	return d.Class == PCIVgaControllerClass
}

// Is3DController if class == 0x302
func (d *XdxctPCIDevice) Is3DController() bool {
	return d.Class == PCI3dControllerClass
}

// GetAllDevices returns all Xdxct PCI devices on the system
func (p *xdxpci) GetAllDevices() ([]*XdxctPCIDevice, error) {
	deviceDirs, err := os.ReadDir(p.pciDevicesRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI bus devices: %v", err)
	}

	var xdxdevices []*XdxctPCIDevice
	for _, deviceDir := range deviceDirs {
		deviceAddress := deviceDir.Name()
		xdxdevice, err := p.GetGPUByPciBusID(deviceAddress)
		if err != nil {
			return nil, fmt.Errorf("error constructing XDXCT PCI device %s: %v", deviceAddress, err)
		}
		if xdxdevice == nil {
			continue
		}
		xdxdevices = append(xdxdevices, xdxdevice)
	}

	addressToID := func(address string) uint64 {
		address = strings.ReplaceAll(address, ":", "")
		address = strings.ReplaceAll(address, ".", "")
		id, _ := strconv.ParseUint(address, 16, 64)
		return id
	}

	sort.Slice(xdxdevices, func(i, j int) bool {
		return addressToID(xdxdevices[i].Address) < addressToID(xdxdevices[j].Address)
	})

	return xdxdevices, nil
}

// GetGPUByPciBusID constructs an XdxctPCIDevice for the specified address (PCI Bus ID)
func (p *xdxpci) GetGPUByPciBusID(address string) (*XdxctPCIDevice, error) {
	devicePath := filepath.Join(p.pciDevicesRoot, address)

	vendor, err := os.ReadFile(path.Join(devicePath, "vendor"))
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI device vendor id for %s: %v", address, err)
	}
	vendorStr := strings.TrimSpace(string(vendor))
	vendorID, err := strconv.ParseUint(vendorStr, 0, 16)
	if err != nil {
		return nil, fmt.Errorf("unable to convert vendor string to uint16: %v", vendorStr)
	}

	if uint16(vendorID) != PCIXdxctVendorID {
		return nil, nil
	}

	class, err := os.ReadFile(path.Join(devicePath, "class"))
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI device class for %s: %v", address, err)
	}
	classStr := strings.TrimSpace(string(class))
	classID, err := strconv.ParseUint(classStr, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("unable to convert class string to uint32: %v", classStr)
	}

	device, err := os.ReadFile(path.Join(devicePath, "device"))
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI device id for %s: %v", address, err)
	}
	deviceStr := strings.TrimSpace(string(device))
	deviceID, err := strconv.ParseUint(deviceStr, 0, 16)
	if err != nil {
		return nil, fmt.Errorf("unable to convert device string to uint16: %v", deviceStr)
	}

	driver, err := filepath.EvalSymlinks(path.Join(devicePath, "driver"))
	if err == nil {
		driver = filepath.Base(driver)
	} else if os.IsNotExist(err) {
		driver = ""
	} else {
		return nil, fmt.Errorf("unable to detect driver for %s: %v", address, err)
	}

	var iommuGroup int64
	iommu, err := filepath.EvalSymlinks(path.Join(devicePath, "iommu_group"))
	if err == nil {
		iommuGroupStr := strings.TrimSpace(filepath.Base(iommu))
		iommuGroup, err = strconv.ParseInt(iommuGroupStr, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to convert iommu_group string to int64: %v", iommuGroupStr)
		}
	} else if os.IsNotExist(err) {
		iommuGroup = -1
	} else {
		return nil, fmt.Errorf("unable to detect iommu_group for %s: %v", address, err)
	}

	// device is a virtual function (VF) if "physfn" symlink exists
	var isVF bool
	_, err = filepath.EvalSymlinks(path.Join(devicePath, "physfn"))
	if err == nil {
		isVF = true
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to resolve %s: %v", path.Join(devicePath, "physfn"), err)
	}

	numa, err := os.ReadFile(path.Join(devicePath, "numa_node"))
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI NUMA node for %s: %v", address, err)
	}
	numaStr := strings.TrimSpace(string(numa))
	numaNode, err := strconv.ParseInt(numaStr, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to convert NUMA node string to int64: %v", numaNode)
	}

	config := &ConfigSpace{
		Path: path.Join(devicePath, "config"),
	}

	resource, err := os.ReadFile(path.Join(devicePath, "resource"))
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI resource file for %s: %v", address, err)
	}

	resources := make(map[int]*MemoryResource)
	for i, line := range strings.Split(strings.TrimSpace(string(resource)), "\n") {
		values := strings.Split(line, " ")
		if len(values) != 3 {
			return nil, fmt.Errorf("more than 3 entries in line '%d' of resource file", i)
		}

		start, _ := strconv.ParseUint(values[0], 0, 64)
		end, _ := strconv.ParseUint(values[1], 0, 64)
		flags, _ := strconv.ParseUint(values[2], 0, 64)

		if (end - start) != 0 {
			resources[i] = &MemoryResource{
				uintptr(start),
				uintptr(end),
				flags,
				fmt.Sprintf("%s/resource%d", devicePath, i),
			}
		}
	}

	pciDB := pciids.NewDB()

	deviceName, err := pciDB.GetDeviceName(uint16(vendorID), uint16(deviceID))
	if err != nil {
		p.logger.Warningf("unable to get device name: %v\n", err)
		deviceName = UnknownDeviceString
	}
	className, err := pciDB.GetClassName(uint32(classID))
	if err != nil {
		p.logger.Warningf("unable to get class name for device: %v\n", err)
		className = UnknownClassString
	}

	nvdevice := &XdxctPCIDevice{
		Path:       devicePath,
		Address:    address,
		Vendor:     uint16(vendorID),
		Class:      uint32(classID),
		Device:     uint16(deviceID),
		Driver:     driver,
		IommuGroup: int(iommuGroup),
		NumaNode:   int(numaNode),
		Config:     config,
		Resources:  resources,
		IsVF:       isVF,
		DeviceName: deviceName,
		ClassName:  className,
	}

	return nvdevice, nil
}

// GetGPUs returns all XDXCT GPU devices on the system
func (p *xdxpci) GetGPUs() ([]*XdxctPCIDevice, error) {
	devices, err := p.GetAllDevices()
	if err != nil {
		return nil, fmt.Errorf("error getting all XDXCT devices: %v", err)
	}

	var filtered []*XdxctPCIDevice
	for _, d := range devices {
		if d.IsGPU() && !d.IsVF {
			filtered = append(filtered, d)
		}
	}

	return filtered, nil
}

// GetGPUByIndex returns an XDXCT GPU device at a particular index
func (p *xdxpci) GetGPUByIndex(i int) (*XdxctPCIDevice, error) {
	gpus, err := p.GetGPUs()
	if err != nil {
		return nil, fmt.Errorf("error getting all gpus: %v", err)
	}

	if i < 0 || i >= len(gpus) {
		return nil, fmt.Errorf("invalid index '%d'", i)
	}

	return gpus[i], nil
}
