package runtime

import (
	"github.com/XDXCT/xdxct-container-toolkit/cmd/xdxct-ctk/runtime/configure"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/urfave/cli/v2"
)

type runtimeCommand struct {
	logger logger.Interface
}

// NewCommand constructs a runtime command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := runtimeCommand{
		logger: logger,
	}
	return c.build()
}

func (m runtimeCommand) build() *cli.Command {
	// Create the 'runtime' command
	runtime := cli.Command{
		Name:  "runtime",
		Usage: "A collection of runtime-related utilities for the XDXCT Container Toolkit",
	}

	runtime.Subcommands = []*cli.Command{
		configure.NewCommand(m.logger),
	}

	return &runtime
}
