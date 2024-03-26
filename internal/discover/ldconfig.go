

package discover

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

// NewLDCacheUpdateHook creates a discoverer that updates the ldcache for the specified mounts. A logger can also be specified
func NewLDCacheUpdateHook(logger logger.Interface, mounts Discover, xdxctCTKPath string) (Discover, error) {
	d := ldconfig{
		logger:        logger,
		xdxctCTKPath: xdxctCTKPath,
		mountsFrom:    mounts,
	}

	return &d, nil
}

type ldconfig struct {
	None
	logger        logger.Interface
	xdxctCTKPath string
	mountsFrom    Discover
}

// Hooks checks the required mounts for libraries and returns a hook to update the LDcache for the discovered paths.
func (d ldconfig) Hooks() ([]Hook, error) {
	mounts, err := d.mountsFrom.Mounts()
	if err != nil {
		return nil, fmt.Errorf("failed to discover mounts for ldcache update: %v", err)
	}
	h := CreateLDCacheUpdateHook(
		d.xdxctCTKPath,
		getLibraryPaths(mounts),
	)
	return []Hook{h}, nil
}

// CreateLDCacheUpdateHook locates the XDXCT Container Toolkit CLI and creates a hook for updating the LD Cache
func CreateLDCacheUpdateHook(executable string, libraries []string) Hook {
	var args []string
	for _, f := range uniqueFolders(libraries) {
		args = append(args, "--folder", f)
	}

	hook := CreateXdxctCTKHook(
		executable,
		"update-ldcache",
		args...,
	)

	return hook

}

// getLibraryPaths extracts the library dirs from the specified mounts
func getLibraryPaths(mounts []Mount) []string {
	var paths []string
	for _, m := range mounts {
		if !isLibName(m.Path) {
			continue
		}
		paths = append(paths, m.Path)
	}
	return paths
}

// isLibName checks if the specified filename is a library (i.e. ends in `.so*`)
func isLibName(filename string) bool {

	base := filepath.Base(filename)

	isLib, err := filepath.Match("lib?*.so*", base)
	if !isLib || err != nil {
		return false
	}

	parts := strings.Split(base, ".so")
	if len(parts) == 1 {
		return true
	}

	return parts[len(parts)-1] == "" || strings.HasPrefix(parts[len(parts)-1], ".")
}

// uniqueFolders returns the unique set of folders for the specified files
func uniqueFolders(libraries []string) []string {
	var paths []string
	checked := make(map[string]bool)

	for _, l := range libraries {
		dir := filepath.Dir(l)
		if dir == "" {
			continue
		}
		if checked[dir] {
			continue
		}
		checked[dir] = true
		paths = append(paths, dir)
	}
	return paths
}
