package modifier

import (
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// xdxctContainerRuntimeHookRemover is a spec modifer that detects and removes inserted xdxct-container-runtime hooks
type xdxctContainerRuntimeHookRemover struct {
	logger logger.Interface
}

var _ oci.SpecModifier = (*xdxctContainerRuntimeHookRemover)(nil)

// Modify removes any XDXCT Container Runtime hooks from the provided spec
func (m xdxctContainerRuntimeHookRemover) Modify(spec *specs.Spec) error {
	if spec == nil {
		return nil
	}

	if spec.Hooks == nil {
		return nil
	}

	if len(spec.Hooks.Prestart) == 0 {
		return nil
	}

	var newPrestart []specs.Hook

	for _, hook := range spec.Hooks.Prestart {
		if isXDXCTContainerRuntimeHook(&hook) {
			m.logger.Debugf("Removing hook %v", hook)
			continue
		}
		newPrestart = append(newPrestart, hook)
	}

	if len(newPrestart) != len(spec.Hooks.Prestart) {
		m.logger.Debugf("Updating 'prestart' hooks to %v", newPrestart)
		spec.Hooks.Prestart = newPrestart
	}

	return nil
}

// isXDXCTContainerRuntimeHook checks if the provided hook is an xdxct-container-runtime-hook
// or xdxct-container-toolkit hook. These are included, for example, by the non-experimental
// xdxct-container-runtime or docker when specifying the --gpus flag.
func isXDXCTContainerRuntimeHook(hook *specs.Hook) bool {
	bins := map[string]struct{}{
		config.XDXCTContainerRuntimeHookExecutable: {},
		config.XDXCTContainerToolkitExecutable:     {},
	}

	_, exists := bins[filepath.Base(hook.Path)]

	return exists
}
