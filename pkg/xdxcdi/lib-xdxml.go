/**
# Copyright (c) NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package xdxcdi

import (
	"fmt"
	"strconv"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"

	"github.com/XDXCT/xdxct-container-toolkit/internal/edits"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/spec"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxml"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/device"
)

type xdxmllib xdxcdilib

var _ Interface = (*xdxmllib)(nil)

// GetSpec should not be called for nvmllib
func (l *xdxmllib) GetSpec() (spec.Interface, error) {
	return nil, fmt.Errorf("Unexpected call to xdxmllib.GetSpec()")
}

// GetAllDeviceSpecs returns the device specs for all available devices.
func (l *xdxmllib) GetAllDeviceSpecs() ([]specs.Device, error) {
	var deviceSpecs []specs.Device

	// TODO
	if r := l.xdxmllib.Init(); r != xdxml.SUCCESS {
		return nil, fmt.Errorf("failed to initialize XDXML: %v", r)
	}
	defer func() {
		if r := l.xdxmllib.Shutdown(); r != xdxml.SUCCESS {
			l.logger.Warningf("failed to shutdown XDXML: %v", r)
		}
	}()

	gpuDeviceSpecs, err := l.getGPUDeviceSpecs()
	if err != nil {
		return nil, err
	}
	deviceSpecs = append(deviceSpecs, gpuDeviceSpecs...)

	return deviceSpecs, nil
}

// GetCommonEdits generates a CDI specification that can be used for ANY devices
func (l *xdxmllib) GetCommonEdits() (*cdi.ContainerEdits, error) {
	common, err := l.newCommonXDXMLDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("failed to create discoverer for common entities: %v", err)
	}

	return edits.FromDiscoverer(common)
}

// GetDeviceSpecsByID returns the CDI device specs for the GPU(s) represented by
// the provided identifiers, where an identifier is an index or UUID of a valid
// GPU device.
func (l *xdxmllib) GetDeviceSpecsByID(identifiers ...string) ([]specs.Device, error) {
	for _, id := range identifiers {
		if id == "all" {
			return l.GetAllDeviceSpecs()
		}
	}

	var deviceSpecs []specs.Device

	if r := l.xdxmllib.Init(); r != xdxml.SUCCESS {
		return nil, fmt.Errorf("failed to initialize XDXML: %w", r)
	}
	defer func() {
		if r := l.xdxmllib.Shutdown(); r != xdxml.SUCCESS {
			l.logger.Warningf("failed to shutdown XDXML: %w", r)
		}
	}()

	xdxmlDevices, err := l.getXDXMLDevicesByID(identifiers...)
	if err != nil {
		return nil, fmt.Errorf("failed to get NVML device handles: %w", err)
	}

	for i, xdxmlDevice := range xdxmlDevices {
		deviceEdits, err := l.getEditsForDevice(xdxmlDevice)
		if err != nil {
			return nil, fmt.Errorf("failed to get CDI device edits for identifier %q: %w", identifiers[i], err)
		}
		deviceSpec := specs.Device{
			Name:           identifiers[i],
			ContainerEdits: *deviceEdits.ContainerEdits,
		}
		deviceSpecs = append(deviceSpecs, deviceSpec)
	}

	return deviceSpecs, nil
}

func (l *xdxmllib) getXDXMLDevicesByID(identifiers ...string) ([]xdxml.Device, error) {
	var devices []xdxml.Device
	for _, id := range identifiers {
		dev, err := l.getXDXMLDeviceByID(id)
		if err != xdxml.SUCCESS {
			return nil, fmt.Errorf("failed to get NVML device handle for identifier %q: %w", id, err)
		}
		devices = append(devices, dev)
	}
	return devices, nil
}

func (l *xdxmllib) getXDXMLDeviceByID(id string) (xdxml.Device, error) {
	var err error
	// TODO: How to get id
	devID := device.Identifier(id)

	if devID.IsUUID() {
		return nil, fmt.Errorf("not support uuid mode: %w", err)
	}

	if devID.IsGpuIndex() {
		if idx, err := strconv.Atoi(id); err == nil {
			return l.xdxmllib.DeviceGetHandleByIndex(idx)
		}
		return nil, fmt.Errorf("failed to convert device index to an int: %w", err)
	}

	return nil, fmt.Errorf("identifier is not a valid UUID or index: %q", id)
}

func (l *xdxmllib) getEditsForDevice(xdxmlDevice xdxml.Device) (*cdi.ContainerEdits, error) {
	return l.getEditsForGPUDevice(xdxmlDevice)
}

func (l *xdxmllib) getEditsForGPUDevice(xdxmlDevice xdxml.Device) (*cdi.ContainerEdits, error) {
	// nvlibDevice, err := l.devicelib.NewDevice(xdxmlDevice)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to construct device: %w", err)
	// }
	nvlibDevice := xdxmlDevice
	deviceEdits, err := l.GetGPUDeviceEdits(nvlibDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU device edits: %w", err)
	}

	return deviceEdits, nil
}

func (l *xdxmllib) getGPUDeviceSpecs() ([]specs.Device, error) {
	var deviceSpecs []specs.Device
	err := l.devicelib.VisitDevices(func(i int, d device.Device) error {
		deviceSpec, err := l.GetGPUDeviceSpecs(i, d)
		if err != nil {
			return err
		}
		deviceSpecs = append(deviceSpecs, *deviceSpec)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate CDI edits for GPU devices: %v", err)
	}
	return deviceSpecs, err
}
