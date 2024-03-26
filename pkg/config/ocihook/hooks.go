package ocihook

// podmanHook is the hook configuration structure.
// This is taken from `Hook` at https://github.com/containers/podman/blob/3c53200e9d61fdf95fe1da825bb2a89372551350/pkg/hooks/1.0.0/hook.go#L18
type podmanHook struct {
	Version string   `json:"version"`
	Hook    specHook `json:"hook"`
	When    When     `json:"when"`
	Stages  []string `json:"stages"`
}

// specHook specifies a command that is run at a particular event in the lifecycle of a container
// This is taken from `Hook` at https://github.com/opencontainers/runtime-spec/blob/9ee22abf867e374c5464c7bbe0d0db01482254ab/specs-go/config.go#L128
type specHook struct {
	Path    string   `json:"path"`
	Args    []string `json:"args,omitempty"`
	Env     []string `json:"env,omitempty"`
	Timeout *int     `json:"timeout,omitempty"`
}

// When holds hook-injection conditions.
// This is taken from `When` at https://github.com/containers/podman/blob/3c53200e9d61fdf95fe1da825bb2a89372551350/pkg/hooks/1.0.0/when.go#L11
type When struct {
	Always        *bool             `json:"always,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Commands      []string          `json:"commands,omitempty"`
	HasBindMounts *bool             `json:"hasBindMounts,omitempty"`

	// Or enables any-of matching.
	//
	// Deprecated: this property is for is backwards-compatibility with
	// 0.1.0 hooks.  It will be removed when we drop support for them.
	Or bool `json:"-"`
}
