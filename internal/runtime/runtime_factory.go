/*
# Copyright (c) 2021-2022, NVIDIA CORPORATION.  All rights reserved.
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
*/

package runtime

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config"
	"github.com/XDXCT/xdxct-container-toolkit/internal/info"
	"github.com/XDXCT/xdxct-container-toolkit/internal/modifier"
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
	"github.com/sirupsen/logrus"
)

// newNVIDIAContainerRuntime is a factory method that constructs a runtime based on the selected configuration and specified logger
func newNVIDIAContainerRuntime(logger *logrus.Logger, cfg *config.Config, argv []string) (oci.Runtime, error) {
	lowLevelRuntime, err := oci.NewLowLevelRuntime(logger, cfg.XDXCTContainerRuntimeConfig.Runtimes)
	if err != nil {
		return nil, fmt.Errorf("error constructing low-level runtime: %v", err)
	}

	if !oci.HasCreateSubcommand(argv) {
		logger.Debugf("Skipping modifier for non-create subcommand")
		return lowLevelRuntime, nil
	}

	ociSpec, err := oci.NewSpec(logger, argv)
	if err != nil {
		return nil, fmt.Errorf("error constructing OCI specification: %v", err)
	}

	specModifier, err := newSpecModifier(logger, cfg, ociSpec, argv)
	if err != nil {
		return nil, fmt.Errorf("failed to construct OCI spec modifier: %v", err)
	}

	// Create the wrapping runtime with the specified modifier
	r := oci.NewModifyingRuntimeWrapper(
		logger,
		lowLevelRuntime,
		ociSpec,
		specModifier,
	)

	return r, nil
}

// newSpecModifier is a factory method that creates constructs an OCI spec modifer based on the provided config.
func newSpecModifier(logger *logrus.Logger, cfg *config.Config, ociSpec oci.Spec, argv []string) (oci.SpecModifier, error) {
	mode := info.ResolveAutoMode(logger, cfg.XDXCTContainerRuntimeConfig.Mode)
	modeModifier, err := newModeModifier(logger, mode, cfg, ociSpec, argv)
	if err != nil {
		return nil, err
	}
	// For CDI mode we make no additional modifications.
	if mode == "cdi" {
		return modeModifier, nil
	}

	graphicsModifier, err := modifier.NewGraphicsModifier(logger, cfg, ociSpec)
	if err != nil {
		return nil, err
	}

	gdsModifier, err := modifier.NewGDSModifier(logger, cfg, ociSpec)
	if err != nil {
		return nil, err
	}

	mofedModifier, err := modifier.NewMOFEDModifier(logger, cfg, ociSpec)
	if err != nil {
		return nil, err
	}

	modifiers := modifier.Merge(
		modeModifier,
		graphicsModifier,
		gdsModifier,
		mofedModifier,
	)
	return modifiers, nil
}

func newModeModifier(logger *logrus.Logger, mode string, cfg *config.Config, ociSpec oci.Spec, argv []string) (oci.SpecModifier, error) {
	switch mode {
	case "legacy":
		return modifier.NewStableRuntimeModifier(logger, cfg.XDXCTContainerRuntimeHookConfig.Path), nil
	case "csv":
		return modifier.NewCSVModifier(logger, cfg, ociSpec)
	case "cdi":
		return modifier.NewCDIModifier(logger, cfg, ociSpec)
	}

	return nil, fmt.Errorf("invalid runtime mode: %v", cfg.XDXCTContainerRuntimeConfig.Mode)
}
