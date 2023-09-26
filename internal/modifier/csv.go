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

	"github.com/XDXCT/xdxct-container-toolkit/internal/config"
	"github.com/XDXCT/xdxct-container-toolkit/internal/config/image"
	"github.com/XDXCT/xdxct-container-toolkit/internal/cuda"
	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
	"github.com/XDXCT/xdxct-container-toolkit/internal/discover/csv"
	"github.com/XDXCT/xdxct-container-toolkit/internal/discover/tegra"
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
	"github.com/XDXCT/xdxct-container-toolkit/internal/requirements"
	"github.com/sirupsen/logrus"
)

// csvMode represents the modifications as performed by the csv runtime mode
type csvMode struct {
	logger     *logrus.Logger
	discoverer discover.Discover
}

const (
	visibleDevicesEnvvar = "XDXCT_VISIBLE_DEVICES"
	visibleDevicesVoid   = "void"

	nvidiaRequireJetpackEnvvar = "NVIDIA_REQUIRE_JETPACK"
)

// NewCSVModifier creates a modifier that applies modications to an OCI spec if required by the runtime wrapper.
// The modifications are defined by CSV MountSpecs.
func NewCSVModifier(logger *logrus.Logger, cfg *config.Config, ociSpec oci.Spec) (oci.SpecModifier, error) {
	rawSpec, err := ociSpec.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load OCI spec: %v", err)
	}

	image, err := image.NewCUDAImageFromSpec(rawSpec)
	if err != nil {
		return nil, err
	}

	if devices := image.DevicesFromEnvvars(visibleDevicesEnvvar); len(devices.List()) == 0 {
		logger.Infof("No modification required; no devices requested")
		return nil, nil
	}
	logger.Infof("Constructing modifier from config: %+v", *cfg)

	if err := checkRequirements(logger, image); err != nil {
		return nil, fmt.Errorf("requirements not met: %v", err)
	}

	csvFiles, err := csv.GetFileList(cfg.XDXCTContainerRuntimeConfig.Modes.CSV.MountSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of CSV files: %v", err)
	}

	if nvidiaRequireJetpack, _ := image[nvidiaRequireJetpackEnvvar]; nvidiaRequireJetpack != "csv-mounts=all" {
		csvFiles = csv.BaseFilesOnly(csvFiles)
	}

	d, err := tegra.New(
		tegra.WithLogger(logger),
		tegra.WithDriverRoot(cfg.XDXCTContainerCLIConfig.Root),
		tegra.WithNVIDIACTKPath(cfg.XDXCTCTKConfig.Path),
		tegra.WithCSVFiles(csvFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct discoverer: %v", err)
	}

	discoverModifier, err := NewModifierFromDiscoverer(logger, d)
	if err != nil {
		return nil, fmt.Errorf("failed to construct modifier: %v", err)
	}

	modifiers := Merge(
		nvidiaContainerRuntimeHookRemover{logger},
		discoverModifier,
	)

	return modifiers, nil
}

func checkRequirements(logger *logrus.Logger, image image.CUDA) error {
	if image.HasDisableRequire() {
		// TODO: We could print the real value here instead
		logger.Debugf("NVIDIA_DISABLE_REQUIRE=%v; skipping requirement checks", true)
		return nil
	}

	imageRequirements, err := image.GetRequirements()
	if err != nil {
		//  TODO: Should we treat this as a failure, or just issue a warning?
		return fmt.Errorf("failed to get image requirements: %v", err)
	}

	r := requirements.New(logger, imageRequirements)

	cudaVersion, err := cuda.Version()
	if err != nil {
		logger.Warnf("Failed to get CUDA version: %v", err)
	} else {
		r.AddVersionProperty(requirements.CUDA, cudaVersion)
	}

	compteCapability, err := cuda.ComputeCapability(0)
	if err != nil {
		logger.Warnf("Failed to get CUDA Compute Capability: %v", err)
	} else {
		r.AddVersionProperty(requirements.ARCH, compteCapability)
	}

	return r.Assert()
}
