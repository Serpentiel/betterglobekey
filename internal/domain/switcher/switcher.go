// Package switcher contains the core input-source switching logic driven by the
// Globe key. It is pure domain logic: all external concerns (reading/selecting
// input sources, the wall clock, logging) are injected as ports so the behaviour
// can be unit-tested in isolation.
package switcher

import (
	"slices"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

// InputSources is the port through which the switcher observes and controls the
// system's input sources.
type InputSources interface {
	// Current returns the ID of the currently active input source.
	Current() string
	// All returns the IDs of every selectable input source.
	All() []string
	// Select activates the input source with the given ID.
	Select(id string)
}

// Clock is the port through which the switcher reads the current time.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// Logger is the minimal logging port the switcher depends on.
type Logger interface {
	// Info logs an informational message with optional key-value context.
	Info(msg string, keysAndValues ...any)
	// Error logs an error message with optional key-value context.
	Error(msg string, keysAndValues ...any)
}

// Switcher cycles between input sources on a single Globe key press and between
// collections on a double press. It is not safe for concurrent use; presses are
// expected to arrive sequentially from a single event loop.
type Switcher struct {
	sources InputSources
	clock   Clock
	log     Logger

	maxDoublePressDelay time.Duration
	collections         []config.Collection

	lastSourceByCollection map[string]string
	currentCollection      int // index into collections, or -1 when there are none
	currentSource          string
	lastPress              time.Time
}

// New builds a Switcher and seeds its state from the current input source. The
// active collection is the one containing the current source, falling back to
// the first configured collection.
func New(cfg config.Config, sources InputSources, clock Clock, log Logger) *Switcher {
	s := &Switcher{
		sources:                sources,
		clock:                  clock,
		log:                    log,
		maxDoublePressDelay:    cfg.DoublePressMaxDelay,
		collections:            cfg.Collections,
		lastSourceByCollection: make(map[string]string, len(cfg.Collections)),
		currentCollection:      -1,
		currentSource:          sources.Current(),
		lastPress:              clock.Now(),
	}

	for i, collection := range s.collections {
		if slices.Contains(collection.Sources, s.currentSource) {
			s.currentCollection = i

			break
		}
	}

	if s.currentCollection < 0 && len(s.collections) > 0 {
		s.currentCollection = 0
	}

	return s
}

// Press handles a Globe key release, classifying it as a single or double press
// based on the time elapsed since the previous press.
func (s *Switcher) Press(reverse bool) {
	now := s.clock.Now()
	elapsed := now.Sub(s.lastPress)
	s.lastPress = now

	if elapsed <= s.maxDoublePressDelay {
		s.doublePress(reverse)
	} else {
		s.singlePress(reverse)
	}
}

// singlePress moves to the next (or previous, when reverse) input source within
// the current collection.
func (s *Switcher) singlePress(reverse bool) {
	s.log.Info("globe key pressed")

	if s.currentCollection < 0 {
		return
	}

	collection := s.collections[s.currentCollection]

	next := step(collection.Sources, s.currentSource, !reverse)
	if next == "" {
		s.log.Error("no input source available", "collection", collection.Name)

		return
	}

	s.lastSourceByCollection[collection.Name] = next
	s.selectSource(next)
}

// doublePress switches to the next (or previous, when reverse) collection. The
// first release of a double press has already been handled as a single press,
// advancing the current collection, so that advance is reverted before switching.
func (s *Switcher) doublePress(reverse bool) {
	s.log.Info("globe key double pressed")

	if s.currentCollection < 0 {
		return
	}

	s.revertLastSource(s.collections[s.currentCollection], reverse)

	count := len(s.collections)
	if reverse {
		s.currentCollection = (s.currentCollection - 1 + count) % count
	} else {
		s.currentCollection = (s.currentCollection + 1) % count
	}

	collection := s.collections[s.currentCollection]

	next := s.lastSourceByCollection[collection.Name]
	if next == "" {
		next = collection.Sources[0]
	}

	s.selectSource(next)
}

// selectSource records and activates the given input source if it is available.
func (s *Switcher) selectSource(id string) {
	s.currentSource = id

	if slices.Contains(s.sources.All(), id) {
		s.sources.Select(id)
		s.log.Info("input source set", "source", id)

		return
	}

	s.log.Info("input source not found", "source", id)
}

// revertLastSource undoes the advance applied by the single press that precedes
// a double press by stepping the remembered source in the opposite direction.
// The single press moved it in the !reverse direction, so the revert steps in
// the reverse direction.
func (s *Switcher) revertLastSource(collection config.Collection, reverse bool) {
	index := slices.Index(collection.Sources, s.lastSourceByCollection[collection.Name])
	if index < 0 {
		return
	}

	count := len(collection.Sources)
	if reverse {
		s.lastSourceByCollection[collection.Name] = collection.Sources[(index+1)%count]
	} else {
		s.lastSourceByCollection[collection.Name] = collection.Sources[(index-1+count)%count]
	}
}

// step returns the source after current (forward) or before it (when forward is
// false), wrapping around. If current is not present, the first source is
// returned. An empty slice yields "".
func step(sources []string, current string, forward bool) string {
	if len(sources) == 0 {
		return ""
	}

	index := slices.Index(sources, current)
	if index < 0 {
		return sources[0]
	}

	count := len(sources)
	if forward {
		return sources[(index+1)%count]
	}

	return sources[(index-1+count)%count]
}
