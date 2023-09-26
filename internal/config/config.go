package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
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
	AcceptEnvvarUnprivileged bool `toml:"accept-xdxct-visible-devices-envvar-when-unprivileged"`

	XDXCTContainerCLIConfig         ContainerCLIConfig `toml:"xdxct-container-cli"`
	XDXCTCTKConfig                  CTKConfig          `toml:"xdxct-ctk"`
	XDXCTContainerRuntimeConfig     RuntimeConfig      `toml:"xdxct-container-runtime"`
	XDXCTContainerRuntimeHookConfig RuntimeHookConfig  `toml:"xdxct-container-runtime-hook"`
}

// GetConfig sets up the config struct. Values are read from a toml file
// or set via the environment.
func GetConfig() (*Config, error) {
	if XDGConfigDir := os.Getenv(configOverride); len(XDGConfigDir) != 0 {
		configDir = XDGConfigDir
	}

	configFilePath := path.Join(configDir, configFilePath)

	tomlFile, err := os.Open(configFilePath)
	if err != nil {
		return getDefaultConfig()
	}
	defer tomlFile.Close()

	cfg, err := loadConfigFrom(tomlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config values: %v", err)
	}

	return cfg, nil
}

// loadRuntimeConfigFrom reads the config from the specified Reader
func loadConfigFrom(reader io.Reader) (*Config, error) {
	toml, err := toml.LoadReader(reader)
	if err != nil {
		return nil, err
	}

	return getConfigFrom(toml)
}

// getConfigFrom reads the xdxct container runtime config from the specified toml Tree.
func getConfigFrom(toml *toml.Tree) (*Config, error) {
	cfg, err := getDefaultConfig()
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return cfg, nil
}

// getDefaultConfig defines the default values for the config
func getDefaultConfig() (*Config, error) {
	tomlConfig, err := GetDefaultConfigToml()
	if err != nil {
		return nil, err
	}

	// tomlConfig above includes information about the default values and comments.
	// we need to marshal it back to a string and then unmarshal it to strip the comments.
	contents, err := tomlConfig.ToTomlString()
	if err != nil {
		return nil, err
	}

	reloaded, err := toml.Load(contents)
	if err != nil {
		return nil, err
	}

	d := Config{}
	if err := reloaded.Unmarshal(&d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// The default value for the accept-xdxct-visible-devices-envvar-when-unprivileged is non-standard.
	// As such we explicitly handle it being set here.
	if reloaded.Get("accept-xdxct-visible-devices-envvar-when-unprivileged") == nil {
		d.AcceptEnvvarUnprivileged = true
	}
	// The default value for the xdxct-container-runtime.debug is non-standard.
	// As such we explicitly handle it being set here.
	if reloaded.Get("xdxct-container-runtime.debug") == nil {
		d.XDXCTContainerRuntimeConfig.DebugFilePath = "/dev/null"
	}
	return &d, nil
}

// GetDefaultConfigToml returns the default config as a toml Tree.
func GetDefaultConfigToml() (*toml.Tree, error) {
	tree, err := toml.TreeFromMap(nil)
	if err != nil {
		return nil, err
	}

	tree.Set("disable-require", false)
	tree.SetWithComment("swarm-resource", "", true, "DOCKER_RESOURCE_GPU")
	tree.SetWithComment("accept-xdxct-visible-devices-envvar-when-unprivileged", "", true, true)
	tree.SetWithComment("accept-xdxct-visible-devices-as-volume-mounts", "", true, false)

	// xdxct-container-cli
	tree.SetWithComment("xdxct-container-cli.root", "", true, "/run/xdxct/driver")
	tree.SetWithComment("xdxct-container-cli.path", "", true, "/usr/bin/xdxct-container-cli")
	tree.Set("xdxct-container-cli.environment", []string{})
	tree.SetWithComment("xdxct-container-cli.debug", "", true, "/var/log/xdxct-container-toolkit.log")
	tree.SetWithComment("xdxct-container-cli.ldcache", "", true, "/etc/ld.so.cache")
	tree.Set("xdxct-container-cli.load-kmods", true)
	tree.SetWithComment("xdxct-container-cli.no-cgroups", "", true, false)

	tree.SetWithComment("xdxct-container-cli.user", "", getCommentedUserGroup(), getUserGroup())
	tree.Set("xdxct-container-cli.ldconfig", getLdConfigPath())

	// xdxct-container-runtime
	tree.SetWithComment("xdxct-container-runtime.debug", "", true, "/var/log/xdxct-container-runtime.log")
	tree.Set("xdxct-container-runtime.log-level", "info")

	commentLines := []string{
		"Specify the runtimes to consider. This list is processed in order and the PATH",
		"searched for matching executables unless the entry is an absolute path.",
	}
	tree.SetWithComment("xdxct-container-runtime.runtimes", strings.Join(commentLines, "\n "), false, []string{"docker-runc", "runc"})

	tree.Set("xdxct-container-runtime.mode", "auto")

	tree.Set("xdxct-container-runtime.modes.csv.mount-spec-path", "/etc/xdxct-container-runtime/host-files-for-container.d")
	tree.Set("xdxct-container-runtime.modes.cdi.default-kind", "xdxct.com/gpu")
	tree.Set("xdxct-container-runtime.modes.cdi.annotation-prefixes", []string{cdi.AnnotationPrefix})

	// xdxct-ctk
	tree.Set("xdxct-ctk.path", xdxctCTKExecutable)

	// xdxct-container-runtime-hook
	tree.Set("xdxct-container-runtime-hook.path", xdxctContainerRuntimeHookExecutable)

	return tree, nil
}

func getLdConfigPath() string {
	if _, err := os.Stat("/sbin/ldconfig.real"); err == nil {
		return "@/sbin/ldconfig.real"
	}
	return "@/sbin/ldconfig"
}

// getUserGroup returns the user and group to use for the xdxct-container-cli and whether the config option should be commented.
func getUserGroup() string {
	return "root:video"
}

// getCommentedUserGroup returns whether the xdxct-container-cli user and group config option should be commented.
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
func ResolveXDXCTCTKPath(logger *logrus.Logger, xdxctCTKPath string) string {
	return resolveWithDefault(
		logger,
		"XDXCT Container Toolkit CLI",
		xdxctCTKPath,
		xdxctCTKDefaultFilePath,
	)
}

// ResolveXDXCTContainerRuntimeHookPath resolves the path the xdxct-container-runtime-hook binary.
func ResolveXDXCTContainerRuntimeHookPath(logger *logrus.Logger, xdxctContainerRuntimeHookPath string) string {
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
func resolveWithDefault(logger *logrus.Logger, label string, path string, defaultPath string) string {
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
		logger.Warnf("Failed to locate %v: %v", path, err)
	} else {
		logger.Debugf("Found %v candidates: %v", path, targets)
		resolvedPath = targets[0]
	}
	logger.Debugf("Using %v path %v", label, path)

	return resolvedPath
}
