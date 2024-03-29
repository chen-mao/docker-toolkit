package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi"
	transformroot "github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform/root"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"tags.cncf.io/container-device-interface/pkg/cdi"
)

const (
	// TODO: DefaultXdxctDriverRoot specifies the default XDXCT driver run directory
	// DefaultXdxctDriverRoot = "/run/xdxct/driver"
	DefaultXdxctDriverRoot = "/"

	xdxctContainerCliSource         = "/usr/bin/xdxct-container-cli"
	xdxctContainerRuntimeHookSource = "/usr/bin/xdxct-container-runtime-hook"

	xdxctContainerToolkitConfigSource = "/etc/xdxct-container-runtime/config.toml"
	configFilename                    = "config.toml"
)

type options struct {
	DriverRoot        string
	DriverRootCtrPath string

	ContainerRuntimeMode     string
	ContainerRuntimeDebug    string
	ContainerRuntimeLogLevel string

	ContainerRuntimeModesCdiDefaultKind        string
	ContainerRuntimeModesCDIAnnotationPrefixes cli.StringSlice

	ContainerRuntimeRuntimes cli.StringSlice

	ContainerRuntimeHookSkipModeDetection bool

	ContainerCLIDebug string
	toolkitRoot       string

	cdiEnabled   bool
	cdiOutputDir string
	cdiKind      string
	cdiVendor    string
	cdiClass     string

	acceptXDXCTVisibleDevicesWhenUnprivileged bool
	acceptXDXCTVisibleDevicesAsVolumeMounts   bool

	ignoreErrors bool
}

func main() {

	opts := options{}

	// Create the top-level CLI
	c := cli.NewApp()
	c.Name = "toolkit"
	c.Usage = "Manage the XDXCT container toolkit"
	c.Version = "0.1.0"

	// Create the 'install' subcommand
	install := cli.Command{}
	install.Name = "install"
	install.Usage = "Install the components of the XDXCT container toolkit"
	install.ArgsUsage = "<toolkit_directory>"
	install.Before = func(c *cli.Context) error {
		return validateOptions(c, &opts)
	}
	install.Action = func(c *cli.Context) error {
		return Install(c, &opts)
	}

	// Create the 'delete' command
	delete := cli.Command{}
	delete.Name = "delete"
	delete.Usage = "Delete the XDXCT container toolkit"
	delete.ArgsUsage = "<toolkit_directory>"
	delete.Before = func(c *cli.Context) error {
		return validateOptions(c, &opts)
	}
	delete.Action = func(c *cli.Context) error {
		return Delete(c, &opts)
	}

	// Register the subcommand with the top-level CLI
	c.Commands = []*cli.Command{
		&install,
		&delete,
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "xdxct-driver-root",
			Value:       DefaultXdxctDriverRoot,
			Destination: &opts.DriverRoot,
			EnvVars:     []string{"XDXCT_DRIVER_ROOT"},
		},
		&cli.StringFlag{
			Name:        "driver-root-ctr-path",
			Value:       DefaultXdxctDriverRoot,
			Destination: &opts.DriverRootCtrPath,
			EnvVars:     []string{"DRIVER_ROOT_CTR_PATH"},
		},
		&cli.StringFlag{
			Name:        "xdxct-container-runtime.debug",
			Aliases:     []string{"xdxct-container-runtime-debug"},
			Usage:       "Specify the location of the debug log file for the XDXCT Container Runtime",
			Destination: &opts.ContainerRuntimeDebug,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_DEBUG"},
		},
		&cli.StringFlag{
			Name:        "xdxct-container-runtime.log-level",
			Aliases:     []string{"xdxct-container-runtime-debug-log-level"},
			Destination: &opts.ContainerRuntimeLogLevel,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_LOG_LEVEL"},
		},
		&cli.StringFlag{
			Name:        "xdxct-container-runtime.mode",
			Aliases:     []string{"xdxct-container-runtime-mode"},
			Destination: &opts.ContainerRuntimeMode,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_MODE"},
		},
		&cli.StringFlag{
			Name:        "xdxct-container-runtime.modes.cdi.default-kind",
			Destination: &opts.ContainerRuntimeModesCdiDefaultKind,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_MODES_CDI_DEFAULT_KIND"},
		},
		&cli.StringSliceFlag{
			Name:        "xdxct-container-runtime.modes.cdi.annotation-prefixes",
			Destination: &opts.ContainerRuntimeModesCDIAnnotationPrefixes,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_MODES_CDI_ANNOTATION_PREFIXES"},
		},
		&cli.StringSliceFlag{
			Name:        "xdxct-container-runtime.runtimes",
			Destination: &opts.ContainerRuntimeRuntimes,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_RUNTIMES"},
		},
		&cli.BoolFlag{
			Name:        "xdxct-container-runtime-hook.skip-mode-detection",
			Value:       true,
			Destination: &opts.ContainerRuntimeHookSkipModeDetection,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_HOOK_SKIP_MODE_DETECTION"},
		},
		&cli.StringFlag{
			Name:        "xdxct-container-cli.debug",
			Aliases:     []string{"xdxct-container-cli-debug"},
			Usage:       "Specify the location of the debug log file for the XDXCT Container CLI",
			Destination: &opts.ContainerCLIDebug,
			EnvVars:     []string{"XDXCT_CONTAINER_CLI_DEBUG"},
		},
		&cli.BoolFlag{
			Name:        "accept-xdxct-visible-devices-envvar-when-unprivileged",
			Usage:       "Set the accept-xdxct-visible-devices-envvar-when-unprivileged config option",
			Value:       true,
			Destination: &opts.acceptXDXCTVisibleDevicesWhenUnprivileged,
			EnvVars:     []string{"ACCEPT_XDXCT_VISIBLE_DEVICES_ENVVAR_WHEN_UNPRIVILEGED"},
		},
		&cli.BoolFlag{
			Name:        "accept-xdxct-visible-devices-as-volume-mounts",
			Usage:       "Set the accept-xdxct-visible-devices-as-volume-mounts config option",
			Destination: &opts.acceptXDXCTVisibleDevicesAsVolumeMounts,
			EnvVars:     []string{"ACCEPT_XDXCT_VISIBLE_DEVICES_AS_VOLUME_MOUNTS"},
		},
		&cli.StringFlag{
			Name:        "toolkit-root",
			Usage:       "The directory where the XDXCT Container toolkit is to be installed",
			Required:    true,
			Destination: &opts.toolkitRoot,
			EnvVars:     []string{"TOOLKIT_ROOT"},
		},
		&cli.BoolFlag{
			Name:        "cdi-enabled",
			Aliases:     []string{"enable-cdi"},
			Usage:       "enable the generation of a CDI specification",
			Destination: &opts.cdiEnabled,
			EnvVars:     []string{"CDI_ENABLED", "ENABLE_CDI"},
		},
		&cli.StringFlag{
			Name:        "cdi-output-dir",
			Usage:       "the directory where the CDI output files are to be written. If this is set to '', no CDI specification is generated.",
			Value:       "/var/run/cdi",
			Destination: &opts.cdiOutputDir,
			EnvVars:     []string{"CDI_OUTPUT_DIR"},
		},
		&cli.StringFlag{
			Name:        "cdi-kind",
			Usage:       "the vendor string to use for the generated CDI specification",
			Value:       "management.xdxct.com/gpu",
			Destination: &opts.cdiKind,
			EnvVars:     []string{"CDI_KIND"},
		},
		&cli.BoolFlag{
			Name:        "ignore-errors",
			Usage:       "ignore errors when installing the XDXCT Container toolkit. This is used for testing purposes only.",
			Hidden:      true,
			Destination: &opts.ignoreErrors,
		},
	}

	// Update the subcommand flags with the common subcommand flags
	install.Flags = append([]cli.Flag{}, flags...)
	delete.Flags = append([]cli.Flag{}, flags...)

	// Run the top-level CLI
	if err := c.Run(os.Args); err != nil {
		log.Fatal(fmt.Errorf("error: %v", err))
	}
}

// validateOptions checks whether the specified options are valid
func validateOptions(c *cli.Context, opts *options) error {
	if opts.toolkitRoot == "" {
		return fmt.Errorf("invalid --toolkit-root option: %v", opts.toolkitRoot)
	}

	vendor, class := cdi.ParseQualifier(opts.cdiKind)
	if err := cdi.ValidateVendorName(vendor); err != nil {
		return fmt.Errorf("invalid CDI vendor name: %v", err)
	}
	if err := cdi.ValidateClassName(class); err != nil {
		return fmt.Errorf("invalid CDI class name: %v", err)
	}
	opts.cdiVendor = vendor
	opts.cdiClass = class

	return nil
}

// Delete removes the XDXCT container toolkit
func Delete(cli *cli.Context, opts *options) error {
	log.Infof("Deleting XDXCT container toolkit from '%v'", opts.toolkitRoot)
	err := os.RemoveAll(opts.toolkitRoot)
	if err != nil {
		return fmt.Errorf("error deleting toolkit directory: %v", err)
	}
	return nil
}

// Install installs the components of the XDXCT container toolkit.
// Any existing installation is removed.
func Install(cli *cli.Context, opts *options) error {
	log.Infof("Installing XDXCT container toolkit to '%v'", opts.toolkitRoot)

	log.Infof("Removing existing XDXCT container toolkit installation")
	err := os.RemoveAll(opts.toolkitRoot)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("error removing toolkit directory: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("error removing toolkit directory: %v", err))
	}

	toolkitConfigDir := filepath.Join(opts.toolkitRoot, ".config", "xdxct-container-runtime")
	toolkitConfigPath := filepath.Join(toolkitConfigDir, configFilename)

	log.Infof("toolkitConfigDir: %v, toolkitConfigPath: %v", toolkitConfigDir, toolkitConfigPath)

	// 创建:/usr/local/xdxct/toolkit/.config/xdxct-container-runtime 目录
	err = createDirectories(opts.toolkitRoot, toolkitConfigDir)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("could not create required directories: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("could not create required directories: %v", err))
	}

	// 安装 libxdxct-container-go.so.1.14.0
	err = installContainerLibraries(opts.toolkitRoot)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("error installing XDXCT container library: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("error installing XDXCT container library: %v", err))
	}
	// 安装 runtimes
	err = installContainerRuntimes(opts.toolkitRoot, opts.DriverRoot)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("error installing XDXCT container runtime: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("error installing XDXCT container runtime: %v", err))
	}
	// 安装 xdxct-container-cli
	xdxctContainerCliExecutable, err := installContainerCLI(opts.toolkitRoot)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("error installing XDXCT container CLI: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("error installing XDXCT container CLI: %v", err))
	}
	// 安装 xdxct-container-hook,创建link xdxct-container-toolkit -> xdxct-container-runtime-hook
	xdxctContainerRuntimeHookPath, err := installRuntimeHook(opts.toolkitRoot, toolkitConfigPath)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("error installing XDXCT container runtime hook: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("error installing XDXCT container runtime hook: %v", err))
	}
	// 安装 xdxct-ctk
	xdxctCTKPath, err := installContainerToolkitCLI(opts.toolkitRoot)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("error installing XDXCT Container Toolkit CLI: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("error installing XDXCT Container Toolkit CLI: %v", err))
	}
	// 安装 config.toml配置文件
	err = installToolkitConfig(cli, toolkitConfigPath, xdxctContainerCliExecutable, xdxctCTKPath, xdxctContainerRuntimeHookPath, opts)
	if err != nil && !opts.ignoreErrors {
		return fmt.Errorf("error installing XDXCT container toolkit config: %v", err)
	} else if err != nil {
		log.Errorf("Ignoring error: %v", fmt.Errorf("error installing XDXCT container toolkit config: %v", err))
	}
	// 生成cdi spec, return nil
	return generateCDISpec(opts, xdxctCTKPath)
}

// installContainerLibraries locates and installs the libraries that are part of
// the xdxct-container-toolkit.
// A predefined set of library candidates are considered, with the first one
// resulting in success being installed to the toolkit folder. The install process
// resolves the symlink for the library and copies the versioned library itself.
func installContainerLibraries(toolkitRoot string) error {
	log.Infof("Installing XDXCT container library to '%v'", toolkitRoot)

	libs := []string{
		"libxdxct-container.so.1",
		"libxdxct-container-go.so.1",
	}

	for _, l := range libs {
		err := installLibrary(l, toolkitRoot)
		if err != nil {
			return fmt.Errorf("failed to install %s: %v", l, err)
		}
	}

	return nil
}

// installLibrary installs the specified library to the toolkit directory.
func installLibrary(libName string, toolkitRoot string) error {
	// 在容器中locate /usr/lib/x86_64-linux-gnu/libxdxct-container.so.1
	// 然后解析出 libraryPath = libxdxct-container.so.1.14.0
	libraryPath, err := findLibrary("", libName)
	if err != nil {
		return fmt.Errorf("error locating XDXCT container library: %v", err)
	}
	// 相当于cp libraryPath toolkitRoot
	installedLibPath, err := installFileToFolder(toolkitRoot, libraryPath)
	if err != nil {
		return fmt.Errorf("error installing %v to %v: %v", libraryPath, toolkitRoot, err)
	}
	log.Infof("Installed '%v' to '%v'", libraryPath, installedLibPath)

	if filepath.Base(installedLibPath) == libName {
		return nil
	}
	// 给libxdxct-container.so.1.14.0 创建 libxdxct-container.so.1软链接
	err = installSymlink(toolkitRoot, libName, installedLibPath)
	if err != nil {
		return fmt.Errorf("error installing symlink for XDXCT container library: %v", err)
	}

	return nil
}

// installToolkitConfig installs the config file for the XDXCT container toolkit ensuring
// that the settings are updated to match the desired install and xdxct driver directories.
// toolkitConfigPath： /usr/local/xdxct/toolkit/.config/xdxct-container-runtime/config.toml
func installToolkitConfig(c *cli.Context, toolkitConfigPath string, xdxctContainerCliExecutablePath string, xdxctCTKPath string, xdxctContainerRuntimeHookPath string, opts *options) error {
	log.Infof("Installing XDXCT container toolkit config '%v'", toolkitConfigPath)

	config, err := loadConfig(xdxctContainerToolkitConfigSource)
	if err != nil {
		return fmt.Errorf("could not open source config file: %v", err)
	}

	targetConfig, err := os.Create(toolkitConfigPath)
	if err != nil {
		return fmt.Errorf("could not create target config file: %v", err)
	}
	defer targetConfig.Close()

	// Read the ldconfig path from the config as this may differ per platform
	// On ubuntu-based systems this ends in `.real`

	ldconfigPath := fmt.Sprintf("%s", config.GetDefault("xdxct-container-cli.ldconfig", getLdConfigPath()))
	// ldconfigPath := fmt.Sprintf("%s", config.GetDefault("xdxct-container-cli.ldconfig", "/sbin/ldconfig.real"))
	// Use the driver run root as the root:
	driverLdconfigPath := "@" + filepath.Join(opts.DriverRoot, strings.TrimPrefix(ldconfigPath, "@/"))

	configValues := map[string]interface{}{
		// Set the options in the root toml table
		"accept-xdxct-visible-devices-envvar-when-unprivileged": opts.acceptXDXCTVisibleDevicesWhenUnprivileged,
		"accept-xdxct-visible-devices-as-volume-mounts":         opts.acceptXDXCTVisibleDevicesAsVolumeMounts,
		// Set the xdxct-container-cli options
		"xdxct-container-cli.root":     opts.DriverRoot,
		"xdxct-container-cli.path":     xdxctContainerCliExecutablePath,
		"xdxct-container-cli.ldconfig": driverLdconfigPath,
		// Set xdxct-ctk options
		"xdxct-ctk.path": xdxctCTKPath,
		// Set the xdxct-container-runtime-hook options
		"xdxct-container-runtime-hook.path":                xdxctContainerRuntimeHookPath,
		"xdxct-container-runtime-hook.skip-mode-detection": opts.ContainerRuntimeHookSkipModeDetection,
	}
	for key, value := range configValues {
		config.Set(key, value)
	}

	// Set the optional config options
	optionalConfigValues := map[string]interface{}{
		"xdxct-container-runtime.debug":                         opts.ContainerRuntimeDebug,
		"xdxct-container-runtime.log-level":                     opts.ContainerRuntimeLogLevel,
		"xdxct-container-runtime.mode":                          opts.ContainerRuntimeMode,
		"xdxct-container-runtime.modes.cdi.annotation-prefixes": opts.ContainerRuntimeModesCDIAnnotationPrefixes,
		"xdxct-container-runtime.modes.cdi.default-kind":        opts.ContainerRuntimeModesCdiDefaultKind,
		"xdxct-container-runtime.runtimes":                      opts.ContainerRuntimeRuntimes,
		"xdxct-container-cli.debug":                             opts.ContainerCLIDebug,
	}
	for key, value := range optionalConfigValues {
		if !c.IsSet(key) {
			log.Infof("Skipping unset option: %v", key)
			continue
		}
		if value == nil {
			log.Infof("Skipping option with nil value: %v", key)
			continue
		}

		switch v := value.(type) {
		case string:
			if v == "" {
				continue
			}
		case cli.StringSlice:
			if len(v.Value()) == 0 {
				continue
			}
			value = v.Value()
		default:
			log.Warnf("Unexpected type for option %v=%v: %T", key, value, v)
		}

		config.Set(key, value)
	}

	_, err = config.WriteTo(targetConfig)
	if err != nil {
		return fmt.Errorf("error writing config: %v", err)
	}

	os.Stdout.WriteString("Using config:\n")
	config.WriteTo(os.Stdout)

	return nil
}

func loadConfig(path string) (*toml.Tree, error) {
	_, err := os.Stat(path)
	if err == nil {
		return toml.LoadFile(path)
	} else if os.IsNotExist(err) {
		return toml.TreeFromMap(nil)
	}
	return nil, err
}

// installContainerToolkitCLI installs the xdxct-ctk CLI executable and wrapper.
func installContainerToolkitCLI(toolkitDir string) (string, error) {
	e := executable{
		source: "/usr/bin/xdxct-ctk",
		target: executableTarget{
			dotfileName: "xdxct-ctk.real",
			wrapperName: "xdxct-ctk",
		},
	}

	return e.install(toolkitDir)
}

// installContainerCLI sets up the XDXCT container CLI executable, copying the executable
// and implementing the required wrapper
func installContainerCLI(toolkitRoot string) (string, error) {
	log.Infof("Installing XDXCT container CLI from '%v'", xdxctContainerCliSource)

	env := map[string]string{
		"LD_LIBRARY_PATH": toolkitRoot,
	}

	e := executable{
		source: xdxctContainerCliSource,
		target: executableTarget{
			dotfileName: "xdxct-container-cli.real",
			wrapperName: "xdxct-container-cli",
		},
		env: env,
	}

	installedPath, err := e.install(toolkitRoot)
	if err != nil {
		return "", fmt.Errorf("error installing XDXCT container CLI: %v", err)
	}
	return installedPath, nil
}

// installRuntimeHook sets up the XDXCT runtime hook, copying the executable
// and implementing the required wrapper
// /etc/.config/xdxct-container-runtime/config.toml
func installRuntimeHook(toolkitRoot string, configFilePath string) (string, error) {
	log.Infof("Installing XDXCT container runtime hook from '%v'", xdxctContainerRuntimeHookSource)

	argLines := []string{
		fmt.Sprintf("-config \"%s\"", configFilePath),
	}

	e := executable{
		source: xdxctContainerRuntimeHookSource,
		target: executableTarget{
			dotfileName: "xdxct-container-runtime-hook.real",
			wrapperName: "xdxct-container-runtime-hook",
		},
		argLines: argLines,
	}

	installedPath, err := e.install(toolkitRoot)
	if err != nil {
		return "", fmt.Errorf("error installing XDXCT container runtime hook: %v", err)
	}
	// 创建软连接xdxct-container-toolkit --> xdxct-container-runtime-hook
	err = installSymlink(toolkitRoot, "xdxct-container-toolkit", installedPath)
	if err != nil {
		return "", fmt.Errorf("error installing symlink to XDXCT container runtime hook: %v", err)
	}

	return installedPath, nil
}

// installSymlink creates a symlink in the toolkitDirectory that points to the specified target.
// Note: The target is assumed to be local to the toolkit directory
func installSymlink(toolkitRoot string, link string, target string) error {
	symlinkPath := filepath.Join(toolkitRoot, link)
	targetPath := filepath.Base(target)
	log.Infof("Creating symlink '%v' -> '%v'", symlinkPath, targetPath)

	err := os.Symlink(targetPath, symlinkPath)
	if err != nil {
		return fmt.Errorf("error creating symlink '%v' => '%v': %v", symlinkPath, targetPath, err)
	}
	return nil
}

// installFileToFolder copies a source file to a destination folder.
// The path of the input file is ignored.
// e.g. installFileToFolder("/some/path/file.txt", "/output/path")
// will result in a file "/output/path/file.txt" being generated
func installFileToFolder(destFolder string, src string) (string, error) {
	name := filepath.Base(src)
	return installFileToFolderWithName(destFolder, name, src)
}

// cp src destFolder/name: cp libxdxct-container.so.1.14.0 /usr/local/xdxct/toolkit
func installFileToFolderWithName(destFolder string, name, src string) (string, error) {
	dest := filepath.Join(destFolder, name)
	err := installFile(dest, src)
	if err != nil {
		return "", fmt.Errorf("error copying '%v' to '%v': %v", src, dest, err)
	}
	return dest, nil
}

// installFile copies a file from src to dest and maintains
// file modes
func installFile(dest string, src string) error {
	log.Infof("Installing '%v' to '%v'", src, dest)

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source: %v", err)
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error creating destination: %v", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	err = applyModeFromSource(dest, src)
	if err != nil {
		return fmt.Errorf("error setting destination file mode: %v", err)
	}
	return nil
}

// applyModeFromSource sets the file mode for a destination file
// to match that of a specified source file
func applyModeFromSource(dest string, src string) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting file info for '%v': %v", src, err)
	}
	err = os.Chmod(dest, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("error setting mode for '%v': %v", dest, err)
	}
	return nil
}

// findLibrary searches a set of candidate libraries in the specified root for
// a given library name
func findLibrary(root string, libName string) (string, error) {
	log.Infof("Finding library %v (root=%v)", libName, root)

	candidateDirs := []string{
		"/usr/lib64",
		"/usr/lib/x86_64-linux-gnu",
		"/usr/lib/aarch64-linux-gnu",
	}

	for _, d := range candidateDirs {
		l := filepath.Join(root, d, libName)
		log.Infof("Checking library candidate '%v'", l)

		libraryCandidate, err := resolveLink(l)
		if err != nil {
			log.Infof("Skipping library candidate '%v': %v", l, err)
			continue
		}

		return libraryCandidate, nil
	}

	return "", fmt.Errorf("error locating library '%v'", libName)
}

// resolveLink finds the target of a symlink or the file itself in the
// case of a regular file.
// This is equivalent to running `readlink -f ${l}`
func resolveLink(l string) (string, error) {
	resolved, err := filepath.EvalSymlinks(l)
	if err != nil {
		return "", fmt.Errorf("error resolving link '%v': %v", l, err)
	}
	if l != resolved {
		log.Infof("Resolved link: '%v' => '%v'", l, resolved)
	}
	return resolved, nil
}

func createDirectories(dir ...string) error {
	for _, d := range dir {
		log.Infof("Creating directory '%v'", d)
		err := os.MkdirAll(d, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}
	return nil
}

func getLdConfigPath() string {
	if _, err := os.Stat("/sbin/ldconfig.real"); err == nil {
		return "@/sbin/ldconfig.real"
	}
	return "@/sbin/ldconfig"
}

// generateCDISpec generates a CDI spec for use in managemnt containers
func generateCDISpec(opts *options, xdxctCTKPath string) error {
	if !opts.cdiEnabled {
		return nil
	}
	if opts.cdiOutputDir == "" {
		log.Info("Skipping CDI spec generation (no output directory specified)")
		return nil
	}

	log.Info("Generating CDI spec for management containers")
	cdilib, err := xdxcdi.New(
		xdxcdi.WithMode(xdxcdi.ModeManagement),
		xdxcdi.WithDriverRoot(opts.DriverRootCtrPath),
		xdxcdi.WithXDXCTCTKPath(xdxctCTKPath),
		xdxcdi.WithVendor(opts.cdiVendor),
		xdxcdi.WithClass(opts.cdiClass),
	)
	if err != nil {
		return fmt.Errorf("failed to create CDI library for management containers: %v", err)
	}

	spec, err := cdilib.GetSpec()
	if err != nil {
		return fmt.Errorf("failed to genereate CDI spec for management containers: %v", err)
	}
	err = transformroot.New(
		transformroot.WithRoot(opts.DriverRootCtrPath),
		transformroot.WithTargetRoot(opts.DriverRoot),
	).Transform(spec.Raw())
	if err != nil {
		return fmt.Errorf("failed to transform driver root in CDI spec: %v", err)
	}

	name, err := cdi.GenerateNameForSpec(spec.Raw())
	if err != nil {
		return fmt.Errorf("failed to generate CDI name for management containers: %v", err)
	}
	err = spec.Save(filepath.Join(opts.cdiOutputDir, name))
	if err != nil {
		return fmt.Errorf("failed to save CDI spec for management containers: %v", err)
	}

	return nil
}
