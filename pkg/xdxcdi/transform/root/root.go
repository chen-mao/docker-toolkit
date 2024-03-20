package root

import (
	"path/filepath"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform"
)

// transformer transforms roots of paths.
type transformer struct {
	root       string
	targetRoot string
}

// New creates a root transformer using the specified options.
func New(opts ...Option) transform.Transformer {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}
	return b.build()
}

func (t transformer) transformPath(path string) string {
	if !strings.HasPrefix(path, t.root) {
		return path
	}

	return filepath.Join(t.targetRoot, strings.TrimPrefix(path, t.root))
}
