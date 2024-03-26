package root

import (
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
)

// Driver represents a filesystem in which a set of drivers or devices is defined.
type Driver struct {
	logger logger.Interface
	// Root represents the root from the perspective of the driver libraries and binaries.
	Root string
	// librarySearchPaths specifies explicit search paths for discovering libraries.
	librarySearchPaths []string
}

// New creates a new Driver root at the specified path.
// TODO: Use functional options here.
func New(logger logger.Interface, path string, librarySearchPaths []string) *Driver {
	return &Driver{
		logger:             logger,
		Root:               path,
		librarySearchPaths: normalizeSearchPaths(librarySearchPaths...),
	}
}

// Drivers returns a Locator for driver libraries.
func (r *Driver) Libraries() lookup.Locator {
	return lookup.NewLibraryLocator(
		lookup.WithLogger(r.logger),
		lookup.WithRoot(r.Root),
		lookup.WithSearchPaths(r.librarySearchPaths...),
	)
}

// normalizeSearchPaths takes a list of paths and normalized these.
// Each of the elements in the list is expanded if it is a path list and the
// resultant list is returned.
// This allows, for example, for the contents of `PATH` or `LD_LIBRARY_PATH` to
// be passed as a search path directly.
func normalizeSearchPaths(paths ...string) []string {
	var normalized []string
	for _, path := range paths {
		normalized = append(normalized, filepath.SplitList(path)...)
	}
	return normalized
}
