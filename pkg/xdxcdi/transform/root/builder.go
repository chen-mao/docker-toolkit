package root

import (
	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform/noop"
)

type builder struct {
	transformer
	relativeTo string
}

func (b *builder) build() transform.Transformer {
	if b.root == b.targetRoot {
		return noop.New()
	}

	if b.relativeTo == "container" {
		return containerRootTransformer(b.transformer)
	}
	return hostRootTransformer(b.transformer)
}
