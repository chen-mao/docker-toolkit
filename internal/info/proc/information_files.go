package proc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GPUInfoField represents the field name for information specified in a GPU's information file
type GPUInfoField string

// The following constants define the fields of interest from the GPU information file
const (
	GPUInfoModel       = GPUInfoField("Model")
	GPUInfoGPUUUID     = GPUInfoField("GPU UUID")
	GPUInfoBusLocation = GPUInfoField("Bus Location")
	GPUInfoDeviceMinor = GPUInfoField("Device Minor")
)

// GPUInfo stores the information for a GPU as determined from its associated information file
type GPUInfo map[GPUInfoField]string

// GetInformationFilePaths returns the list of information files associated with XDXCT GPUs.
func GetInformationFilePaths(root string) ([]string, error) {
	return filepath.Glob(filepath.Join(root, "/proc/driver/xdxct/gpus/*/information"))
}

// ParseGPUInformationFile parses the specified GPU information file and constructs a GPUInfo structure
func ParseGPUInformationFile(path string) (GPUInfo, error) {
	infoFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v: %v", path, err)
	}
	defer infoFile.Close()

	return gpuInfoFrom(infoFile), nil
}

// gpuInfoFrom parses a GPUInfo struct from the specified reader
func gpuInfoFrom(reader io.Reader) GPUInfo {
	info := make(GPUInfo)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := GPUInfoField(parts[0])
		value := strings.TrimSpace(parts[1])

		info[field] = value
	}

	return info
}
