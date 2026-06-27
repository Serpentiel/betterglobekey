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
func (s *Switcher) Press() {
	now := s.clock.Now()
	elapsed := now.Sub(s.lastPress)
	s.lastPress = now

	if elapsed <= s.maxDoublePressDelay {
		s.doublePress()
	} else {
		s.singlePress()
	}
}

// singlePress advances to the next input source within the current collection.
func (s *Switcher) singlePress() {
	s.log.Info("globe key pressed")

	if s.currentCollection < 0 {
		return
	}

	collection := s.collections[s.currentCollection]

	next := nextInCycle(collection.Sources, s.currentSource)
	if next == "" {
		s.log.Error("no input source available", "collection", collection.Name)

		return
	}

	s.lastSourceByCollection[collection.Name] = next
	s.selectSource(next)
}

// doublePress switches to the next collection. The first release of a double
// press has already been handled as a single press, advancing the current
// collection, so that advance is reverted before switching.
func (s *Switcher) doublePress() {
	s.log.Info("globe key double pressed")

	if s.currentCollection < 0 {
		return
	}

	s.revertLastSource(s.collections[s.currentCollection])

	s.currentCollection = (s.currentCollection + 1) % len(s.collections)

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

// revertLastSource moves the remembered source for a collection one step back,
// undoing the advance applied by the single press that precedes a double press.
func (s *Switcher) revertLastSource(collection config.Collection) {
	index := slices.Index(collection.Sources, s.lastSourceByCollection[collection.Name])

	switch {
	case index > 0:
		s.lastSourceByCollection[collection.Name] = collection.Sources[index-1]
	case index == 0:
		s.lastSourceByCollection[collection.Name] = collection.Sources[len(collection.Sources)-1]
	}
}

// nextInCycle returns the source after current, wrapping around. If current is
// not present, the first source is returned. An empty slice yields "".
func nextInCycle(sources []string, current string) string {
	if len(sources) == 0 {
		return ""
	}

	index := slices.Index(sources, current)
	if index < 0 {
		return sources[0]
	}

	return sources[(index+1)%len(sources)]
}
