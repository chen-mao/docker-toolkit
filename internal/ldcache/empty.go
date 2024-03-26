package ldcache

import "github.com/XDXCT/xdxct-container-toolkit/internal/logger"

type empty struct {
	logger logger.Interface
	path   string
}

var _ LDCache = (*empty)(nil)

// List always returns nil for an empty ldcache
func (e *empty) List() ([]string, []string) {
	return nil, nil
}

// Lookup logs a debug message and returns nil for an empty ldcache
func (e *empty) Lookup(prefixes ...string) ([]string, []string) {
	e.logger.Debugf("Calling Lookup(%v) on empty ldcache: %v", prefixes, e.path)
	return nil, nil
}
