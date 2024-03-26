package hook

import (
	chmod "github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/hook/chmod"
	
	ldcache "github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/hook/update-ldcache"
	symlinks "github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/hook/create-symlinks"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/urfave/cli/v2"
)

type hookCommand struct {
	logger logger.Interface
}

// NewCommand constructs a hook command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := hookCommand{
		logger: logger,
	}
	return c.build()
}

// build
func (m hookCommand) build() *cli.Command {
	// Create the 'hook' command
	hook := cli.Command{
		Name:  "hook",
		Usage: "A collection of hooks that may be injected into an OCI spec",
	}

	hook.Subcommands = []*cli.Command{
		ldcache.NewCommand(m.logger),
		symlinks.NewCommand(m.logger),
		chmod.NewCommand(m.logger),
	}

	return &hook
}
