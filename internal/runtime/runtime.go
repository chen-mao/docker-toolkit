package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config"
	"github.com/XDXCT/xdxct-container-toolkit/internal/info"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// Run is an entry point that allows for idiomatic handling of errors
// when calling from the main function.
func (r rt) Run(argv []string) (rerr error) {
	defer func() {
		if rerr != nil {
			r.logger.Errorf("%v", rerr)
		}
	}()

	printVersion := hasVersionFlag(argv)
	if printVersion {
		fmt.Printf("%v version %v\n", "XDXCT Container Runtime", info.GetVersionString(fmt.Sprintf("spec: %v", specs.Version)))
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	r.logger.Update(
		cfg.XDXCTContainerRuntimeConfig.DebugFilePath,
		cfg.XDXCTContainerRuntimeConfig.LogLevel,
		argv,
	)
	defer func() {
		if rerr != nil {
			r.logger.Errorf("%v", rerr)
		}
		r.logger.Reset()
	}()

	// We apply some config updates here to ensure that the config is valid in
	// all cases.
	if r.modeOverride != "" {
		cfg.XDXCTContainerRuntimeConfig.Mode = r.modeOverride
	}
	cfg.XDXCTCTKConfig.Path = config.ResolveXDXCTCTKPath(r.logger, cfg.XDXCTCTKConfig.Path)
	cfg.XDXCTContainerRuntimeHookConfig.Path = config.ResolveXDXCTContainerRuntimeHookPath(r.logger, cfg.XDXCTContainerRuntimeHookConfig.Path)

	// Print the config to the output.
	configJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		r.logger.Infof("Running with config:\n%v", string(configJSON))
	} else {
		r.logger.Infof("Running with config:\n%+v", cfg)
	}

	r.logger.Debugf("Command line arguments: %v", argv)
	runtime, err := newNVIDIAContainerRuntime(r.logger, cfg, argv)
	if err != nil {
		return fmt.Errorf("failed to create NVIDIA Container Runtime: %v", err)
	}

	if printVersion {
		fmt.Print("\n")
	}
	return runtime.Exec(argv)
}

func (r rt) Errorf(format string, args ...interface{}) {
	r.logger.Errorf(format, args...)
}

// TODO: This should be refactored / combined with parseArgs in logger.
func hasVersionFlag(args []string) bool {
	for i := 0; i < len(args); i++ {
		param := args[i]

		parts := strings.SplitN(param, "=", 2)
		trimmed := strings.TrimLeft(parts[0], "-")
		// If this is not a flag we continue
		if parts[0] == trimmed {
			continue
		}

		// Check the version flag
		if trimmed == "version" {
			return true
		}
	}

	return false
}
