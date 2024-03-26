package modifier

import (
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// NewStableRuntimeModifier creates an OCI spec modifier that inserts the XDXCT Container Runtime Hook into an OCI
// spec. The specified logger is used to capture log output.
func NewStableRuntimeModifier(logger logger.Interface, xdxctContainerRuntimeHookPath string) oci.SpecModifier {
	m := stableRuntimeModifier{
		logger:                         logger,
		xdxctContainerRuntimeHookPath: xdxctContainerRuntimeHookPath,
	}

	return &m
}

// stableRuntimeModifier modifies an OCI spec inplace, inserting the xdxct-container-runtime-hook as a
// prestart hook. If the hook is already present, no modification is made.
type stableRuntimeModifier struct {
	logger                         logger.Interface
	xdxctContainerRuntimeHookPath string
}

// Modify applies the required modification to the incoming OCI spec, inserting the xdxct-container-runtime-hook
// as a prestart hook.
func (m stableRuntimeModifier) Modify(spec *specs.Spec) error {
	// If an XDXCT Container Runtime Hook already exists, we don't make any modifications to the spec.
	if spec.Hooks != nil {
		for _, hook := range spec.Hooks.Prestart {
			if isXDXCTContainerRuntimeHook(&hook) {
				m.logger.Infof("Existing xdxct prestart hook (%v) found in OCI spec", hook.Path)
				return nil
			}
		}
	}

	path := m.xdxctContainerRuntimeHookPath
	m.logger.Infof("Using prestart hook path: %v", path)
	args := []string{filepath.Base(path)}
	if spec.Hooks == nil {
		spec.Hooks = &specs.Hooks{}
	}
	spec.Hooks.Prestart = append(spec.Hooks.Prestart, specs.Hook{
		Path: path,
		Args: append(args, "prestart"),
	})

	return nil
}
