package image

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	// "golang.org/x/mod/semver"
	"tags.cncf.io/container-device-interface/pkg/parser"
)

const (
	envGPUVersion            = "GPU_VERSION"
	envXDXRequirePrefix      = "XDXCT_REQUIRE_"
	envXDXRequireGPU         = envXDXRequirePrefix + "GPU"
	envXDXDisableRequire     = "XDXCT_DISABLE_REQUIRE"
	envXDXDriverCapabilities = "XDXCT_DRIVER_CAPABILITIES"
)

// GPU represents a GPU image that can be used for GPU computing. This wraps
// a map of environment variable to values that can be used to perform lookups
// such as requirements.
type GPU struct {
	env    map[string]string
	mounts []specs.Mount
}

// NewGPUImageFromSpec creates a GPU image from the input OCI runtime spec.
// The process environment is read (if present) to construc the GPU Image.
func NewGPUImageFromSpec(spec *specs.Spec) (GPU, error) {
	var env []string
	if spec != nil && spec.Process != nil {
		env = spec.Process.Env
	}

	return New(
		WithEnv(env),
		WithMounts(spec.Mounts),
	)
}

// NewGPUImageFromEnv creates a GPU image from the input environment. The environment
// is a list of strings of the form ENVAR=VALUE.
func NewGPUImageFromEnv(env []string) (GPU, error) {
	return New(WithEnv(env))
}

// Getenv returns the value of the specified environment variable.
// If the environment variable is not specified, an empty string is returned.
func (i GPU) Getenv(key string) string {
	return i.env[key]
}

// HasEnvvar checks whether the specified envvar is defined in the image.
func (i GPU) HasEnvvar(key string) bool {
	_, exists := i.env[key]
	return exists
}

// IsLegacy returns whether the associated GPU image is a "legacy" image. An
// image is considered legacy if it has a GPU_VERSION environment variable defined
// and no XDXCT_REQUIRE_GPU environment variable defined.
func (i GPU) IsLegacy() bool {
	legacyGpuVersion := i.env[envGPUVersion]
	gpuRequire := i.env[envXDXRequireGPU]
	return len(legacyGpuVersion) > 0 && len(gpuRequire) == 0
}

// GetRequirements returns the requirements from all XDXCT_REQUIRE_ environment
// variables.
func (i GPU) GetRequirements() ([]string, error) {
	if i.HasDisableRequire() {
		return nil, nil
	}

	// All variables with the "XDXCT_REQUIRE_" prefix are passed to xdxct-container-cli
	var requirements []string
	for name, value := range i.env {
		if strings.HasPrefix(name, envXDXRequirePrefix) {
			requirements = append(requirements, value)
		}
	}
	return requirements, nil
}

// HasDisableRequire checks for the value of the XDXCT_DISABLE_REQUIRE. If set
// to a valid (true) boolean value this can be used to disable the requirement checks
func (i GPU) HasDisableRequire() bool {
	if disable, exists := i.env[envXDXDisableRequire]; exists {
		// i.logger.Debugf("XDXCT_DISABLE_REQUIRE=%v; skipping requirement checks", disable)
		d, _ := strconv.ParseBool(disable)
		return d
	}

	return false
}

// DevicesFromEnvvars returns the devices requested by the image through environment variables
func (i GPU) DevicesFromEnvvars(envVars ...string) VisibleDevices {
	// We concantenate all the devices from the specified env.
	var isSet bool
	var devices []string
	requested := make(map[string]bool)
	for _, envVar := range envVars {
		if devs, ok := i.env[envVar]; ok {
			isSet = true
			for _, d := range strings.Split(devs, ",") {
				trimmed := strings.TrimSpace(d)
				if len(trimmed) == 0 {
					continue
				}
				devices = append(devices, trimmed)
				requested[trimmed] = true
			}
		}
	}

	// Environment variable unset with legacy image: default to "all".
	if !isSet && len(devices) == 0 && i.IsLegacy() {
		return NewVisibleDevices("all")
	}

	// Environment variable unset or empty or "void": return nil
	if len(devices) == 0 || requested["void"] {
		return NewVisibleDevices("void")
	}

	return NewVisibleDevices(devices...)
}

// GetDriverCapabilities returns the requested driver capabilities.
func (i GPU) GetDriverCapabilities() DriverCapabilities {
	env := i.env[envXDXDriverCapabilities]

	capabilities := make(DriverCapabilities)
	for _, c := range strings.Split(env, ",") {
		capabilities[DriverCapability(c)] = true
	}

	return capabilities
}

// OnlyFullyQualifiedCDIDevices returns true if all devices requested in the image are requested as CDI devices/
func (i GPU) OnlyFullyQualifiedCDIDevices() bool {
	var hasCDIdevice bool
	for _, device := range i.DevicesFromEnvvars("XDXCT_VISIBLE_DEVICES").List() {
		if !parser.IsQualifiedName(device) {
			return false
		}
		hasCDIdevice = true
	}

	for _, device := range i.DevicesFromMounts() {
		if !strings.HasPrefix(device, "cdi/") {
			return false
		}
		hasCDIdevice = true
	}
	return hasCDIdevice
}

const (
	deviceListAsVolumeMountsRoot = "/var/run/xdxct-container-devices"
)

// DevicesFromMounts returns a list of device specified as mounts.
// TODO: This should be merged with getDevicesFromMounts used in the XDXCT Container Runtime
func (i GPU) DevicesFromMounts() []string {
	root := filepath.Clean(deviceListAsVolumeMountsRoot)
	seen := make(map[string]bool)
	var devices []string
	for _, m := range i.mounts {
		source := filepath.Clean(m.Source)
		// Only consider mounts who's host volume is /dev/null
		if source != "/dev/null" {
			continue
		}

		destination := filepath.Clean(m.Destination)
		if seen[destination] {
			continue
		}
		seen[destination] = true

		// Only consider container mount points that begin with 'root'
		if !strings.HasPrefix(destination, root) {
			continue
		}

		// Grab the full path beyond 'root' and add it to the list of devices
		device := strings.Trim(strings.TrimPrefix(destination, root), "/")
		if len(device) == 0 {
			continue
		}
		devices = append(devices, device)
	}
	return devices
}

// CDIDevicesFromMounts returns a list of CDI devices specified as mounts on the image.
func (i GPU) CDIDevicesFromMounts() []string {
	var devices []string
	for _, mountDevice := range i.DevicesFromMounts() {
		if !strings.HasPrefix(mountDevice, "cdi/") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(mountDevice, "cdi/"), "/", 3)
		if len(parts) != 3 {
			continue
		}
		vendor := parts[0]
		class := parts[1]
		device := parts[2]
		devices = append(devices, fmt.Sprintf("%s/%s=%s", vendor, class, device))
	}
	return devices
}
