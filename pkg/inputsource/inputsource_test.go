package inputsource

import (
	"slices"
	"testing"
)

// These are read-only smoke tests against the live Text Input Source APIs. Every
// macOS system exposes at least one keyboard input source, so they are
// deterministic there. Select is never called, as it would change the user's
// active input source.

func TestAllReturnsSortedNonEmptySources(t *testing.T) {
	all := All()

	if len(all) == 0 {
		t.Fatal("All() returned no input sources; macOS always has at least one")
	}

	if !slices.IsSorted(all) {
		t.Errorf("All() is not sorted: %v", all)
	}

	for _, id := range all {
		if id == "" {
			t.Errorf("All() contains an empty source ID: %v", all)
		}
	}
}

func TestCurrentAndName(t *testing.T) {
	current := Current()
	if current == "" {
		t.Fatal("Current() returned an empty source ID")
	}

	if Name(current) == "" {
		t.Errorf("Name(%q) returned an empty display name", current)
	}
}
