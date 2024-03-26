package info

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs an info command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build
func (m command) build() *cli.Command {
	// Create the 'info' command
	info := cli.Command{
		Name:  "info",
		Usage: "Provide information about the system",
	}

	info.Subcommands = []*cli.Command{}

	return &info
}
