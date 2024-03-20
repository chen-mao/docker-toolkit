package cuda

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
)

type cudaLocator struct {
	lookup.Locator
}

func New(libraries lookup.Locator) lookup.Locator {
	c := cudaLocator{
		Locator: libraries,
	}
	return &c
}

// Locate returns the path to the libcuda.so.RMVERSION file.
// libcuda.so is prefixed to the specified pattern.
func (l *cudaLocator) Locate(pattern string) ([]string, error) {
	return l.Locator.Locate("libcuda.so" + pattern)
}
