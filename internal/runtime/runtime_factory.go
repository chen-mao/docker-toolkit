package runtime

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config"
	"github.com/XDXCT/xdxct-container-toolkit/internal/config/image"
	"github.com/XDXCT/xdxct-container-toolkit/internal/info"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/modifier"
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
)

func newXDXCTContainerRuntime(logger logger.Interface, cfg *config.Config, argv []string) (oci.Runtime, error) {
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

	specModifier, err := newSpecModifier(logger, cfg, ociSpec)
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
func newSpecModifier(logger logger.Interface, cfg *config.Config, ociSpec oci.Spec) (oci.SpecModifier, error) {
	rawSpec, err := ociSpec.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load OCI spec: %v", err)
	}

	image, err := image.NewGPUImageFromSpec(rawSpec)
	if err != nil {
		return nil, err
	}

	mode := info.ResolveAutoMode(logger, cfg.XDXCTContainerRuntimeConfig.Mode, image)
	modeModifier, err := newModeModifier(logger, mode, cfg, ociSpec, image)
	if err != nil {
		return nil, err
	}
	// For CDI mode we make no additional modifications.
	if mode == "cdi" {
		return modeModifier, nil
	}

	graphicsModifier, err := modifier.NewGraphicsModifier(logger, cfg, image)
	if err != nil {
		return nil, err
	}

	modifiers := modifier.Merge(
		modeModifier,
		graphicsModifier,
	)
	return modifiers, nil
}

func newModeModifier(logger logger.Interface, mode string, cfg *config.Config, ociSpec oci.Spec, image image.GPU) (oci.SpecModifier, error) {
	switch mode {
	case "legacy":
		return modifier.NewStableRuntimeModifier(logger, cfg.XDXCTContainerRuntimeHookConfig.Path), nil
	// CSV mode is to supports tegra device.
	// case "csv":
	// 	return modifier.NewCSVModifier(logger, cfg, image)
	case "cdi":
		return modifier.NewCDIModifier(logger, cfg, ociSpec)
	}

	return nil, fmt.Errorf("invalid runtime mode: %v", cfg.XDXCTContainerRuntimeConfig.Mode)
}
