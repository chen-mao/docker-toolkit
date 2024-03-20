package info

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/config/image"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

//go:generate moq -stub -out info-interface_mock.go . infoInterface
// type infoInterface interface {
// 	info.Interface
// 	// UsesNVGPUModule indicates whether the system is using the nvgpu kernel module
// 	UsesNVGPUModule() (bool, string)
// }

type resolver struct {
	logger logger.Interface
	// info   infoInterface
}

// ResolveAutoMode determines the correct mode for the platform if set to "auto"
func ResolveAutoMode(logger logger.Interface, mode string, image image.CUDA) (rmode string) {
	// nvinfo := info.New()
	// nvmllib := nvml.New()
	// devicelib := device.New(
	// 	device.WithNvml(nvmllib),
	// )

	// info := additionalInfo{
	// 	Interface: nvinfo,
	// 	nvmllib:   nvmllib,
	// 	devicelib: devicelib,
	// }

	r := resolver{
		logger: logger,
		// info:   info,
	}
	return r.resolveMode(mode, image)
}

// ResolveAutoMode determines the correct mode for the platform if set to "auto"
func (r resolver) resolveMode(mode string, image image.CUDA) (rmode string) {
	if mode != "auto" {
		return mode
	}
	defer func() {
		r.logger.Infof("Auto-detected mode as '%v'", rmode)
	}()

	if image.OnlyFullyQualifiedCDIDevices() {
		return "cdi"
	}

	// isTegra, reason := r.info.IsTegraSystem()
	// r.logger.Debugf("Is Tegra-based system? %v: %v", isTegra, reason)

	// hasNVML, reason := r.info.HasNvml()
	// r.logger.Debugf("Has NVML? %v: %v", hasNVML, reason)

	// usesNVGPUModule, reason := r.info.UsesNVGPUModule()
	// r.logger.Debugf("Uses nvgpu kernel module? %v: %v", usesNVGPUModule, reason)

	// if (isTegra && !hasNVML) || usesNVGPUModule {
	// 	return "csv"
	// }

	return "legacy"
}
