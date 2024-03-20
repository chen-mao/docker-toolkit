package cdi

import (
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/cdi/generate"
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/cdi/list"
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/cdi/transform"
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
	// Create the 'hook' command
	hook := cli.Command{
		Name:  "cdi",
		Usage: "Provide tools for interacting with Container Device Interface specifications",
	}

	hook.Subcommands = []*cli.Command{
		generate.NewCommand(m.logger),
		transform.NewCommand(m.logger),
		list.NewCommand(m.logger),
	}

	return &hook
}
