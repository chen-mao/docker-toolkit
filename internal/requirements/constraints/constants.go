package constraints

import "fmt"

const (
	equal        = "="
	notEqual     = "!="
	less         = "<"
	lessEqual    = "<="
	greater      = ">"
	greaterEqual = ">="
)

// always is a constraint that is always met
type always struct{}

func (c always) Assert() error {
	return nil
}

func (c always) String() string {
	return "true"
}

// invalid is an invalid constraint and can never be met
type invalid string

func (c invalid) Assert() error {
	return fmt.Errorf("invalid constraint: %v", c.String())
}

// String returns the string representation of the contraint
func (c invalid) String() string {
	return string(c)
}
