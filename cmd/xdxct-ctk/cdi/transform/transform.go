package transform

import (
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/cdi/transform/root"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs a command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build creates the CLI command
func (m command) build() *cli.Command {
	c := cli.Command{
		Name:  "transform",
		Usage: "Apply a transform to a CDI specification",
	}

	c.Flags = []cli.Flag{}

	c.Subcommands = []*cli.Command{
		root.NewCommand(m.logger),
	}

	return &c
}
