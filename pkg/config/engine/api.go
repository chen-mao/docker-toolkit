package engine

// Interface defines the API for a runtime config updater.
type Interface interface {
	DefaultRuntime() string
	AddRuntime(string, string, bool) error
	Set(string, interface{})
	RemoveRuntime(string) error
	Save(string) (int64, error)
}
