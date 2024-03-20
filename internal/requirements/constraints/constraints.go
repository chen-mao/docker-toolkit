package constraints

// Constraint represents a constraint that is to be evaluated
//
//go:generate moq -stub -out constraint_mock.go . Constraint
type Constraint interface {
	String() string
	Assert() error
}
