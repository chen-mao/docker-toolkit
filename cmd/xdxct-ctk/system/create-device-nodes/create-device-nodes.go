package createdevicenodes

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/system/modules"
	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

type options struct {
	driverRoot string

	dryRun bool

	control bool

	loadKernelModules bool
}

// NewCommand constructs a command sub-command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build
func (m command) build() *cli.Command {
	opts := options{}

	c := cli.Command{
		Name:  "create-device-nodes",
		Usage: "A utility to create XDXCT device nodes",
		Before: func(c *cli.Context) error {
			return m.validateFlags(c, &opts)
		},
		Action: func(c *cli.Context) error {
			return m.run(c, &opts)
		},
	}

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "driver-root",
			Usage:       "the path to the driver root. Device nodes will be created at `DRIVER_ROOT`/dev",
			Value:       "/",
			Destination: &opts.driverRoot,
			EnvVars:     []string{"DRIVER_ROOT"},
		},
		&cli.BoolFlag{
			Name:        "control-devices",
			Usage:       "create all control device nodes.",
			Destination: &opts.control,
		},
		&cli.BoolFlag{
			Name:        "load-kernel-modules",
			Usage:       "load the XDXCT Kernel Modules before creating devices nodes",
			Destination: &opts.loadKernelModules,
		},
		&cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "if set, the command will not create any symlinks.",
			Value:       false,
			Destination: &opts.dryRun,
			EnvVars:     []string{"DRY_RUN"},
		},
	}

	return &c
}

func (m command) validateFlags(r *cli.Context, opts *options) error {
	return nil
}

func (m command) run(c *cli.Context, opts *options) error {
	if opts.loadKernelModules {
		modules := modules.New(
			modules.WithLogger(m.logger),
			modules.WithDryRun(opts.dryRun),
			modules.WithRoot(opts.driverRoot),
		)
		if err := modules.LoadAll(); err != nil {
			return fmt.Errorf("failed to load XDXCT kernel modules: %v", err)
		}
	}

	return nil
}
