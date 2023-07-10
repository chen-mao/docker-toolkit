/**
# Copyright (c) 2021-2022, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/internal/config"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine/crio"
	"github.com/XDXCT/xdxct-container-toolkit/tools/container/operator"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

const (
	restartModeSystemd = "systemd"
	restartModeNone    = "none"

	defaultConfigMode = "hook"

	// Hook-based settings
	defaultHooksDir     = "/usr/share/containers/oci/hooks.d"
	defaultHookFilename = "oci-nvidia-hook.json"

	// Config-based settings
	defaultConfig        = "/etc/crio/crio.conf"
	defaultRuntimeClass  = "nvidia"
	defaultSetAsDefault  = true
	defaultRestartMode   = restartModeSystemd
	defaultHostRootMount = "/host"
)

// options stores the configuration from the command linek or environment variables
type options struct {
	configMode string

	hooksDir     string
	hookFilename string
	runtimeDir   string

	config        string
	runtimeClass  string
	setAsDefault  bool
	restartMode   string
	hostRootMount string
}

func main() {
	options := options{}

	// Create the top-level CLI
	c := cli.NewApp()
	c.Name = "crio"
	c.Usage = "Update cri-o hooks to include the NVIDIA runtime hook"
	c.ArgsUsage = "<toolkit_dirname>"
	c.Version = "0.1.0"

	// Create the 'setup' subcommand
	setup := cli.Command{}
	setup.Name = "setup"
	setup.Usage = "Configure cri-o for NVIDIA GPU containers"
	setup.ArgsUsage = "<toolkit_dirname>"
	setup.Action = func(c *cli.Context) error {
		return Setup(c, &options)
	}
	setup.Before = func(c *cli.Context) error {
		return ParseArgs(c, &options)
	}

	// Create the 'cleanup' subcommand
	cleanup := cli.Command{}
	cleanup.Name = "cleanup"
	cleanup.Usage = "Remove the NVIDIA-specific cri-o configuration"
	cleanup.Action = func(c *cli.Context) error {
		return Cleanup(c, &options)
	}
	// Register the subcommands with the top-level CLI
	c.Commands = []*cli.Command{
		&setup,
		&cleanup,
	}

	// Setup common flags across both subcommands. All subcommands get the same
	// set of flags even if they don't use some of them. This is so that we
	// only require the user to specify one set of flags for both 'startup'
	// and 'cleanup' to simplify things.
	commonFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "hooks-dir",
			Aliases:     []string{"d"},
			Usage:       "path to the cri-o hooks directory",
			Value:       defaultHooksDir,
			Destination: &options.hooksDir,
			EnvVars:     []string{"CRIO_HOOKS_DIR"},
			DefaultText: defaultHooksDir,
		},
		&cli.StringFlag{
			Name:        "hook-filename",
			Aliases:     []string{"f"},
			Usage:       "filename of the cri-o hook that will be created / removed in the hooks directory",
			Value:       defaultHookFilename,
			Destination: &options.hookFilename,
			EnvVars:     []string{"CRIO_HOOK_FILENAME"},
			DefaultText: defaultHookFilename,
		},
		&cli.StringFlag{
			Name:        "config-mode",
			Usage:       "the configuration mode to use. One of [hook | config]",
			Value:       defaultConfigMode,
			Destination: &options.configMode,
			EnvVars:     []string{"CRIO_CONFIG_MODE"},
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "Path to the cri-o config file",
			Value:       defaultConfig,
			Destination: &options.config,
			EnvVars:     []string{"CRIO_CONFIG"},
		},
		&cli.StringFlag{
			Name:        "runtime-class",
			Usage:       "The name of the runtime class to set for the nvidia-container-runtime",
			Value:       defaultRuntimeClass,
			Destination: &options.runtimeClass,
			EnvVars:     []string{"CRIO_RUNTIME_CLASS"},
		},
		// The flags below are only used by the 'setup' command.
		&cli.BoolFlag{
			Name:        "set-as-default",
			Usage:       "Set nvidia-container-runtime as the default runtime",
			Value:       defaultSetAsDefault,
			Destination: &options.setAsDefault,
			EnvVars:     []string{"CRIO_SET_AS_DEFAULT"},
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:        "restart-mode",
			Usage:       "Specify how cri-o should be restarted;  If 'none' is selected, it will not be restarted [systemd | none]",
			Value:       defaultRestartMode,
			Destination: &options.restartMode,
			EnvVars:     []string{"CRIO_RESTART_MODE"},
		},
		&cli.StringFlag{
			Name:        "host-root",
			Usage:       "Specify the path to the host root to be used when restarting crio using systemd",
			Value:       defaultHostRootMount,
			Destination: &options.hostRootMount,
			EnvVars:     []string{"HOST_ROOT_MOUNT"},
		},
	}

	// Update the subcommand flags with the common subcommand flags
	setup.Flags = append([]cli.Flag{}, commonFlags...)
	cleanup.Flags = append([]cli.Flag{}, commonFlags...)

	// Run the top-level CLI
	if err := c.Run(os.Args); err != nil {
		log.Fatal(fmt.Errorf("error: %v", err))
	}
}

// Setup installs the prestart hook required to launch GPU-enabled containers
func Setup(c *cli.Context, o *options) error {
	log.Infof("Starting 'setup' for %v", c.App.Name)

	switch o.configMode {
	case "hook":
		return setupHook(o)
	case "config":
		return setupConfig(o)
	default:
		return fmt.Errorf("invalid config-mode '%v'", o.configMode)
	}
}

// setupHook installs the prestart hook required to launch GPU-enabled containers
func setupHook(o *options) error {
	log.Infof("Installing prestart hook")

	err := os.MkdirAll(o.hooksDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating hooks directory %v: %v", o.hooksDir, err)
	}

	hookPath := getHookPath(o.hooksDir, o.hookFilename)
	err = createHook(o.runtimeDir, hookPath)
	if err != nil {
		return fmt.Errorf("error creating hook: %v", err)
	}

	return nil
}

// setupConfig updates the cri-o config for the NVIDIA container runtime
func setupConfig(o *options) error {
	log.Infof("Updating config file")

	cfg, err := crio.New(
		crio.WithPath(o.config),
	)
	if err != nil {
		return fmt.Errorf("unable to load config: %v", err)
	}

	err = UpdateConfig(cfg, o)
	if err != nil {
		return fmt.Errorf("unable to update config: %v", err)
	}

	log.Infof("Flushing cri-o config to %v", o.config)
	n, err := cfg.Save(o.config)
	if err != nil {
		return fmt.Errorf("unable to flush config: %v", err)
	}
	if n == 0 {
		log.Infof("Config file is empty, removed")
	}

	err = RestartCrio(o)
	if err != nil {
		return fmt.Errorf("unable to restart crio: %v", err)
	}

	return nil
}

// Cleanup removes the specified prestart hook
func Cleanup(c *cli.Context, o *options) error {
	log.Infof("Starting 'cleanup' for %v", c.App.Name)

	switch o.configMode {
	case "hook":
		return cleanupHook(o)
	case "config":
		return cleanupConfig(o)
	default:
		return fmt.Errorf("invalid config-mode '%v'", o.configMode)
	}
}

// cleanupHook removes the prestart hook
func cleanupHook(o *options) error {
	log.Infof("Removing prestart hook")

	hookPath := getHookPath(o.hooksDir, o.hookFilename)
	err := os.Remove(hookPath)
	if err != nil {
		return fmt.Errorf("error removing hook '%v': %v", hookPath, err)
	}

	return nil
}

// cleanupConfig removes the NVIDIA container runtime from the cri-o config
func cleanupConfig(o *options) error {
	log.Infof("Reverting config file modifications")

	cfg, err := crio.New(
		crio.WithPath(o.config),
	)
	if err != nil {
		return fmt.Errorf("unable to load config: %v", err)
	}

	err = RevertConfig(cfg, o)
	if err != nil {
		return fmt.Errorf("unable to update config: %v", err)
	}

	log.Infof("Flushing cri-o config to %v", o.config)
	n, err := cfg.Save(o.config)
	if err != nil {
		return fmt.Errorf("unable to flush config: %v", err)
	}
	if n == 0 {
		log.Infof("Config file is empty, removed")
	}

	err = RestartCrio(o)
	if err != nil {
		return fmt.Errorf("unable to restart crio: %v", err)
	}

	return nil
}

// ParseArgs parses the command line arguments to the CLI
func ParseArgs(c *cli.Context, o *options) error {
	args := c.Args()

	log.Infof("Parsing arguments: %v", args.Slice())
	if c.NArg() != 1 {
		return fmt.Errorf("incorrect number of arguments")
	}
	o.runtimeDir = args.Get(0)
	log.Infof("Successfully parsed arguments")

	return nil
}

func createHook(toolkitDir string, hookPath string) error {
	hook, err := os.Create(hookPath)
	if err != nil {
		return fmt.Errorf("error creating hook file '%v': %v", hookPath, err)
	}
	defer hook.Close()

	encoder := json.NewEncoder(hook)
	err = encoder.Encode(generateOciHook(toolkitDir))
	if err != nil {
		return fmt.Errorf("error writing hook file '%v': %v", hookPath, err)
	}
	return nil
}

func getHookPath(hooksDir string, hookFilename string) string {
	return filepath.Join(hooksDir, hookFilename)
}

func generateOciHook(toolkitDir string) podmanHook {
	hookPath := filepath.Join(toolkitDir, config.NVIDIAContainerRuntimeHookExecutable)
	envPath := "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:" + toolkitDir
	always := true

	hook := podmanHook{
		Version: "1.0.0",
		Stages:  []string{"prestart"},
		Hook: specHook{
			Path: hookPath,
			Args: []string{filepath.Base(config.NVIDIAContainerRuntimeHookExecutable), "prestart"},
			Env:  []string{envPath},
		},
		When: When{
			Always:   &always,
			Commands: []string{".*"},
		},
	}
	return hook
}

// UpdateConfig updates the cri-o config to include the NVIDIA Container Runtime
func UpdateConfig(cfg engine.Interface, o *options) error {
	runtimes := operator.GetRuntimes(
		operator.WithNvidiaRuntimeName(o.runtimeClass),
		operator.WithSetAsDefault(o.setAsDefault),
		operator.WithRoot(o.runtimeDir),
	)
	for class, runtime := range runtimes {
		err := cfg.AddRuntime(class, runtime.Path, runtime.SetAsDefault)
		if err != nil {
			return fmt.Errorf("unable to update config for runtime class '%v': %v", class, err)
		}
	}

	return nil
}

// RevertConfig reverts the cri-o config to remove the NVIDIA Container Runtime
func RevertConfig(cfg engine.Interface, o *options) error {
	runtimes := operator.GetRuntimes(
		operator.WithNvidiaRuntimeName(o.runtimeClass),
		operator.WithSetAsDefault(o.setAsDefault),
		operator.WithRoot(o.runtimeDir),
	)
	for class := range runtimes {
		err := cfg.RemoveRuntime(class)
		if err != nil {
			return fmt.Errorf("unable to revert config for runtime class '%v': %v", class, err)
		}
	}
	return nil
}

// RestartCrio restarts crio depending on the value of restartModeFlag
func RestartCrio(o *options) error {
	switch o.restartMode {
	case restartModeNone:
		log.Warnf("Skipping restart of crio due to --restart-mode=%v", o.restartMode)
		return nil
	case restartModeSystemd:
		return RestartCrioSystemd(o.hostRootMount)
	default:
		return fmt.Errorf("invalid restart mode specified: %v", o.restartMode)
	}
}

// RestartCrioSystemd restarts cri-o using systemctl
func RestartCrioSystemd(hostRootMount string) error {
	log.Infof("Restarting cri-o using systemd and host root mounted at %v", hostRootMount)

	command := "chroot"
	args := []string{hostRootMount, "systemctl", "restart", "crio"}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error restarting crio using systemd: %v", err)
	}

	return nil
}
