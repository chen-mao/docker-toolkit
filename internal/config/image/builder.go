package image

import (
	"fmt"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
)

type builder struct {
	env    map[string]string
	mounts []specs.Mount
	disableRequire bool
}

// New creates a new GPU image from the input options.
func New(opt ...Option) (GPU, error) {
	b := &builder{}
	for _, o := range opt {
		if err := o(b); err != nil {
			return GPU{}, err
		}
	}
	if b.env == nil {
		b.env = make(map[string]string)
	}

	return b.build()
}

// build creates a GPU image from the builder.
func (b builder) build() (GPU, error) {
	if b.disableRequire {
		b.env[envXDXDisableRequire] = "true"
	}

	c := GPU{
		env:    b.env,
		mounts: b.mounts,
	}
	return c, nil
}

// Option is a functional option for creating a GPU image.
type Option func(*builder) error

// WithDisableRequire sets the disable require option.
func WithDisableRequire(disableRequire bool) Option {
	return func(b *builder) error {
		b.disableRequire = disableRequire
		return nil
	}
}

// WithEnv sets the environment variables to use when creating the GPU image.
// Note that this also overwrites the values set with WithEnvMap.
func WithEnv(env []string) Option {
	return func(b *builder) error {
		envmap := make(map[string]string)
		for _, e := range env {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid environment variable: %v", e)
			}
			envmap[parts[0]] = parts[1]
		}
		return WithEnvMap(envmap)(b)
	}
}

// WithEnvMap sets the environment variable map to use when creating the GPU image.
// Note that this also overwrites the values set with WithEnv.
func WithEnvMap(env map[string]string) Option {
	return func(b *builder) error {
		b.env = env
		return nil
	}
}

// WithMounts sets the mounts associated with the GPU image.
func WithMounts(mounts []specs.Mount) Option {
	return func(b *builder) error {
		b.mounts = mounts
		return nil
	}
}
