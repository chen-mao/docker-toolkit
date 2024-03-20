package spec

import (
	"io"

	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	// DetectMinimumVersion is a constant that triggers a spec to detect the minimum required version.
	DetectMinimumVersion = "DETECT_MINIMUM_VERSION"

	// FormatJSON indicates a JSON output format
	FormatJSON = "json"
	// FormatYAML indicates a YAML output format
	FormatYAML = "yaml"
)

// Interface is the interface for the spec API
type Interface interface {
	io.WriterTo
	Save(string) error
	Raw() *specs.Spec
}
