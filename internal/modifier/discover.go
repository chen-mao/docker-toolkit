package modifier

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
	"github.com/XDXCT/xdxct-container-toolkit/internal/edits"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type discoverModifier struct {
	logger     logger.Interface
	discoverer discover.Discover
}

// NewModifierFromDiscoverer creates a modifier that applies the discovered
// modifications to an OCI spec if required by the runtime wrapper.
func NewModifierFromDiscoverer(logger logger.Interface, d discover.Discover) (oci.SpecModifier, error) {
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
