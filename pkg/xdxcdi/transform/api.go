package transform

import "tags.cncf.io/container-device-interface/specs-go"

// Transformer defines the API for applying arbitrary transforms to a spec in-place
type Transformer interface {
	Transform(*specs.Spec) error
}
