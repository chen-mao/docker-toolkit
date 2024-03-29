package list

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/urfave/cli/v2"
	"tags.cncf.io/container-device-interface/pkg/cdi"
)

type command struct {
	logger logger.Interface
}

type config struct{}

// NewCommand constructs a cdi list command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build creates the CLI command
func (m command) build() *cli.Command {
	cfg := config{}

	// Create the command
	c := cli.Command{
		Name:  "list",
		Usage: "List the available CDI devices",
		Before: func(c *cli.Context) error {
			return m.validateFlags(c, &cfg)
		},
		Action: func(c *cli.Context) error {
			return m.run(c, &cfg)
		},
	}

	c.Flags = []cli.Flag{}

	return &c
}

func (m command) validateFlags(c *cli.Context, cfg *config) error {
	return nil
}

func (m command) run(c *cli.Context, cfg *config) error {
	registry, err := cdi.NewCache(
		cdi.WithAutoRefresh(false),
		cdi.WithSpecDirs(cdi.DefaultSpecDirs...),
	)
	if err != nil {
		return fmt.Errorf("failed to create CDI cache: %v", err)
	}

	devices := registry.ListDevices()
	if len(devices) == 0 {
		m.logger.Info("No CDI devices found")
		return nil
	}

	m.logger.Infof("Found %d CDI devices", len(devices))
	for _, device := range devices {
		fmt.Printf("%s\n", device)
	}

	return nil
}
