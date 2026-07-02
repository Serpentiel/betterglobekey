package switcher

import (
	"slices"
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

type fakeNotifier struct {
	sources     []string
	collections []string
}

func (n *fakeNotifier) Notify(source, collection string) {
	n.sources = append(n.sources, source)
	n.collections = append(n.collections, collection)
}

func TestStep(t *testing.T) {
	src := []string{"a", "b", "c"}

	cases := []struct {
		current string
		forward bool
		want    string
	}{
		{"a", true, "b"},
		{"c", true, "a"},  // wraps forward
		{"b", false, "a"}, // backward
		{"a", false, "c"}, // wraps backward
		{"x", true, "a"},  // absent -> first
		{"x", false, "a"}, // absent -> first
	}

	for _, tc := range cases {
		if got := step(src, tc.current, tc.forward); got != tc.want {
			t.Errorf("step(%v, %q, forward=%v) = %q, want %q", src, tc.current, tc.forward, got, tc.want)
		}
	}

	if got := step(nil, "a", true); got != "" {
		t.Errorf("step(nil, ...) = %q, want empty", got)
	}
}

func TestNotifierReceivesSourceAndCollection(t *testing.T) {
	sources := &fakeSources{current: "s1", all: []string{"s1", "s2"}}
	clock := &fakeClock{now: time.Unix(1000, 0)}
	notifier := &fakeNotifier{}
	cfg := config.Config{
		DoublePress: config.DoublePress{Enabled: true, MaxDelay: testDelay},
		Reverse:     config.Reverse{Enabled: true, Modifier: "shift"},
		Collections: []config.Collection{{Name: "a", Sources: []string{"s1", "s2"}}},
	}
	s := New(cfg, sources, clock, noopLogger{}, notifier)

	single(s, clock) // s1 -> s2

	if !slices.Equal(notifier.sources, []string{"s2"}) {
		t.Errorf("notified sources = %v, want [s2]", notifier.sources)
	}

	if !slices.Equal(notifier.collections, []string{"a"}) {
		t.Errorf("notified collections = %v, want [a]", notifier.collections)
	}
}

func TestSinglePressEmptyCollectionSelectsNothing(t *testing.T) {
	empty := config.Collection{Name: "empty", Sources: nil}
	s, src, clock := newTestSwitcher(t, "x", []string{"x"}, empty)

	single(s, clock)

	if len(src.selected) != 0 {
		t.Fatalf("selected = %v, want none for an empty collection", src.selected)
	}
}

func TestReverseSinglePressAcrossCollectionsRealigns(t *testing.T) {
	a := config.Collection{Name: "a", Sources: []string{"a1"}}
	b := config.Collection{Name: "b", Sources: []string{"b1"}}
	s, src, clock := newTestSwitcher(t, "a1", []string{"a1", "b1"}, a, b)

	double(s, clock)          // a -> b: selects b1, previous source becomes a1
	pressOnce(s, clock, true) // reverse single: jump to a1, which lives in collection a

	if got := src.selected[len(src.selected)-1]; got != "a1" {
		t.Fatalf("last selected = %q, want a1 (selected: %v)", got, src.selected)
	}

	// A forward single press must now cycle within collection a (its only source),
	// confirming the switcher realigned onto a rather than staying on b.
	single(s, clock)

	if got := src.selected[len(src.selected)-1]; got != "a1" {
		t.Fatalf("after realignment, cycling produced %q, want a1 (selected: %v)", got, src.selected)
	}
}

func TestImmediateDoublePressRestoresFromEmptyState(t *testing.T) {
	a := config.Collection{Name: "a", Sources: []string{"a1"}}
	b := config.Collection{Name: "b", Sources: []string{"b1"}}
	sources := &fakeSources{current: "a1", all: []string{"a1", "b1"}}
	clock := &fakeClock{now: time.Unix(1000, 0)}
	cfg := config.Config{
		DoublePress: config.DoublePress{Enabled: true, MaxDelay: testDelay},
		Reverse:     config.Reverse{Enabled: true, Modifier: "shift"},
		Collections: []config.Collection{a, b},
	}
	s := New(cfg, sources, clock, noopLogger{}, nil)

	// The very first press arrives with zero elapsed time, so it is classified as
	// a double press while the pre-single snapshot is still empty. restore must
	// tolerate that empty state rather than clobbering the collection index.
	s.Press(false)

	if !slices.Equal(sources.selected, []string{"b1"}) {
		t.Fatalf("selected = %v, want [b1] (double press advanced a -> b)", sources.selected)
	}
}
