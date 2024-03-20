package containerd

import (
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine"
	"github.com/pelletier/go-toml"
)

// Config represents the containerd config
type Config struct {
	*toml.Tree
	RuntimeType           string
	UseDefaultRuntimeName bool
	ContainerAnnotations  []string
}

// New creates a containerd config with the specified options
func New(opts ...Option) (engine.Interface, error) {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}

	return b.build()
}
