package discover

import (
	"path/filepath"

	"tags.cncf.io/container-device-interface/pkg/cdi"
)

var _ Discover = (*Hook)(nil)

// Devices returns an empty list of devices for a Hook discoverer.
func (h Hook) Devices() ([]Device, error) {
	return nil, nil
}

// Mounts returns an empty list of mounts for a Hook discoverer.
func (h Hook) Mounts() ([]Mount, error) {
	return nil, nil
}

// Hooks allows the Hook type to also implement the Discoverer interface.
// It returns a single hook
func (h Hook) Hooks() ([]Hook, error) {
	return []Hook{h}, nil
}

// CreateCreateSymlinkHook creates a hook which creates a symlink from link -> target.
func CreateCreateSymlinkHook(xdxctCTKPath string, links []string) Discover {
	if len(links) == 0 {
		return None{}
	}

	var args []string
	for _, link := range links {
		args = append(args, "--link", link)
	}
	return CreateXdxctCTKHook(
		xdxctCTKPath,
		"create-symlinks",
		args...,
	)
}

func CreateXdxctCTKHook(xdxctCTKPath string, hookName string, additionalArgs ...string) Hook {
	return Hook{
		Lifecycle: cdi.CreateContainerHook,
		Path:      xdxctCTKPath,
		Args:      append([]string{filepath.Base(xdxctCTKPath), "hook", hookName}, additionalArgs...),
	}
}
