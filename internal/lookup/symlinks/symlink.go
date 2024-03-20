package symlinks

import (
	"fmt"
	"os"
)

// Resolve returns the link target of the specified filename or the filename if it is not a link.
func Resolve(filename string) (string, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return filename, fmt.Errorf("failed to get file info: %v", info)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return filename, nil
	}

	return os.Readlink(filename)
}
