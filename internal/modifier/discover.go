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

package modifier

import (
	"fmt"

	"github.com/NVIDIA/xdxct-container-toolkit/internal/discover"
	"github.com/NVIDIA/xdxct-container-toolkit/internal/edits"
	"github.com/NVIDIA/xdxct-container-toolkit/internal/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

type discoverModifier struct {
	logger     *logrus.Logger
	discoverer discover.Discover
}

// NewModifierFromDiscoverer creates a modifier that applies the discovered
// modifications to an OCI spec if required by the runtime wrapper.
func NewModifierFromDiscoverer(logger *logrus.Logger, d discover.Discover) (oci.SpecModifier, error) {
	m := discoverModifier{
		logger:     logger,
		discoverer: d,
	}
	return &m, nil
}

// Modify applies the modifications required by discoverer to the incomming OCI spec.
// These modifications are applied in-place.
func (m discoverModifier) Modify(spec *specs.Spec) error {
	specEdits, err := edits.NewSpecEdits(m.logger, m.discoverer)
	if err != nil {
		return fmt.Errorf("failed to get required container edits: %v", err)
	}

	return specEdits.Modify(spec)
}
