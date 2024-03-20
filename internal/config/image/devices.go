package image

import (
	"strings"
)

// VisibleDevices represents the devices selected in a container image
// through the XDXCT_VISIBLE_DEVICES or other environment variables
type VisibleDevices interface {
	List() []string
	Has(string) bool
}

var _ VisibleDevices = (*all)(nil)
var _ VisibleDevices = (*none)(nil)
var _ VisibleDevices = (*void)(nil)
var _ VisibleDevices = (*devices)(nil)

// NewVisibleDevices creates a VisibleDevices based on the value of the specified envvar.
func NewVisibleDevices(envvars ...string) VisibleDevices {
	for _, envvar := range envvars {
		if envvar == "all" {
			return all{}
		}
		if envvar == "none" {
			return none{}
		}
		if envvar == "" || envvar == "void" {
			return void{}
		}
	}

	return newDevices(envvars...)
}

type all struct{}

// List returns ["all"] for all devices
func (a all) List() []string {
	return []string{"all"}
}

// Has for all devices is true for any id except the empty ID
func (a all) Has(id string) bool {
	return id != ""
}

type none struct{}

// List returns [""] for the none devices
func (n none) List() []string {
	return []string{""}
}

// Has for none devices is false for any id
func (n none) Has(id string) bool {
	return false
}

type void struct {
	none
}

// List returns nil for the void devices
func (v void) List() []string {
	return nil
}

type devices struct {
	len    int
	lookup map[string]int
}

func newDevices(idOrCommaSeparated ...string) devices {
	lookup := make(map[string]int)

	i := 0
	for _, commaSeparated := range idOrCommaSeparated {
		for _, id := range strings.Split(commaSeparated, ",") {
			lookup[id] = i
			i++
		}
	}

	d := devices{
		len:    i,
		lookup: lookup,
	}
	return d
}

// List returns the list of requested devices
func (d devices) List() []string {
	list := make([]string, d.len)

	for id, i := range d.lookup {
		list[i] = id
	}

	return list
}

// Has checks whether the specified ID is in the set of requested devices
func (d devices) Has(id string) bool {
	if id == "" {
		return false
	}

	_, exist := d.lookup[id]
	return exist
}
