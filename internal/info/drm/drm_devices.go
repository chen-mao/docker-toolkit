/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
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
