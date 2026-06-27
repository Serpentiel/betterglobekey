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

// Notifier is an optional port notified whenever the input source changes,
// receiving the source ID and the name of its collection.
type Notifier interface {
	// Notify reports that source (in the given collection) has been activated.
	Notify(source, collection string)
}

// state is a snapshot of the mutable switcher state. A double press restores the
// snapshot taken before the single press that necessarily precedes it, so the
// collection switch starts from a clean slate regardless of what that single
// press did.
type state struct {
	collection       int
	current          string
	previous         string
	lastByCollection map[string]string
}

// Switcher cycles between input sources on a single Globe key press and between
// collections on a double press; a reverse press jumps to the previously-used
// source (single) or the previous collection (double). It is not safe for
// concurrent use; presses arrive sequentially from a single event loop.
type Switcher struct {
	sources  InputSources
	clock    Clock
	log      Logger
	notifier Notifier

	maxDoublePressDelay time.Duration
	collections         []config.Collection

	lastSourceByCollection map[string]string
	currentCollection      int // index into collections, or -1 when there are none
	currentSource          string
	previousSource         string // the input source used before the current one
	lastPress              time.Time
	saved                  state // state captured before the most recent single press
}

// New builds a Switcher and seeds its state from the current input source. The
// active collection is the one containing the current source, falling back to
// the first configured collection.
func New(cfg config.Config, sources InputSources, clock Clock, log Logger, notifier Notifier) *Switcher {
	s := &Switcher{
		sources:                sources,
		clock:                  clock,
		log:                    log,
		notifier:               notifier,
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
// based on the time elapsed since the previous press. When reverse is true the
// press jumps to the previous source (single) or previous collection (double).
func (s *Switcher) Press(reverse bool) {
	now := s.clock.Now()
	elapsed := now.Sub(s.lastPress)
	s.lastPress = now

	if elapsed <= s.maxDoublePressDelay {
		s.doublePress(reverse)

		return
	}

	s.saved = s.snapshot()
	s.singlePress(reverse)
}

// singlePress advances to the next input source within the current collection,
// or jumps to the previously-used source when reverse is true.
func (s *Switcher) singlePress(reverse bool) {
	s.log.Info("globe key pressed")

	if s.currentCollection < 0 {
		return
	}

	if reverse {
		if s.previousSource != "" && s.previousSource != s.currentSource {
			s.moveTo(s.previousSource)
		}

		return
	}

	collection := s.collections[s.currentCollection]

	next := step(collection.Sources, s.currentSource, true)
	if next == "" {
		s.log.Error("no input source available", "collection", collection.Name)

		return
	}

	s.moveTo(next)
}

// doublePress switches to the next collection, or the previous collection when
// reverse is true. The single press that opened the double is first undone.
func (s *Switcher) doublePress(reverse bool) {
	s.log.Info("globe key double pressed")

	s.restore(s.saved)

	if s.currentCollection < 0 {
		return
	}

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

	s.moveTo(next)
}

// moveTo records the prior source, aligns the current collection with the chosen
// source, remembers the per-collection position, and activates the source if it
// is available.
func (s *Switcher) moveTo(id string) {
	if id != s.currentSource {
		s.previousSource = s.currentSource
	}

	s.currentSource = id

	if s.currentCollection < 0 || !slices.Contains(s.collections[s.currentCollection].Sources, id) {
		for i, collection := range s.collections {
			if slices.Contains(collection.Sources, id) {
				s.currentCollection = i

				break
			}
		}
	}

	if s.currentCollection >= 0 {
		s.lastSourceByCollection[s.collections[s.currentCollection].Name] = id
	}

	if slices.Contains(s.sources.All(), id) {
		s.sources.Select(id)
		s.log.Info("input source set", "source", id)

		if s.notifier != nil {
			s.notifier.Notify(id, s.collectionName())
		}

		return
	}

	s.log.Info("input source not found", "source", id)
}

// collectionName returns the name of the current collection, or "" if there is none.
func (s *Switcher) collectionName() string {
	if s.currentCollection < 0 {
		return ""
	}

	return s.collections[s.currentCollection].Name
}

// snapshot copies the current mutable state.
func (s *Switcher) snapshot() state {
	last := make(map[string]string, len(s.lastSourceByCollection))
	for name, id := range s.lastSourceByCollection {
		last[name] = id
	}

	return state{
		collection:       s.currentCollection,
		current:          s.currentSource,
		previous:         s.previousSource,
		lastByCollection: last,
	}
}

// restore reinstates a previously captured state.
func (s *Switcher) restore(snap state) {
	if snap.lastByCollection == nil {
		return
	}

	s.currentCollection = snap.collection
	s.currentSource = snap.current
	s.previousSource = snap.previous
	s.lastSourceByCollection = snap.lastByCollection
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
