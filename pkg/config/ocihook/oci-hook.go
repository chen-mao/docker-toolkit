package ocihook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateHook creates an OCI hook file for the specified XDXCT Container Runtime hook path
func CreateHook(hookFilePath string, xdxctContainerRuntimeHookExecutablePath string) error {
	var output io.Writer
	if hookFilePath == "" {
		output = os.Stdout
	} else {
		if hooksDir := filepath.Dir(hookFilePath); hooksDir != "" {
			err := os.MkdirAll(hooksDir, 0755)
			if err != nil {
				return fmt.Errorf("error creating hooks directory %v: %v", hooksDir, err)
			}
		}

		hookFile, err := os.Create(hookFilePath)
		if err != nil {
			return fmt.Errorf("error creating hook file '%v': %v", hookFilePath, err)
		}
		defer hookFile.Close()
		output = hookFile
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(generateOciHook(xdxctContainerRuntimeHookExecutablePath)); err != nil {
		return fmt.Errorf("error writing hook file: %v", err)
	}
	return nil
}

func generateOciHook(executablePath string) podmanHook {
	pathParts := []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"}

	dir := filepath.Dir(executablePath)
	var found bool
	for _, pathPart := range pathParts {
		if pathPart == dir {
			found = true
			break
		}
	}
	if !found {
		pathParts = append(pathParts, dir)
	}

	envPath := "PATH=" + strings.Join(pathParts, ":")
	always := true

	hook := podmanHook{
		Version: "1.0.0",
		Stages:  []string{"prestart"},
		Hook: specHook{
			Path: executablePath,
			Args: []string{filepath.Base(executablePath), "prestart"},
			Env:  []string{envPath},
		},
		When: When{
			Always:   &always,
			Commands: []string{".*"},
		},
	}
	return hook
}
