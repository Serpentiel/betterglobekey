// Package feedback decorates an input-source controller so that a system
// notification is posted whenever the input source changes.
package feedback

import (
	"fmt"
	"os/exec"
)

// Source is the input-source controller the feedback wrapper decorates.
type Source interface {
	Current() string
	All() []string
	Select(id string)
	Name(id string) string
}

// Notifying wraps a Source, posting a notification after each Select.
type Notifying struct {
	inner Source
}

// Wrap returns a Source that posts a notification on every Select.
func Wrap(inner Source) *Notifying {
	return &Notifying{inner: inner}
}

// Current returns the ID of the currently active input source.
func (n *Notifying) Current() string {
	return n.inner.Current()
}

// All returns the IDs of every selectable input source.
func (n *Notifying) All() []string {
	return n.inner.All()
}

// Name returns the localized display name for the input source with the given ID.
func (n *Notifying) Name(id string) string {
	return n.inner.Name(id)
}

// Select activates the input source and posts a notification with its name.
func (n *Notifying) Select(id string) {
	n.inner.Select(id)

	label := n.inner.Name(id)
	if label == "" {
		label = id
	}

	notify(label)
}

// notify posts a macOS notification without blocking the caller.
func notify(text string) {
	script := fmt.Sprintf("display notification %q with title %q", text, "betterglobekey")

	go func() {
		// Best-effort: a failure (e.g. osascript unavailable) is not actionable.
		//nolint:gosec // fixed command; text is a system-provided input-source name
		_ = exec.Command("osascript", "-e", script).Run()
	}()
}
