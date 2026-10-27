// Package keymap remaps the configured trigger key to the Globe (fn) key in the
// system-wide HID key mapping table, so a keyboard without a Globe key can drive
// the daemon through its existing Globe handling.
//
// The table is shared with hidutil and other tools, so the remapper only ever
// touches its own entry: it merges into the current table, remembers any entry
// it displaced for the same key, and puts that back when it lets go. What it
// added is recorded in a state file, so a crashed run is cleaned up on the next.
package keymap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"sync"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

// keyboardPage is the HID Keyboard/Keypad usage page, shifted into place.
const keyboardPage = 0x07 << 32

// GlobeUsage is the Globe (fn) key: usage 0x03 on the Apple vendor top-case page.
const GlobeUsage uint64 = 0xFF<<32 | 0x03

// stateFileMode is the permission mode used when writing the state file.
const stateFileMode os.FileMode = 0o600

// Keyboard/Keypad page usages (HID Usage Tables, section 10), page included.
const (
	usageCapsLock     = keyboardPage | 0x39
	usageF13          = keyboardPage | 0x68
	usageF14          = keyboardPage | 0x69
	usageF15          = keyboardPage | 0x6A
	usageF16          = keyboardPage | 0x6B
	usageF17          = keyboardPage | 0x6C
	usageF18          = keyboardPage | 0x6D
	usageF19          = keyboardPage | 0x6E
	usageF20          = keyboardPage | 0x6F
	usageApplication  = keyboardPage | 0x65
	usageLeftControl  = keyboardPage | 0xE0
	usageLeftShift    = keyboardPage | 0xE1
	usageLeftOption   = keyboardPage | 0xE2
	usageLeftCommand  = keyboardPage | 0xE3
	usageRightControl = keyboardPage | 0xE4
	usageRightShift   = keyboardPage | 0xE5
	usageRightOption  = keyboardPage | 0xE6
	usageRightCommand = keyboardPage | 0xE7
)

// Usages maps every remappable trigger key to its HID usage.
var Usages = map[string]uint64{
	config.TriggerCapsLock: usageCapsLock,
	"right_command":        usageRightCommand,
	"right_option":         usageRightOption,
	"right_control":        usageRightControl,
	"right_shift":          usageRightShift,
	"left_command":         usageLeftCommand,
	"left_option":          usageLeftOption,
	"left_control":         usageLeftControl,
	"left_shift":           usageLeftShift,
	"f13":                  usageF13,
	"f14":                  usageF14,
	"f15":                  usageF15,
	"f16":                  usageF16,
	"f17":                  usageF17,
	"f18":                  usageF18,
	"f19":                  usageF19,
	"f20":                  usageF20,
	"menu":                 usageApplication,
}

// Mapping remaps the HID usage Src to Dst.
type Mapping struct {
	Src uint64 `json:"src"`
	Dst uint64 `json:"dst"`
}

// Store reads and replaces the system-wide key mapping table.
type Store interface {
	Mappings() ([]Mapping, error)
	SetMappings(mappings []Mapping) error
}

// state records the entry the remapper added and the one it displaced, if any.
type state struct {
	Src      uint64   `json:"src"`
	Replaced *Mapping `json:"replaced,omitempty"`
}

// Remapper keeps the trigger key's Globe mapping in a Store.
type Remapper struct {
	mu        sync.Mutex
	store     Store
	statePath string
}

// New returns a Remapper that edits store and records its entry at statePath.
func New(store Store, statePath string) *Remapper {
	return &Remapper{store: store, statePath: statePath}
}

// Apply makes the table match trigger: the trigger key mapped to Globe when it
// is remapped, and no entry of ours otherwise. It is idempotent and re-adds an
// entry the system dropped, for example across a restart.
func (r *Remapper) Apply(trigger config.Trigger) error {
	if !trigger.Remapped() {
		return r.Restore()
	}

	src, ok := Usages[trigger.Key]
	if !ok {
		return fmt.Errorf("trigger key %q cannot be remapped", trigger.Key)
	}

	return r.update(&src)
}

// Restore removes the remapper's entry and puts back the one it displaced.
func (r *Remapper) Restore() error {
	return r.update(nil)
}

// update releases any previous entry and, when src is set, maps it to Globe.
func (r *Remapper) update(src *uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	previous, err := r.load()
	if err != nil {
		return err
	}

	current, err := r.store.Mappings()
	if err != nil {
		return fmt.Errorf("reading key mappings: %w", err)
	}

	next := release(current, previous)

	var added *state

	if src != nil {
		next, added = claim(next, *src)
	}

	// Record an entry before adding it and forget one only after removing it, so
	// a crash in between never leaves an entry the next run cannot clean up.
	if added != nil {
		if err = r.save(added); err != nil {
			return err
		}
	}

	if !slices.Equal(next, current) {
		if err = r.store.SetMappings(next); err != nil {
			return fmt.Errorf("writing key mappings: %w", err)
		}
	}

	if added == nil {
		return r.save(nil)
	}

	return nil
}

// release removes the entry recorded in previous and restores what it displaced.
func release(mappings []Mapping, previous *state) []Mapping {
	next := slices.Clone(mappings)
	if previous == nil {
		return next
	}

	ours := Mapping{Src: previous.Src, Dst: GlobeUsage}
	next = slices.DeleteFunc(next, func(m Mapping) bool { return m == ours })

	if previous.Replaced != nil && !slices.ContainsFunc(next, func(m Mapping) bool { return m.Src == previous.Src }) {
		next = append(next, *previous.Replaced)
	}

	return next
}

// claim maps src to Globe, taking out (and reporting) any entry for src.
func claim(mappings []Mapping, src uint64) ([]Mapping, *state) {
	added := &state{Src: src}

	if index := slices.IndexFunc(mappings, func(m Mapping) bool { return m.Src == src }); index >= 0 {
		replaced := mappings[index]
		added.Replaced = &replaced
		mappings = slices.Delete(mappings, index, index+1)
	}

	return append(mappings, Mapping{Src: src, Dst: GlobeUsage}), added
}

// load reads the recorded entry, or nil when there is none.
func (r *Remapper) load() (*state, error) {
	data, err := os.ReadFile(r.statePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil //nolint:nilnil // no state file means nothing was added
	}

	if err != nil {
		return nil, fmt.Errorf("reading keymap state %q: %w", r.statePath, err)
	}

	var recorded state
	if err = json.Unmarshal(data, &recorded); err != nil {
		return nil, fmt.Errorf("parsing keymap state %q: %w", r.statePath, err)
	}

	return &recorded, nil
}

// save records added, or removes the state file when nothing was added.
func (r *Remapper) save(added *state) error {
	if added == nil {
		if err := os.Remove(r.statePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("removing keymap state %q: %w", r.statePath, err)
		}

		return nil
	}

	data, err := json.Marshal(added)
	if err != nil {
		return fmt.Errorf("encoding keymap state: %w", err)
	}

	if err = os.WriteFile(r.statePath, data, stateFileMode); err != nil {
		return fmt.Errorf("writing keymap state %q: %w", r.statePath, err)
	}

	return nil
}
