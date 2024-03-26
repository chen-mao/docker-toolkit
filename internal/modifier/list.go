package modifier

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type list struct {
	modifiers []oci.SpecModifier
}

// Merge merges a set of OCI specification modifiers as a list.
// This can be used to compose modifiers.
func Merge(modifiers ...oci.SpecModifier) oci.SpecModifier {
	var filteredModifiers []oci.SpecModifier
	for _, m := range modifiers {
		if m == nil {
			continue
		}
		filteredModifiers = append(filteredModifiers, m)
	}

	return list{
		modifiers: filteredModifiers,
	}
}

// Modify applies a list of modifiers in sequence and returns on any errors encountered.
func (m list) Modify(spec *specs.Spec) error {
	for _, mm := range m.modifiers {
		err := mm.Modify(spec)
		if err != nil {
			return err
		}
	}

	return nil
}
