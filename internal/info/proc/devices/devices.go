package devices

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Device major numbers and device names for XDXCT devices
const (
	XDXCTUVMMinor      = 0
	XDXCTUVMToolsMinor = 1
	XDXCTCTLMinor      = 226

	XDXCTFrontend     = Name("xdxct-frontend")
	XDXCTGPU          = XDXCTFrontend
	procDevicesPath   = "/proc/devices"
	xdxctDevicePrefix = "xdxct"
)

// Name represents the name of a device as specified under /proc/devices
type Name string

// Major represents a device major as specified under /proc/devices
type Major int

// Devices represents the set of devices under /proc/devices
//
//go:generate moq -stub -out devices_mock.go . Devices
type Devices interface {
	Exists(Name) bool
	Get(Name) (Major, bool)
}

type devices map[Name]Major

var _ Devices = devices(nil)

// Exists checks if a Device with a given name exists or not
func (d devices) Exists(name Name) bool {
	_, exists := d[name]
	return exists
}

// Get a Device from Devices
func (d devices) Get(name Name) (Major, bool) {
	device, exists := d[name]
	return device, exists
}

// GetXDXCTDevices returns the set of XDXCT Devices on the machine
func GetXDXCTDevices() (Devices, error) {
	return xdxctDevices(procDevicesPath)
}

// xdxctDevices returns the set of XDXCT Devices from the specified devices file.
// This is useful for testing since we may be testing on a system where `/proc/devices` does
// contain a reference to XDXCT devices.
func xdxctDevices(devicesPath string) (Devices, error) {
	devicesFile, err := os.Open(devicesPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error opening devices file: %v", err)
	}
	defer devicesFile.Close()

	return xdxctDeviceFrom(devicesFile)
}

var errNoXdxctDevices = errors.New("no XDXCT devices found")

func xdxctDeviceFrom(reader io.Reader) (devices, error) {
	allDevices := devicesFrom(reader)
	xdxctDevices := make(devices)

	var hasXdxctDevices bool
	for n, d := range allDevices {
		if !strings.HasPrefix(string(n), xdxctDevicePrefix) {
			continue
		}
		xdxctDevices[n] = d
		hasXdxctDevices = true
	}

	if !hasXdxctDevices {
		return nil, errNoXdxctDevices
	}
	return xdxctDevices, nil
}

func devicesFrom(reader io.Reader) devices {
	allDevices := make(devices)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		device, major, err := processProcDeviceLine(scanner.Text())
		if err != nil {
			continue
		}
		allDevices[device] = major
	}
	return allDevices
}

func processProcDeviceLine(line string) (Name, Major, error) {
	trimmed := strings.TrimSpace(line)

	var name Name
	var major Major

	n, _ := fmt.Sscanf(trimmed, "%d %s", &major, &name)
	if n == 2 {
		return name, major, nil
	}

	return "", 0, fmt.Errorf("unparsable line: %v", line)
}
