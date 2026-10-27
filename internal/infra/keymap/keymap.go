// Package keymap remaps the configured trigger key on each keyboard's own HID key
// mapping table: to fn (the Globe key's usage), so a keyboard without a Globe
// key drives the daemon through its existing Globe key handling, or to a key
// that produces no event, for a dedicated trigger the daemon reads raw.
//
// The tables are shared with hidutil and other tools, so the remapper only ever
// touches its own entries: it merges into a keyboard's current table, remembers
// any entry it displaced for the same key, and puts that back when it lets go.
// What it added is recorded in a state file, so a crashed run is cleaned up on
// the next. Keyboards come and go, so the daemon calls Sync periodically.
package keymap

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

// keyboardPage is the HID Keyboard/Keypad usage page, shifted into place.
const keyboardPage = 0x07 << 32

// GlobeUsage is the Globe key: usage 0x03 on the Apple vendor top-case page.
const GlobeUsage uint64 = 0xFF<<32 | 0x03

// DedicatedUsage is the key a dedicated trigger is remapped to: F24, which
// macOS has no key code for, so the key produces no event at all and the daemon
// watches it in the keyboard's raw input.
const DedicatedUsage uint64 = keyboardPage | 0x73

// stateFileMode is the permission mode used when writing the state file.
const stateFileMode os.FileMode = 0o600

// stateVersion identifies the per-keyboard state file format. A file without a
// version holds a single entry in the system-wide table.
const stateVersion = 2

// Vendor IDs Apple keyboards report over USB and Bluetooth.
const (
	appleUSBVendor       = 0x05AC
	appleBluetoothVendor = 0x004C
)

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

// Keyboard is a connected keyboard service and its own key mappings.
type Keyboard struct {
	// ID identifies the service while it is connected.
	ID uint64
	// Vendor, Product, and Location identify the keyboard across connections.
	Vendor, Product, Location uint32
	// Name is the product name, if the keyboard reports one.
	Name string
	// Transport is how the keyboard is connected, such as USB or Bluetooth.
	Transport string
	// BuiltIn reports a laptop's own keyboard.
	BuiltIn bool
	// Mappings is the keyboard's key mapping table.
	Mappings []Mapping
}

// KeyboardID is the keyboard's ID in configuration (see config.KeyboardID).
func (k Keyboard) KeyboardID() string {
	return config.KeyboardID(k.Vendor, k.Product)
}

// HasGlobe reports whether the keyboard has a Globe key of its own: Apple
// keyboards do.
func (k Keyboard) HasGlobe() bool {
	return k.BuiltIn || k.Vendor == appleUSBVendor || k.Vendor == appleBluetoothVendor
}

// Store reads and replaces key mapping tables: each keyboard's own, and the
// system-wide one older runs used.
type Store interface {
	Keyboards() ([]Keyboard, error)
	SetKeyboardMappings(id uint64, mappings []Mapping) error
	Mappings() ([]Mapping, error)
	SetMappings(mappings []Mapping) error
}

// identity identifies a keyboard across connections.
type identity struct {
	Vendor   uint32 `json:"vendor"`
	Product  uint32 `json:"product"`
	Location uint32 `json:"location"`
}

func identityOf(keyboard Keyboard) identity {
	return identity{Vendor: keyboard.Vendor, Product: keyboard.Product, Location: keyboard.Location}
}

// entry records a mapping the remapper added to a keyboard and the one it
// displaced, if any.
type entry struct {
	identity

	Mapping  Mapping  `json:"mapping"`
	Replaced *Mapping `json:"replaced,omitempty"`
}

// state is the state file: the remapper's entries, or the single system-wide
// entry an older run left behind.
type state struct {
	Version int     `json:"version"`
	Entries []entry `json:"entries,omitempty"`

	// Src and Replaced hold a system-wide entry from an unversioned file.
	Src      uint64   `json:"src,omitempty"`
	Replaced *Mapping `json:"replaced,omitempty"`
}

// Remapper keeps the trigger key's mapping on every keyboard in scope.
type Remapper struct {
	mu        sync.Mutex
	store     Store
	statePath string
	trigger   config.Trigger
	active    bool
	observer  func([]Applied)
}

// New returns a Remapper that edits store and records its entries at statePath.
func New(store Store, statePath string) *Remapper {
	return &Remapper{store: store, statePath: statePath}
}

// Apply makes every keyboard match trigger: the trigger key remapped on the
// keyboards in scope when it is remapped, and no entry of ours otherwise.
func (r *Remapper) Apply(trigger config.Trigger) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.trigger, r.active = trigger, true

	return r.sync()
}

// Restore removes the remapper's entries and puts back the ones they displaced.
func (r *Remapper) Restore() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.active = false

	return r.sync()
}

// Sync re-applies the last trigger, covering keyboards connected since and
// entries the system dropped, for example across sleep. It is idempotent.
func (r *Remapper) Sync() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.sync()
}

// Run calls Sync every interval until ctx is done, reporting an error only
// when it differs from the previous one.
func (r *Remapper) Run(ctx context.Context, interval time.Duration, onError func(error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var last string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := r.Sync()

			message := ""
			if err != nil {
				message = err.Error()
			}

			if message != last && err != nil {
				onError(err)
			}

			last = message
		}
	}
}

// target returns the mapping the trigger wants, and false when it wants none.
func (r *Remapper) target() (Mapping, bool, error) {
	if !r.active || !r.trigger.Remapped() {
		return Mapping{}, false, nil
	}

	src, ok := Usages[r.trigger.Key]
	if !ok {
		return Mapping{}, false, fmt.Errorf("trigger key %q cannot be remapped", r.trigger.Key)
	}

	dst := GlobeUsage
	if r.trigger.Dedicated() {
		dst = DedicatedUsage
	}

	return Mapping{Src: src, Dst: dst}, true, nil
}

// change is a keyboard table sync rewrites: next replaces it, and the entry
// before gives way to the entry after.
type change struct {
	keyboard Keyboard
	next     []Mapping
	before   *entry
	after    *entry
}

// plan is what sync writes: the changed tables, the entries that may exist
// while they are written, and the entries that exist once they are.
type plan struct {
	changes []change
	pending []entry
	final   []entry
}

// sync brings every connected keyboard in line with the trigger.
func (r *Remapper) sync() error {
	want, wanted, err := r.target()
	if err != nil {
		return err
	}

	previous, err := r.load()
	if err != nil {
		return err
	}

	if previous.Src != 0 {
		if err = r.releaseSystemWide(previous); err != nil {
			return err
		}
	}

	keyboards, err := r.store.Keyboards()
	if err != nil {
		return fmt.Errorf("listing keyboards: %w", err)
	}

	var include func(Keyboard) bool
	if wanted {
		include = func(keyboard Keyboard) bool {
			return r.trigger.Keyboards.Includes(keyboard.KeyboardID(), keyboard.HasGlobe())
		}
	}

	p := planSync(previous.Entries, keyboards, want, include, r.active)

	// Record every entry that may exist before writing any table, and forget an
	// entry only after it is gone, so a crash in between never leaves one the
	// next run cannot clean up.
	if err = r.save(p.pending); err != nil {
		return err
	}

	final := p.final

	var errs []error

	for _, c := range p.changes {
		kept := c.after

		if writeErr := r.store.SetKeyboardMappings(c.keyboard.ID, c.next); writeErr != nil {
			errs = append(errs, fmt.Errorf("writing key mappings for %s: %w", describe(c.keyboard), writeErr))
			kept = c.before
		}

		if kept != nil {
			final = append(final, *kept)
		}
	}

	r.notify(final)

	return errors.Join(append(errs, r.save(final))...)
}

// planSync works out how to map want on the keyboards include accepts (none
// when include is nil), replacing the entries recorded. While active, entries
// of keyboards not listed are kept, in case a listing missed one; a keyboard
// that reconnects comes back without them.
func planSync(recorded []entry, keyboards []Keyboard, want Mapping, include func(Keyboard) bool, active bool) plan {
	var p plan

	connected := make(map[identity]bool, len(keyboards))
	for _, keyboard := range keyboards {
		connected[identityOf(keyboard)] = true
	}

	byIdentity := make(map[identity]entry, len(recorded))

	for _, e := range recorded {
		byIdentity[e.identity] = e

		if active && !connected[e.identity] {
			p.pending = append(p.pending, e)
			p.final = append(p.final, e)
		}
	}

	for _, keyboard := range keyboards {
		c := planKeyboard(keyboard, byIdentity, want, include)

		for _, e := range []*entry{c.before, c.after} {
			if e != nil {
				p.pending = append(p.pending, *e)
			}
		}

		switch {
		case !slices.Equal(c.next, keyboard.Mappings):
			p.changes = append(p.changes, c)
		case c.after != nil:
			p.final = append(p.final, *c.after)
		}
	}

	return p
}

// planKeyboard works out a keyboard's new table: the recorded entry released,
// and want claimed when include accepts the keyboard.
func planKeyboard(keyboard Keyboard, recorded map[identity]entry, want Mapping, include func(Keyboard) bool) change {
	c := change{keyboard: keyboard}

	if e, ok := recorded[identityOf(keyboard)]; ok {
		c.before = &e
	}

	c.next = release(keyboard.Mappings, c.before)

	if include != nil && include(keyboard) {
		c.next, c.after = claim(c.next, want)
		c.after.identity = identityOf(keyboard)
	}

	return c
}

// releaseSystemWide removes the system-wide entry an older run left behind.
func (r *Remapper) releaseSystemWide(previous state) error {
	current, err := r.store.Mappings()
	if err != nil {
		return fmt.Errorf("reading key mappings: %w", err)
	}

	next := release(current, &entry{Mapping: Mapping{Src: previous.Src, Dst: GlobeUsage}, Replaced: previous.Replaced})
	if slices.Equal(next, current) {
		return nil
	}

	if err = r.store.SetMappings(next); err != nil {
		return fmt.Errorf("writing key mappings: %w", err)
	}

	return nil
}

// describe names a keyboard in messages.
func describe(keyboard Keyboard) string {
	if keyboard.Name == "" {
		return keyboard.KeyboardID()
	}

	return fmt.Sprintf("%s (%s)", keyboard.Name, keyboard.KeyboardID())
}

// release removes the entry recorded in previous and restores what it displaced.
func release(mappings []Mapping, previous *entry) []Mapping {
	next := slices.Clone(mappings)
	if previous == nil {
		return next
	}

	// An entry that is no longer there was replaced by someone else, or dropped
	// with the keyboard's table; either way, what it displaced is not ours to restore.
	if !slices.Contains(next, previous.Mapping) {
		return next
	}

	next = slices.DeleteFunc(next, func(m Mapping) bool { return m == previous.Mapping })

	src := previous.Mapping.Src
	if previous.Replaced != nil && !slices.ContainsFunc(next, func(m Mapping) bool { return m.Src == src }) {
		next = append(next, *previous.Replaced)
	}

	return next
}

// claim adds want, taking out (and reporting) any entry for its source key.
func claim(mappings []Mapping, want Mapping) ([]Mapping, *entry) {
	added := &entry{Mapping: want}

	if index := slices.IndexFunc(mappings, func(m Mapping) bool { return m.Src == want.Src }); index >= 0 {
		replaced := mappings[index]
		if replaced != want {
			added.Replaced = &replaced
		}

		mappings = slices.Delete(slices.Clone(mappings), index, index+1)
	}

	return append(mappings, want), added
}

// load reads the recorded entries, or an empty state when there are none.
func (r *Remapper) load() (state, error) {
	data, err := os.ReadFile(r.statePath)
	if errors.Is(err, fs.ErrNotExist) {
		return state{}, nil
	}

	if err != nil {
		return state{}, fmt.Errorf("reading keymap state %q: %w", r.statePath, err)
	}

	var recorded state
	if err = json.Unmarshal(data, &recorded); err != nil {
		return state{}, fmt.Errorf("parsing keymap state %q: %w", r.statePath, err)
	}

	if recorded.Version != 0 && recorded.Version != stateVersion {
		return state{}, fmt.Errorf("keymap state %q has unknown version %d", r.statePath, recorded.Version)
	}

	return recorded, nil
}

// save records entries, or removes the state file when there are none.
func (r *Remapper) save(entries []entry) error {
	if len(entries) == 0 {
		if err := os.Remove(r.statePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("removing keymap state %q: %w", r.statePath, err)
		}

		return nil
	}

	data, err := json.Marshal(state{Version: stateVersion, Entries: entries})
	if err != nil {
		return fmt.Errorf("encoding keymap state: %w", err)
	}

	if err = os.WriteFile(r.statePath, data, stateFileMode); err != nil {
		return fmt.Errorf("writing keymap state %q: %w", r.statePath, err)
	}

	return nil
}

// Applied is a mapping the remapper keeps on a keyboard.
type Applied struct {
	// Vendor, Product, and Location identify the keyboard.
	Vendor, Product, Location uint32
	Mapping                   Mapping
}

// Observe registers fn to receive the mappings in effect after every sync.
func (r *Remapper) Observe(fn func([]Applied)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.observer = fn
}

// notify reports the entries in effect to the observer.
func (r *Remapper) notify(entries []entry) {
	if r.observer == nil {
		return
	}

	applied := make([]Applied, 0, len(entries))
	for _, e := range entries {
		applied = append(applied, Applied{
			Vendor:   e.Vendor,
			Product:  e.Product,
			Location: e.Location,
			Mapping:  e.Mapping,
		})
	}

	r.observer(applied)
}

// Status is a connected keyboard as the remapper sees it.
type Status struct {
	Keyboard

	// InScope reports whether trigger's remap applies to the keyboard.
	InScope bool
	// Ours is the remapper's entry on the keyboard, if any.
	Ours *Mapping
	// Others are the keyboard's other mappings, set by other tools.
	Others []Mapping
}

// Keyboards reports every connected keyboard against trigger, merging the
// services of one physical keyboard, and the system-wide mappings.
func (r *Remapper) Keyboards(trigger config.Trigger) ([]Status, []Mapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	recorded, err := r.load()
	if err != nil {
		return nil, nil, err
	}

	keyboards, err := r.store.Keyboards()
	if err != nil {
		return nil, nil, fmt.Errorf("listing keyboards: %w", err)
	}

	systemWide, err := r.store.Mappings()
	if err != nil {
		return nil, nil, fmt.Errorf("reading key mappings: %w", err)
	}

	ours := make(map[identity]Mapping, len(recorded.Entries))
	for _, e := range recorded.Entries {
		ours[e.identity] = e.Mapping
	}

	var (
		statuses []Status
		index    = map[identity]int{}
	)

	for _, keyboard := range keyboards {
		id := identityOf(keyboard)

		i, seen := index[id]
		if !seen {
			i = len(statuses)
			index[id] = i

			statuses = append(statuses, Status{
				Keyboard: keyboard,
				InScope:  trigger.Remapped() && trigger.Keyboards.Includes(keyboard.KeyboardID(), keyboard.HasGlobe()),
			})
			statuses[i].Mappings = nil
		}

		status := &statuses[i]
		status.Name = cmp.Or(status.Name, keyboard.Name)

		for _, mapping := range keyboard.Mappings {
			if mine, ok := ours[id]; ok && mapping == mine {
				status.Ours = &mine

				continue
			}

			if !slices.Contains(status.Others, mapping) {
				status.Others = append(status.Others, mapping)
			}
		}
	}

	return statuses, systemWide, nil
}
