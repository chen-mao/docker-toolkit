package constraints

import (
	"fmt"
	"strings"
)

// or represents an OR operation on a set of constraints
type or []Constraint

// and represents an AND (ALL) operation on a set of contraints
type and []Constraint

// AND constructs a new constraint that is the logical AND of the supplied constraints
func AND(constraints []Constraint) Constraint {
	if len(constraints) == 0 {
		return &always{}
	}
	if len(constraints) == 1 {
		return constraints[0]
	}
	return and(constraints)
}

// OR constructs a new constrant that is the logical OR of the supplied constraints
func OR(constraints []Constraint) Constraint {
	if len(constraints) == 0 {
		return nil
	}
	if len(constraints) == 1 {
		return constraints[0]
	}

	return or(constraints)
}

func (operands or) Assert() error {
	for _, o := range operands {
		// We stop on the first nil
		if err := o.Assert(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%v not met", operands)
}

func (operands or) String() string {
	var terms []string

	for _, o := range operands {
		terms = append(terms, o.String())
	}

	return strings.Join(terms, "||")
}

func (operands and) Assert() error {
	for _, o := range operands {
		// We stop on the first Assert
		if err := o.Assert(); err != nil {
			return err
		}
	}
	return nil
}

func (operands and) String() string {
	var terms []string

	for _, o := range operands {
		terms = append(terms, o.String())
	}

	return strings.Join(terms, "&&")
}
