package inputsource

import "testing"

// Read-only smoke tests over the adapter, which delegates to the live macOS Text
// Input Source APIs. Select is exercised by re-selecting the already-active
// source: a no-op that leaves the user's selection unchanged.
func TestControllerReadsSources(t *testing.T) {
	controller := New()

	if len(controller.All()) == 0 {
		t.Fatal("All() returned no sources")
	}

	current := controller.Current()
	if current == "" {
		t.Fatal("Current() returned an empty source ID")
	}

	if controller.Name(current) == "" {
		t.Errorf("Name(%q) returned an empty display name", current)
	}

	// Re-selecting the current source is a no-op but exercises Select.
	controller.Select(current)
}
