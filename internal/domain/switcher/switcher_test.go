package switcher

import (
	"slices"
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

const testDelay = 250 * time.Millisecond

// fakeSources records every Select call and reports a configurable source set.
type fakeSources struct {
	current  string
	all      []string
	selected []string
}

func (f *fakeSources) Current() string { return f.current }
func (f *fakeSources) All() []string   { return f.all }

func (f *fakeSources) Select(id string) {
	f.current = id
	f.selected = append(f.selected, id)
}

// fakeClock is a manually advanced clock.
type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) advance(d time.Duration) { c.now = c.now.Add(d) }

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

func newTestSwitcher(
	t *testing.T, current string, all []string, collections ...config.Collection,
) (*Switcher, *fakeSources, *fakeClock) {
	t.Helper()

	sources := &fakeSources{current: current, all: all}
	clock := &fakeClock{now: time.Unix(1000, 0)}
	cfg := config.Config{
		DoublePress: config.DoublePress{Enabled: true, MaxDelay: testDelay},
		Reverse:     config.Reverse{Enabled: true, Modifier: "shift"},
		Collections: collections,
	}

	return New(cfg, sources, clock, noopLogger{}, nil), sources, clock
}

// single simulates a single press: enough time elapses to not be a double press.
func single(s *Switcher, c *fakeClock) {
	pressOnce(s, c, false)
}

// double simulates a user double press, which the system sees as a single press
// (the first release) followed by a double press (the second release).
func double(s *Switcher, c *fakeClock) {
	single(s, c)
	c.advance(testDelay)
	s.Press(false)
}

// pressOnce simulates a single press in the given direction.
func pressOnce(s *Switcher, c *fakeClock, reverse bool) {
	c.advance(testDelay + time.Millisecond)
	s.Press(reverse)
}

// doubleReverse simulates a reverse double press (e.g. Shift held for both releases).
func doubleReverse(s *Switcher, c *fakeClock) {
	pressOnce(s, c, true)
	c.advance(testDelay)
	s.Press(true)
}

func TestSinglePressCyclesWithinCollectionWithWraparound(t *testing.T) {
	col := config.Collection{Name: "a", Sources: []string{"s1", "s2", "s3"}}
	s, src, clock := newTestSwitcher(t, "s1", []string{"s1", "s2", "s3"}, col)

	single(s, clock)
	single(s, clock)
	single(s, clock)

	want := []string{"s2", "s3", "s1"}
	if !slices.Equal(src.selected, want) {
		t.Fatalf("selected = %v, want %v", src.selected, want)
	}
}

func TestSinglePressStartsAtFirstWhenCurrentNotInCollection(t *testing.T) {
	col := config.Collection{Name: "a", Sources: []string{"s1", "s2"}}
	s, src, clock := newTestSwitcher(t, "unknown", []string{"s1", "s2"}, col)

	single(s, clock)
	single(s, clock)

	want := []string{"s1", "s2"}
	if !slices.Equal(src.selected, want) {
		t.Fatalf("selected = %v, want %v", src.selected, want)
	}
}

func TestInitialCollectionDerivedFromCurrentSource(t *testing.T) {
	a := config.Collection{Name: "a", Sources: []string{"a1"}}
	b := config.Collection{Name: "b", Sources: []string{"b1", "b2"}}
	// current source b1 lives in collection b, so the first single press must
	// cycle within b, not the first-declared collection a.
	s, src, clock := newTestSwitcher(t, "b1", []string{"a1", "b1", "b2"}, a, b)

	single(s, clock)

	if got := src.selected; len(got) != 1 || got[0] != "b2" {
		t.Fatalf("selected = %v, want [b2]", got)
	}
}

func TestDoublePressSwitchesCollectionsInDeterministicOrder(t *testing.T) {
	a := config.Collection{Name: "a", Sources: []string{"a1"}}
	b := config.Collection{Name: "b", Sources: []string{"b1"}}
	c := config.Collection{Name: "c", Sources: []string{"c1"}}
	s, src, clock := newTestSwitcher(t, "a1", []string{"a1", "b1", "c1"}, a, b, c)

	double(s, clock) // a -> b
	double(s, clock) // b -> c
	double(s, clock) // c -> a (wraparound)

	// The source selected by each double press reveals the destination collection.
	switched := []string{src.selected[1], src.selected[3], src.selected[5]}
	want := []string{"b1", "c1", "a1"}

	if !slices.Equal(switched, want) {
		t.Fatalf("collection switch order = %v, want %v", switched, want)
	}
}

func TestDoublePressRevertsTheImplicitSinglePress(t *testing.T) {
	a := config.Collection{Name: "a", Sources: []string{"a1", "a2", "a3"}}
	b := config.Collection{Name: "b", Sources: []string{"b1"}}
	s, src, clock := newTestSwitcher(t, "a1", []string{"a1", "a2", "a3", "b1"}, a, b)

	single(s, clock) // a1 -> a2
	single(s, clock) // a2 -> a3, user is now on a3
	double(s, clock) // jump to b (the implicit single must not strand a on a1)
	double(s, clock) // jump back to a

	// Returning to a must resume on a3, where the user actually was.
	if got := src.selected[len(src.selected)-1]; got != "a3" {
		t.Fatalf("resumed collection a on %q, want a3 (selected: %v)", got, src.selected)
	}
}

func TestDoublePressDetectionRespectsDelayBoundary(t *testing.T) {
	a := config.Collection{Name: "a", Sources: []string{"a1", "a2"}}
	b := config.Collection{Name: "b", Sources: []string{"b1"}}
	s, src, clock := newTestSwitcher(t, "a1", []string{"a1", "a2", "b1"}, a, b)

	clock.advance(testDelay + time.Millisecond) // > delay -> single
	s.Press(false)
	clock.advance(testDelay) // exactly == delay -> still a double press
	s.Press(false)

	want := []string{"a2", "b1"}
	if !slices.Equal(src.selected, want) {
		t.Fatalf("selected = %v, want %v", src.selected, want)
	}
}

func TestUnavailableSourceIsNotSelected(t *testing.T) {
	col := config.Collection{Name: "a", Sources: []string{"s1", "s2"}}
	// s2 is configured but not actually available on the system.
	s, src, clock := newTestSwitcher(t, "s1", []string{"s1"}, col)

	single(s, clock) // would advance to s2, which is unavailable

	if len(src.selected) != 0 {
		t.Fatalf("selected = %v, want none (unavailable source must not be selected)", src.selected)
	}
}

func TestNoCollectionsIsNoOp(t *testing.T) {
	s, src, clock := newTestSwitcher(t, "x", []string{"x"})

	single(s, clock)
	double(s, clock)

	if len(src.selected) != 0 {
		t.Fatalf("selected = %v, want none", src.selected)
	}
}

func TestReverseSinglePressTogglesToPreviousSource(t *testing.T) {
	col := config.Collection{Name: "a", Sources: []string{"s1", "s2", "s3"}}
	s, src, clock := newTestSwitcher(t, "s1", []string{"s1", "s2", "s3"}, col)

	single(s, clock)          // forward s1 -> s2 (previous becomes s1)
	pressOnce(s, clock, true) // reverse -> previous source s1
	pressOnce(s, clock, true) // reverse -> previous source s2 (toggles s1 <-> s2)

	want := []string{"s2", "s1", "s2"}
	if !slices.Equal(src.selected, want) {
		t.Fatalf("selected = %v, want %v", src.selected, want)
	}
}

func TestReverseSinglePressIsNoOpWithoutHistory(t *testing.T) {
	col := config.Collection{Name: "a", Sources: []string{"s1", "s2"}}
	s, src, clock := newTestSwitcher(t, "s1", []string{"s1", "s2"}, col)

	pressOnce(s, clock, true) // no previous source yet -> nothing happens

	if len(src.selected) != 0 {
		t.Fatalf("selected = %v, want none", src.selected)
	}
}

func TestReverseDoublePressSwitchesToPreviousCollection(t *testing.T) {
	a := config.Collection{Name: "a", Sources: []string{"a1"}}
	b := config.Collection{Name: "b", Sources: []string{"b1"}}
	c := config.Collection{Name: "c", Sources: []string{"c1"}}
	s, src, clock := newTestSwitcher(t, "a1", []string{"a1", "b1", "c1"}, a, b, c)

	doubleReverse(s, clock) // a -> c (wrap backward)

	if got := src.selected[len(src.selected)-1]; got != "c1" {
		t.Fatalf("reverse double press selected %q, want c1 (previous collection)", got)
	}
}

func TestDoublePressDisabledTreatsQuickPressesAsSingles(t *testing.T) {
	sources := &fakeSources{current: "a1", all: []string{"a1", "a2", "b1"}}
	clock := &fakeClock{now: time.Unix(1000, 0)}
	cfg := config.Config{
		DoublePress: config.DoublePress{Enabled: false, MaxDelay: testDelay},
		Reverse:     config.Reverse{Enabled: true, Modifier: "shift"},
		Collections: []config.Collection{
			{Name: "a", Sources: []string{"a1", "a2"}},
			{Name: "b", Sources: []string{"b1"}},
		},
	}
	s := New(cfg, sources, clock, noopLogger{}, nil)

	single(s, clock) // a1 -> a2
	clock.advance(testDelay)
	s.Press(false) // quick: would switch collections if double press were enabled

	if sources.current != "a1" {
		t.Fatalf("current = %q, want a1 (double press disabled: cycled within collection)", sources.current)
	}
}

func TestReverseDisabledPressesForward(t *testing.T) {
	sources := &fakeSources{current: "s1", all: []string{"s1", "s2", "s3"}}
	clock := &fakeClock{now: time.Unix(1000, 0)}
	cfg := config.Config{
		DoublePress: config.DoublePress{Enabled: true, MaxDelay: testDelay},
		Reverse:     config.Reverse{Enabled: false, Modifier: "shift"},
		Collections: []config.Collection{{Name: "a", Sources: []string{"s1", "s2", "s3"}}},
	}
	s := New(cfg, sources, clock, noopLogger{}, nil)

	pressOnce(s, clock, true) // reverse requested but disabled -> forward s1 -> s2

	if sources.current != "s2" {
		t.Fatalf("current = %q, want s2 (reverse disabled falls back to forward)", sources.current)
	}
}
