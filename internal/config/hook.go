package config

// RuntimeHookConfig stores the config options for the XDXCT Container Runtime
type RuntimeHookConfig struct {
	// Path specifies the path to the XDXCT Container Runtime hook binary.
	// If an executable name is specified, this will be resolved in the path.
	Path string `toml:"path"`
	// SkipModeDetection disables the mode check for the runtime hook.
	SkipModeDetection bool `toml:"skip-mode-detection"`
}

// GetDefaultRuntimeHookConfig defines the default values for the config
func GetDefaultRuntimeHookConfig() (*RuntimeHookConfig, error) {
	cfg, err := GetDefault()
	if err != nil {
		return nil, err
	}

	return &cfg.XDXCTContainerRuntimeHookConfig, nil
}
