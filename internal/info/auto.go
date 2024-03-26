package info

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/config/image"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

type resolver struct {
	logger logger.Interface
	// info   infoInterface
}

// ResolveAutoMode determines the correct mode for the platform if set to "auto"
func ResolveAutoMode(logger logger.Interface, mode string, image image.GPU) (rmode string) {
	r := resolver{
		logger: logger,
		// info:   info,
	}
	return r.resolveMode(mode, image)
}

// ResolveAutoMode determines the correct mode for the platform if set to "auto"
func (r resolver) resolveMode(mode string, image image.GPU) (rmode string) {
	if mode != "auto" {
		return mode
	}
	defer func() {
		r.logger.Infof("Auto-detected mode as '%v'", rmode)
	}()

	if image.OnlyFullyQualifiedCDIDevices() {
		return "cdi"
	}

	return "legacy"
}
