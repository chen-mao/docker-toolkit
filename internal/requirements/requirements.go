package requirements

import (
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/requirements/constraints"
)

// Requirements represents a collection of requirements that can be compared to properties
type Requirements struct {
	logger       logger.Interface
	requirements []string
	properties   map[string]constraints.Property
}

// New creates a new set of requirements
func New(logger logger.Interface, requirements []string) *Requirements {
	r := Requirements{
		logger:       logger,
		requirements: requirements,
		properties: map[string]constraints.Property{
			// Set up the supported properties. These are overridden with actual values.
			CUDA:   constraints.NewVersionProperty(CUDA, ""),
			ARCH:   constraints.NewVersionProperty(ARCH, ""),
			DRIVER: constraints.NewVersionProperty(DRIVER, ""),
			BRAND:  constraints.NewStringProperty(BRAND, ""),
		},
	}

	return &r
}

// AddVersionProperty adds the specified property (name, value pair) to the requirements
func (r *Requirements) AddVersionProperty(name string, value string) {
	r.properties[name] = constraints.NewVersionProperty(name, value)
}

// AddStringProperty adds the specified property (name, value pair) to the requirements
func (r *Requirements) AddStringProperty(name string, value string) {
	r.properties[name] = constraints.NewStringProperty(name, value)
}

// Assert checks the specified requirements
func (r Requirements) Assert() error {
	if len(r.requirements) == 0 {
		return nil
	}

	r.logger.Debugf("Checking properties %+v against requirements %v", r.properties, r.requirements)
	c, err := constraints.New(r.logger, r.requirements, r.properties)
	if err != nil {
		return err
	}
	return c.Assert()
}
