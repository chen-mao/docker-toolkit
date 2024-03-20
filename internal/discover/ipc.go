

package discover

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
)

type ipcMounts mounts

// NOT NEED
// NewIPCDiscoverer creats a discoverer for xdxct IPC sockets.
func NewIPCDiscoverer(logger logger.Interface, driverRoot string) (Discover, error) {
	sockets := newMounts(
		logger,
		lookup.NewFileLocator(
			lookup.WithLogger(logger),
			lookup.WithRoot(driverRoot),
			lookup.WithSearchPaths("/run", "/var/run"),
			lookup.WithCount(1),
		),
		driverRoot,
		[]string{
			"/xdxct-persistenced/socket",
			"/xdxct-fabricmanager/socket",
		},
	)

	mps := newMounts(
		logger,
		lookup.NewFileLocator(
			lookup.WithLogger(logger),
			lookup.WithRoot(driverRoot),
			lookup.WithCount(1),
		),
		driverRoot,
		[]string{
			"/tmp/xdxct-mps",
		},
	)

	d := Merge(
		(*ipcMounts)(sockets),
		(*ipcMounts)(mps),
	)
	return d, nil
}

// Mounts returns the discovered mounts with "noexec" added to the mount options.
func (d *ipcMounts) Mounts() ([]Mount, error) {
	mounts, err := (*mounts)(d).Mounts()
	if err != nil {
		return nil, err
	}

	var modifiedMounts []Mount
	for _, m := range mounts {
		mount := m
		mount.Options = append(m.Options, "noexec")
		modifiedMounts = append(modifiedMounts, mount)
	}

	return modifiedMounts, nil
}
