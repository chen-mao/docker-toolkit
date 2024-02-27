/*
 * Copyright (c) 2024, XDXCT CORPORATION.  All rights reserved.
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

package xdxml

// xdxml.DeviceGetMinorNumber()
func DeviceGetMinorNumber(Device Device) (int, Return) {
	var minorNumber int32
	ret := xdxml_device_get_minor_number(Device, &minorNumber)
	return int(minorNumber), ret
}

func (Device Device) GetMinorNumber() (int, Return) {
	return DeviceGetMinorNumber(Device)
}

// xdxml.DeviceGetCount()
func DeviceGetCount() (int, Return) {
	var DeviceCount uint32
	ret := xdxml_device_get_count(&DeviceCount)
	return int(DeviceCount), ret
}

// xdxml.DeviceGetHandleByIndex()
func DeviceGetHandleByIndex(Index int) (Device, Return) {
	var Device Device
	ret := xdxml_device_get_handle_by_index(uint32(Index), &Device)
	return Device, ret
}

// xdxml.DeviceGetHandleByIndex()
func DeviceGetProductName(device Device, name []byte) Return {
	ret := xdxml_device_get_product_name(device, &name[0])
	return ret
}

func (Device Device) GetProductName(name []byte) Return {
	return DeviceGetProductName(Device, name)
}

// xdxml.DeviceGetUUID()
func DeviceGetUUID(Device Device) (string, Return) {
	ret := xdxml_device_get_uuid(Device)
	var uuidStr string
	for _, num := range Device.Handle.uuid {
		uuidStr += string(num)
	}
	return uuidStr, ret
}

func (Device Device) GetUUID() (string, Return) {
	return DeviceGetUUID(Device)
}

func (Device Device) GetPciInfo() (Pci_info, Return) {
	var pci Pci_info
	ret := xdxml_device_get_pci_info(Device, &pci)
	return pci, ret
}
