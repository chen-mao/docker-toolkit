package noop

import (
	"tags.cncf.io/container-device-interface/specs-go"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform"
)

type noop struct{}

var _ transform.Transformer = (*noop)(nil)

// New returns a no-op transformer.
func New() transform.Transformer {
	return noop{}
}

// Transform is a no-op for a noop transformer.
func (n noop) Transform(spec *specs.Spec) error {
	return nil
}
