package config

import (
	"bufio"
	// "fmt"
	"os"
	"path/filepath"
	"strings"

	"tags.cncf.io/container-device-interface/pkg/cdi"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config/image"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
	// "github.com/pelletier/go-toml"
)

const (
	configOverride = "XDG_CONFIG_HOME"
	configFilePath = "xdxct-container-runtime/config.toml"

	xdxctCTKExecutable      = "xdxct-ctk"
	xdxctCTKDefaultFilePath = "/usr/bin/xdxct-ctk"

	xdxctContainerRuntimeHookExecutable  = "xdxct-container-runtime-hook"
	xdxctContainerRuntimeHookDefaultPath = "/usr/bin/xdxct-container-runtime-hook"
)

var (
	// DefaultExecutableDir specifies the default path to use for executables if they cannot be located in the path.
	DefaultExecutableDir = "/usr/bin"

	// XDXCTContainerRuntimeHookExecutable is the executable name for the XDXCT Container Runtime Hook
	XDXCTContainerRuntimeHookExecutable = "xdxct-container-runtime-hook"
	// XDXCTContainerToolkitExecutable is the executable name for the XDXCT Container Toolkit (an alias for the XDXCT Container Runtime Hook)
	XDXCTContainerToolkitExecutable = "xdxct-container-toolkit"

	configDir = "/etc/"
)

// Config represents the contents of the config.toml file for the XDXCT Container Toolkit
// Note: This is currently duplicated by the HookConfig in cmd/xdxct-container-toolkit/hook_config.go
type Config struct {
	// AcceptEnvvarUnprivileged bool `toml:"accept-xdxct-visible-devices-envvar-when-unprivileged"`
	DisableRequire                 bool   `toml:"disable-require"`
	SwarmResource                  string `toml:"swarm-resource"`
	AcceptEnvvarUnprivileged       bool   `toml:"accept-xdxct-visible-devices-envvar-when-unprivileged"`
	AcceptDeviceListAsVolumeMounts bool   `toml:"accept-xdxct-visible-devices-as-volume-mounts"`
	SupportedDriverCapabilities    string `toml:"supported-driver-capabilities"`

	XDXCTContainerCLIConfig         ContainerCLIConfig `toml:"xdxct-container-cli"`
	XDXCTCTKConfig                  CTKConfig          `toml:"xdxct-ctk"`
	XDXCTContainerRuntimeConfig     RuntimeConfig      `toml:"xdxct-container-runtime"`
	XDXCTContainerRuntimeHookConfig RuntimeHookConfig  `toml:"xdxct-container-runtime-hook"`
}

// GetConfigFilePath returns the path to the config file for the configured system
func GetConfigFilePath() string {
	if XDGConfigDir := os.Getenv(configOverride); len(XDGConfigDir) != 0 {
		return filepath.Join(XDGConfigDir, configFilePath)
	}

	return filepath.Join("/etc", configFilePath)
}

// GetConfig sets up the config struct. Values are read from a toml file
// or set via the environment.
func GetConfig() (*Config, error) {
	cfg, err := New(
		WithConfigFile(GetConfigFilePath()),
	)
	if err != nil {
		return nil, err
	}

	return cfg.Config()
}

// GetDefault defines the default values for the config
func GetDefault() (*Config, error) {
	d := Config{
		AcceptEnvvarUnprivileged:    true,
		SupportedDriverCapabilities: image.SupportedDriverCapabilities.String(),
		XDXCTContainerCLIConfig: ContainerCLIConfig{
			LoadKmods: true,
			Ldconfig:  getLdConfigPath(),
		},
		XDXCTCTKConfig: CTKConfig{
			Path: xdxctCTKExecutable,
		},
		XDXCTContainerRuntimeConfig: RuntimeConfig{
			DebugFilePath: "/dev/null",
			LogLevel:      "info",
			Runtimes:      []string{"docker-runc", "runc"},
			Mode:          "auto",
			Modes: modesConfig{
				CSV: csvModeConfig{
					MountSpecPath: "/etc/xdxct-container-runtime/host-files-for-container.d",
				},
				CDI: cdiModeConfig{
					DefaultKind:        "xdxct.com/gpu",
					AnnotationPrefixes: []string{cdi.AnnotationPrefix},
					SpecDirs:           cdi.DefaultSpecDirs,
				},
			},
		},
		XDXCTContainerRuntimeHookConfig: RuntimeHookConfig{
			Path: XDXCTContainerRuntimeHookExecutable,
		},
	}
	return &d, nil
}

func getLdConfigPath() string {
	if _, err := os.Stat("/sbin/ldconfig.real"); err == nil {
		return "@/sbin/ldconfig.real"
	}
	return "@/sbin/ldconfig"
}

// getCommentedUserGroup returns whether the nvidia-container-cli user and group config option should be commented.
func getCommentedUserGroup() bool {
	uncommentIf := map[string]bool{
		"suse":     true,
		"opensuse": true,
	}

	idsLike := getDistIDLike()
	for _, id := range idsLike {
		if uncommentIf[id] {
			return false
		}
	}
	return true
}

// getDistIDLike returns the ID_LIKE field from /etc/os-release.
func getDistIDLike() []string {
	releaseFile, err := os.Open("/etc/os-release")
	if err != nil {
		return nil
	}
	defer releaseFile.Close()

	scanner := bufio.NewScanner(releaseFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID_LIKE=") {
			value := strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), "\"")
			return strings.Split(value, " ")
		}
	}
	return nil
}

// ResolveXDXCTCTKPath resolves the path to the xdxct-ctk binary.
// This executable is used in hooks and needs to be an absolute path.
// If the path is specified as an absolute path, it is used directly
// without checking for existence of an executable at that path.
func ResolveXDXCTCTKPath(logger logger.Interface, xdxctCTKPath string) string {
	return resolveWithDefault(
		logger,
		"XDXCT Container Toolkit CLI",
		xdxctCTKPath,
		xdxctCTKDefaultFilePath,
	)
}

// ResolveXDXCTContainerRuntimeHookPath resolves the path the xdxct-container-runtime-hook binary.
func ResolveXDXCTContainerRuntimeHookPath(logger logger.Interface, xdxctContainerRuntimeHookPath string) string {
	return resolveWithDefault(
		logger,
		"XDXCT Container Runtime Hook",
		xdxctContainerRuntimeHookPath,
		xdxctContainerRuntimeHookDefaultPath,
	)
}

// resolveWithDefault resolves the path to the specified binary.
// If an absolute path is specified, it is used directly without searching for the binary.
// If the binary cannot be found in the path, the specified default is used instead.
func resolveWithDefault(logger logger.Interface, label string, path string, defaultPath string) string {
	if filepath.IsAbs(path) {
		logger.Debugf("Using specified %v path %v", label, path)
		return path
	}

	if path == "" {
		path = filepath.Base(defaultPath)
	}
	logger.Debugf("Locating %v as %v", label, path)
	lookup := lookup.NewExecutableLocator(logger, "")

	resolvedPath := defaultPath
	targets, err := lookup.Locate(path)
	if err != nil {
		logger.Warningf("Failed to locate %v: %v", path, err)
	} else {
		logger.Debugf("Found %v candidates: %v", path, targets)
		resolvedPath = targets[0]
	}
	logger.Debugf("Using %v path %v", label, path)

	return resolvedPath
}
