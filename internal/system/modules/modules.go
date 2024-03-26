package modules

import (
	"fmt"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

// Interface provides a set of utilities for interacting with XDXCT modules on the system.
type Interface struct {
	logger logger.Interface

	dryRun bool
	root   string

	cmder
}

// New constructs a new Interface struct with the specified options.
func New(opts ...Option) *Interface {
	m := &Interface{}
	for _, opt := range opts {
		opt(m)
	}

	if m.logger == nil {
		m.logger = logger.New()
	}
	if m.root == "" {
		m.root = "/"
	}

	if m.dryRun {
		m.cmder = &cmderLogger{m.logger}
	} else {
		m.cmder = &cmderExec{}
	}
	return m
}

// LoadAll loads all the XDXCT kernel modules.
func (m *Interface) LoadAll() error {
	modules := []string{"xdxct"}

	for _, module := range modules {
		if err := m.Load(module); err != nil {
			return fmt.Errorf("failed to load module %s: %w", module, err)
		}
	}
	return nil
}

var errInvalidModule = fmt.Errorf("invalid module")

// Load loads the specified XDXCT kernel module.
// If the root is specified we first chroot into this root.
func (m *Interface) Load(module string) error {
	if !strings.HasPrefix(module, "xdxct") {
		return errInvalidModule
	}

	var args []string
	if m.root != "/" {
		args = append(args, "chroot", m.root)
	}
	args = append(args, "/sbin/modprobe", module)

	m.logger.Debugf("Loading kernel module %s: %v", module, args)
	err := m.Run(args[0], args[1:]...)
	if err != nil {
		m.logger.Debugf("Failed to load kernel module %s: %v", module, err)
		return err
	}

	return nil
}
