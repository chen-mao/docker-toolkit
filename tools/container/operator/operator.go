package operator

import "path/filepath"

const (
	defaultRuntimeName = "xdxct"
	// experimentalRuntimeName = "xdxct-experimental"

	defaultRoot = "/usr/bin"
)

// Runtime defines a runtime to be configured.
// The path and whether the runtime is the default runtime can be specfied
type Runtime struct {
	name         string
	Path         string
	SetAsDefault bool
}

// Runtimes defines a set of runtimes to be configure for use in the GPU Operator
type Runtimes map[string]Runtime

type config struct {
	root             string
	xdxctRuntimeName string
	setAsDefault     bool
}

// GetRuntimes returns the set of runtimes to be configured for use with the GPU Operator.
func GetRuntimes(opts ...Option) Runtimes {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}

	if c.root == "" {
		c.root = defaultRoot
	}
	if c.xdxctRuntimeName == "" {
		c.xdxctRuntimeName = defaultRuntimeName
	}

	runtimes := make(Runtimes)
	runtimes.add(c.xdxctRuntime())

	// modes := []string{"experimental", "cdi", "legacy"}
	// for _, mode := range modes {
	// 	runtimes.add(c.modeRuntime(mode))
	// }
	return runtimes
}

// DefaultRuntimeName returns the name of the default runtime.
func (r Runtimes) DefaultRuntimeName() string {
	for _, runtime := range r {
		if runtime.SetAsDefault {
			return runtime.name
		}
	}
	return ""
}

// Add a runtime to the set of runtimes.
func (r *Runtimes) add(runtime Runtime) {
	(*r)[runtime.name] = runtime
}

// xdxctRuntime creates a runtime that corresponds to the xdxct runtime.
// If name is equal to one of the predefined runtimes, `xdxct` is used as the runtime name instead.
func (c config) xdxctRuntime() Runtime {
	// predefinedRuntimes := map[string]struct{}{
	// 	"xdxct-experimental": {},
	// 	"xdxct-cdi":          {},
	// 	"xdxct-legacy":       {},
	// }
	name := c.xdxctRuntimeName
	// if _, isPredefinedRuntime := predefinedRuntimes[name]; isPredefinedRuntime {
	// 	name = defaultRuntimeName
	// }
	return c.newRuntime(name, "xdxct-container-runtime")
}

// modeRuntime creates a runtime for the specified mode.
func (c config) modeRuntime(mode string) Runtime {
	return c.newRuntime("xdxct-"+mode, "xdxct-container-runtime."+mode)
}

// newRuntime creates a runtime based on the configuration
func (c config) newRuntime(name string, binary string) Runtime {
	return Runtime{
		name:         name,
		Path:         filepath.Join(c.root, binary),
		SetAsDefault: c.setAsDefault && name == c.xdxctRuntimeName,
	}
}

// Option is a functional option for configuring set of runtimes.
type Option func(*config)

// WithRoot sets the root directory for the runtime binaries.
func WithRoot(root string) Option {
	return func(c *config) {
		c.root = root
	}
}

// WithXdxctRuntimeName sets the name of the xdxct runtime.
func WithXdxctRuntimeName(name string) Option {
	return func(c *config) {
		c.xdxctRuntimeName = name
	}
}

// WithSetAsDefault sets the default runtime to the xdxct runtime.
func WithSetAsDefault(set bool) Option {
	return func(c *config) {
		c.setAsDefault = set
	}
}
