package modules

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

//go:generate moq -stub -out cmd_mock.go . cmder
type cmder interface {
	// Run executes the command and returns the stdout, stderr, and an error if any
	Run(string, ...string) error
}

type cmderLogger struct {
	logger.Interface
}

func (c *cmderLogger) Run(cmd string, args ...string) error {
	c.Infof("Running: %v %v", cmd, strings.Join(args, " "))
	return nil
}

type cmderExec struct{}

func (c *cmderExec) Run(cmd string, args ...string) error {
	if output, err := exec.Command(cmd, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("%w; output=%v", err, string(output))
	}
	return nil
}
