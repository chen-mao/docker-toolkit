package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/cdi"
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/config"
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/hook"
	infoCLI "github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/info"
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/runtime"
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/system"
	"github.com/XDXCT/xdxct-container-toolkit/internal/info"

	cli "github.com/urfave/cli/v2"
)

// options defines the options that can be set for the CLI through config files,
// environment variables, or command line flags
type options struct {
	// Debug indicates whether the CLI is started in "debug" mode
	Debug bool
	// Quiet indicates whether the CLI is started in "quiet" mode
	Quiet bool
}

func main() {
	// Create a options struct to hold the parsed environment variables or command line flags
	logger := log.New()
	opts := options{}

	// Create the top-level CLI
	c := cli.NewApp()
	c.Name = "XDXCT Container Toolkit CLI"
	c.UseShortOptionHandling = true
	c.EnableBashCompletion = true
	c.Usage = "Tools to configure the XDXCT Container Toolkit"
	c.Version = info.GetVersionString()

	// Setup the flags for this command
	c.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			Usage:       "Enable debug-level logging",
			Destination: &opts.Debug,
			EnvVars:     []string{"XDXCT_CTK_DEBUG"},
		},
		&cli.BoolFlag{
			Name:        "quiet",
			Usage:       "Suppress all output except for errors; overrides --debug",
			Destination: &opts.Quiet,
			EnvVars:     []string{"XDXCT_CTK_QUIET"},
		},
	}

	// Set log-level for all subcommands
	c.Before = func(c *cli.Context) error {
		logLevel := log.InfoLevel
		if opts.Debug {
			logLevel = log.DebugLevel
		}
		logger.SetLevel(logLevel)
		return nil
	}

	// Define the subcommands
	c.Commands = []*cli.Command{
		hook.NewCommand(logger),
		runtime.NewCommand(logger),
		infoCLI.NewCommand(logger),
		cdi.NewCommand(logger),
		system.NewCommand(logger),
		config.NewCommand(logger),
	}

	// Run the CLI
	err := c.Run(os.Args)
	if err != nil {
		logger.Errorf("%v", err)
		logger.Exit(1)
	}
}
