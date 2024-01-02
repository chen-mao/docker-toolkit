/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
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

package configure

import (
	"fmt"
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine/containerd"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine/crio"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine/docker"
	"github.com/urfave/cli/v2"
)

const (
	defaultRuntime = "docker"

	// defaultXDXCTRuntimeName is the default name to use in configs for the NVIDIA Container Runtime
	defaultXDXCTRuntimeName = "xdxct"
	// defaultXDXCTRuntimeExecutable is the default NVIDIA Container Runtime executable file name
	defaultXDXCTRuntimeExecutable      = "xdxct-container-runtime"
	defailtXDXCTRuntimeExpecutablePath = "/usr/bin/xdxct-container-runtime"

	defaultContainerdConfigFilePath = "/etc/containerd/config.toml"
	defaultCrioConfigFilePath       = "/etc/crio/crio.conf"
	defaultDockerConfigFilePath     = "/etc/docker/daemon.json"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs an configure command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// config defines the options that can be set for the CLI through config files,
// environment variables, or command line config
type config struct {
	dryRun         bool
	runtime        string
	configFilePath string

	xdxctRuntime struct {
		name         string
		path         string
		setAsDefault bool
	}
}

func (m command) build() *cli.Command {
	// Create a config struct to hold the parsed environment variables or command line flags
	config := config{}

	// Create the 'configure' command
	configure := cli.Command{
		Name:  "configure",
		Usage: "Add a runtime to the specified container engine",
		Before: func(c *cli.Context) error {
			return validateFlags(c, &config)
		},
		Action: func(c *cli.Context) error {
			return m.configureWrapper(c, &config)
		},
	}

	configure.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "update the runtime configuration as required but don't write changes to disk",
			Destination: &config.dryRun,
		},
		&cli.StringFlag{
			Name:        "runtime",
			Usage:       "the target runtime engine; one of [containerd, crio, docker]",
			Value:       defaultRuntime,
			Destination: &config.runtime,
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "path to the config file for the target runtime",
			Destination: &config.configFilePath,
		},
		&cli.StringFlag{
			Name:        "xdxct-runtime-name",
			Usage:       "specify the name of the XDXCT runtime that will be added",
			Value:       defaultXDXCTRuntimeName,
			Destination: &config.xdxctRuntime.name,
		},
		&cli.StringFlag{
			Name:        "xdxct-runtime-path",
			Aliases:     []string{"runtime-path"},
			Usage:       "specify the path to the XDXCT runtime executable",
			Value:       defaultXDXCTRuntimeExecutable,
			Destination: &config.xdxctRuntime.path,
		},
		&cli.BoolFlag{
			Name:        "xdxct-set-as-default",
			Aliases:     []string{"set-as-default"},
			Usage:       "set the XDXCT runtime as the default runtime",
			Destination: &config.xdxctRuntime.setAsDefault,
		},
	}

	return &configure
}

func validateFlags(c *cli.Context, config *config) error {
	switch config.runtime {
	case "containerd", "crio", "docker":
		break
	default:
		return fmt.Errorf("unrecognized runtime '%v'", config.runtime)
	}

	switch config.runtime {
	case "containerd", "crio":
		if config.xdxctRuntime.path == defaultXDXCTRuntimeExecutable {
			config.xdxctRuntime.path = defailtXDXCTRuntimeExpecutablePath
		}
		if !filepath.IsAbs(config.xdxctRuntime.path) {
			return fmt.Errorf("the XDXCT runtime path %q is not an absolute path", config.xdxctRuntime.path)
		}
	}

	return nil
}

// configureWrapper updates the specified container engine config to enable the NVIDIA runtime
func (m command) configureWrapper(c *cli.Context, config *config) error {
	configFilePath := config.resolveConfigFilePath()

	var cfg engine.Interface
	var err error
	switch config.runtime {
	case "containerd":
		cfg, err = containerd.New(
			containerd.WithPath(configFilePath),
		)
	case "crio":
		cfg, err = crio.New(
			crio.WithPath(configFilePath),
		)
	case "docker":
		cfg, err = docker.New(
			docker.WithPath(configFilePath),
		)
	default:
		err = fmt.Errorf("unrecognized runtime '%v'", config.runtime)
	}
	if err != nil || cfg == nil {
		return fmt.Errorf("unable to load config for runtime %v: %v", config.runtime, err)
	}

	err = cfg.AddRuntime(
		config.xdxctRuntime.name,
		config.xdxctRuntime.path,
		config.xdxctRuntime.setAsDefault,
	)
	if err != nil {
		return fmt.Errorf("unable to update config: %v", err)
	}

	outputPath := config.getOuputConfigPath()
	n, err := cfg.Save(outputPath)
	if err != nil {
		return fmt.Errorf("unable to flush config: %v", err)
	}

	if n == 0 {
		m.logger.Infof("Removed empty config from %v", outputPath)
	} else {
		m.logger.Infof("Wrote updated config to %v", outputPath)
	}
	m.logger.Infof("It is recommended that %v daemon be restarted.", config.runtime)

	return nil
}

// resolveConfigFilePath returns the default config file path for the configured container engine
func (c *config) resolveConfigFilePath() string {
	if c.configFilePath != "" {
		return c.configFilePath
	}
	switch c.runtime {
	case "containerd":
		return defaultContainerdConfigFilePath
	case "crio":
		return defaultCrioConfigFilePath
	case "docker":
		return defaultDockerConfigFilePath
	}
	return ""
}

// getOuputConfigPath returns the configured config path or "" if dry-run is enabled
func (c *config) getOuputConfigPath() string {
	if c.dryRun {
		return ""
	}
	return c.resolveConfigFilePath()
}
