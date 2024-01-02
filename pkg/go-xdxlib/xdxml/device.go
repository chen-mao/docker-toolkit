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

import (
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxml/xdxml"
)

type xdxmlDevice xdxml.Device

var _ Device = (*xdxmlDevice)(nil)

func (d xdxmlDevice) GetUUID() (string, Return) {
	u, r := xdxml.Device(d).GetUUID()
	return u, Return(r)
}

func (d xdxmlDevice) GetMinorNumber() (int, Return) {
	m, r := xdxml.Device(d).GetMinorNumber()
	return m, Return(r)
}

func (d xdxmlDevice) GetArchitecture() (string, Return) {
	name := make([]byte, 64)
	r := xdxml.Device(d).GetProductName(name)
	return string(name), Return(r)
}

func (d xdxmlDevice) GetPciInfo() (PciInfo, Return) {
	p, r := xdxml.Device(d).GetPciInfo()
	return PciInfo(p), Return(r)
}
