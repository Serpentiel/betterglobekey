// Package inputsource adapts the public pkg/inputsource library to the domain's
// input-source port.
package inputsource

import source "github.com/Serpentiel/betterglobekey/pkg/inputsource"

// Controller controls the system's input sources via pkg/inputsource.
type Controller struct{}

// New returns a Controller.
func New() Controller {
	return Controller{}
}

// Current returns the ID of the currently active input source.
func (Controller) Current() string {
	return source.Current()
}

// All returns the IDs of every selectable input source.
func (Controller) All() []string {
	return source.All()
}

// Select activates the input source with the given ID.
func (Controller) Select(id string) {
	source.Select(id)
}
