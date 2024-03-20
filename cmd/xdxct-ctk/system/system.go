package system

import (
	devchar "github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/system/create-dev-char-symlinks"
	devicenodes "github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/system/create-device-nodes"
	ldcache "github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/system/print-ldcache"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs a runtime command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	// Create the 'system' command
	system := cli.Command{
		Name:  "system",
		Usage: "A collection of system-related utilities for the XDXCT Container Toolkit",
	}

	system.Subcommands = []*cli.Command{
		devchar.NewCommand(m.logger),
		devicenodes.NewCommand(m.logger),
		ldcache.NewCommand(m.logger),
	}

	return &system
}
