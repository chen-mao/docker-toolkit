package cdi

import (
	"fmt"

	"github.com/opencontainers/runtime-spec/specs-go"
	"tags.cncf.io/container-device-interface/pkg/cdi"

	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
)

// fromCDISpec represents the modifications performed from a raw CDI spec.
type fromCDISpec struct {
	cdiSpec *cdi.Spec
}

var _ oci.SpecModifier = (*fromCDISpec)(nil)

// Modify applies the mofiications defined by the raw CDI spec to the incomming OCI spec.
func (m fromCDISpec) Modify(spec *specs.Spec) error {
	for _, device := range m.cdiSpec.Devices {
		device := device
		cdiDevice := cdi.Device{
			Device: &device,
		}
		if err := cdiDevice.ApplyEdits(spec); err != nil {
			return fmt.Errorf("failed to apply edits for device %q: %v", cdiDevice.GetQualifiedName(), err)
		}
	}

	return m.cdiSpec.ApplyEdits(spec)
}
