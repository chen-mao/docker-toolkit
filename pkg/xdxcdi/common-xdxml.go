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
