package test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetModuleRoot returns the path to the root of the go module
func GetModuleRoot() (string, error) {
	_, filename, _, _ := runtime.Caller(0)

	return hasGoMod(filename)
}

// PrependToPath prefixes the specified additional paths to the PATH environment variable
func PrependToPath(additionalPaths ...string) string {
	paths := strings.Split(os.Getenv("PATH"), ":")
	paths = append(additionalPaths, paths...)

	return strings.Join(paths, ":")
}

func hasGoMod(dir string) (string, error) {
	if dir == "" || dir == "/" {
		return "", fmt.Errorf("module root not found")
	}

	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	if err != nil {
		return hasGoMod(filepath.Dir(dir))
	}
	return dir, nil
}
