package keymap

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

// fakeStore is a set of in-memory keyboards plus the system-wide table.
type fakeStore struct {
	keyboards  []Keyboard
	systemWide []Mapping
	writes     int
	listErr    error
	writeErr   map[uint64]error
}

func (s *fakeStore) Keyboards() ([]Keyboard, error) {
	keyboards := slices.Clone(s.keyboards)
	for i := range keyboards {
		keyboards[i].Mappings = slices.Clone(keyboards[i].Mappings)
	}

	return keyboards, s.listErr
}

func (s *fakeStore) SetKeyboardMappings(id uint64, mappings []Mapping) error {
	if err := s.writeErr[id]; err != nil {
		return err
	}

	for i := range s.keyboards {
		if s.keyboards[i].ID == id {
			s.keyboards[i].Mappings = slices.Clone(mappings)
			s.writes++

			return nil
		}
	}

	return errors.New("no such keyboard")
}

func (s *fakeStore) Mappings() ([]Mapping, error) {
	return slices.Clone(s.systemWide), nil
}

func (s *fakeStore) SetMappings(mappings []Mapping) error {
	s.systemWide = slices.Clone(mappings)
	s.writes++

	return nil
}

func (s *fakeStore) keyboard(t *testing.T, id uint64) Keyboard {
	t.Helper()

	for _, keyboard := range s.keyboards {
		if keyboard.ID == id {
			return keyboard
		}
	}

	t.Fatalf("keyboard %d not found", id)

	return Keyboard{}
}

var (
	capsLock     = Usages[config.TriggerCapsLock]
	rightCommand = Usages["right_command"]
	userMapping  = Mapping{Src: keyboardPage | 0x04, Dst: keyboardPage | 0x05} // a -> b
	capsToGlobe  = Mapping{Src: capsLock, Dst: GlobeUsage}
)

// Keyboards: a built-in Apple one, a Magic Keyboard over Bluetooth, and a
// third-party USB one.
func newStore(mappings ...Mapping) *fakeStore {
	return &fakeStore{keyboards: []Keyboard{
		{ID: 1, Vendor: appleUSBVendor, Product: 0x0281, Location: 1, Name: "Apple Internal Keyboard", BuiltIn: true},
		{ID: 2, Vendor: appleBluetoothVendor, Product: 0x029c, Location: 2, Name: "Magic Keyboard"},
		{ID: 3, Vendor: 0x046d, Product: 0xc33f, Location: 3, Name: "Logitech", Mappings: slices.Clone(mappings)},
	}}
}

// IDs of the store's keyboards.
const (
	builtInID  = "05ac:0281"
	magicID    = "004c:029c"
	logitechID = "046d:c33f"
)

func caps(mode string) config.KeyboardTrigger {
	return config.KeyboardTrigger{Key: config.TriggerCapsLock, Mode: mode}
}

// everywhere makes trigger the trigger of every keyboard in the store.
func everywhere(trigger config.KeyboardTrigger) config.Trigger {
	return config.Trigger{
		Keyboards: map[string]config.KeyboardTrigger{builtInID: trigger, magicID: trigger, logitechID: trigger},
	}
}

// onLogitech makes trigger the trigger of the third-party keyboard.
func onLogitech(trigger config.KeyboardTrigger) config.Trigger {
	return config.Trigger{Keyboards: map[string]config.KeyboardTrigger{logitechID: trigger}}
}

// capsGlobe remaps Caps Lock to fn on the third-party keyboard.
var capsGlobe = onLogitech(caps(config.TriggerModeFn))

func newRemapper(t *testing.T, store *fakeStore) (*Remapper, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "keymap.json")

	return New(store, path), path
}

func assertMappings(t *testing.T, store *fakeStore, id uint64, want ...Mapping) {
	t.Helper()

	if got := store.keyboard(t, id).Mappings; !slices.Equal(got, want) {
		t.Errorf("keyboard %d mappings = %+v, want %+v", id, got, want)
	}
}

func apply(t *testing.T, remapper *Remapper, trigger config.Trigger) {
	t.Helper()

	if err := remapper.Apply(trigger); err != nil {
		t.Fatalf("Apply: %v", err)
	}
}

func restore(t *testing.T, remapper *Remapper) {
	t.Helper()

	if err := remapper.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}
}

func assertNoState(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("state file left behind: %v", err)
	}
}

func TestUsagesCoverRemappableTriggerKeys(t *testing.T) {
	for _, key := range config.TriggerKeys {
		_, ok := Usages[key]
		if ok == (key == config.TriggerGlobe) {
			t.Errorf("Usages[%q] present = %v", key, ok)
		}
	}
}

func TestTriggerKeyFor(t *testing.T) {
	for key, usage := range Usages {
		if got, ok := TriggerKeyFor(usage); !ok || got != key {
			t.Errorf("TriggerKeyFor(%#x) = %q, %v; want %q", usage, got, ok, key)
		}
	}

	if got, ok := TriggerKeyFor(GlobeUsage); !ok || got != config.TriggerGlobe {
		t.Errorf("TriggerKeyFor(Globe) = %q, %v", got, ok)
	}

	if _, ok := TriggerKeyFor(keyboardPage | 0x04); ok {
		t.Error("A is a trigger key")
	}
}

func TestDedicatedUsageIsNoTriggerKey(t *testing.T) {
	for key, usage := range Usages {
		if usage == DedicatedUsage {
			t.Errorf("DedicatedUsage is the usage of trigger key %q", key)
		}
	}
}

func TestObserveReportsResolvedTriggers(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)

	var got []Resolved

	remapper.Observe(func(resolved []Resolved) { got = resolved })

	insert := config.KeyboardTrigger{Key: "insert", Mode: config.TriggerModePassThrough}
	apply(t, remapper, config.Trigger{
		Keyboards: map[string]config.KeyboardTrigger{magicID: insert, logitechID: caps(config.TriggerModeDedicated)},
	})

	want := []Resolved{
		{Vendor: appleUSBVendor, Product: 0x0281, Location: 1, Trigger: config.GlobeTrigger},
		{Vendor: appleBluetoothVendor, Product: 0x029c, Location: 2, Trigger: insert},
		{Vendor: 0x046d, Product: 0xc33f, Location: 3, Trigger: caps(config.TriggerModeDedicated), Applied: true},
	}
	if !slices.Equal(got, want) {
		t.Errorf("resolved = %+v, want %+v", got, want)
	}

	restore(t, remapper)

	if got != nil {
		t.Errorf("resolved after restore = %+v, want none", got)
	}
}

func TestHasGlobe(t *testing.T) {
	for _, keyboard := range newStore().keyboards {
		if got, want := keyboard.HasGlobe(), keyboard.ID != 3; got != want {
			t.Errorf("%s HasGlobe = %v, want %v", keyboard.Name, got, want)
		}
	}
}

func TestApplyRemapsKeyboardsWithoutGlobe(t *testing.T) {
	store := newStore(userMapping)
	remapper, path := newRemapper(t, store)

	apply(t, remapper, capsGlobe)

	assertMappings(t, store, 1)
	assertMappings(t, store, 2)
	assertMappings(t, store, 3, userMapping, capsToGlobe)

	if _, err := os.Stat(path); err != nil {
		t.Errorf("state file not written: %v", err)
	}

	restore(t, remapper)
	assertMappings(t, store, 3, userMapping)
	assertNoState(t, path)
}

func TestApplyPerKeyboard(t *testing.T) {
	capsToF24 := Mapping{Src: capsLock, Dst: DedicatedUsage}
	commandToFn := Mapping{Src: rightCommand, Dst: GlobeUsage}

	tests := []struct {
		name    string
		trigger config.Trigger
		want    map[uint64][]Mapping
	}{
		{"every keyboard", everywhere(caps(config.TriggerModeFn)), map[uint64][]Mapping{
			1: {capsToGlobe}, 2: {capsToGlobe}, 3: {capsToGlobe},
		}},
		{"apple keyboard listed", config.Trigger{
			Keyboards: map[string]config.KeyboardTrigger{magicID: caps(config.TriggerModeFn)},
		}, map[uint64][]Mapping{2: {capsToGlobe}}},
		{"none listed", config.Trigger{}, map[uint64][]Mapping{}},
		{"a key and mode per keyboard", config.Trigger{
			Keyboards: map[string]config.KeyboardTrigger{
				magicID:    {Key: "right_command", Mode: config.TriggerModeFn},
				logitechID: caps(config.TriggerModeDedicated),
			},
		}, map[uint64][]Mapping{2: {commandToFn}, 3: {capsToF24}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newStore()
			remapper, _ := newRemapper(t, store)

			apply(t, remapper, test.trigger)

			for _, keyboard := range store.keyboards {
				assertMappings(t, store, keyboard.ID, test.want[keyboard.ID]...)
			}
		})
	}
}

func TestApplyDedicatedMapsToDedicatedKey(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)

	apply(t, remapper, everywhere(caps(config.TriggerModeDedicated)))

	for _, keyboard := range store.keyboards {
		assertMappings(t, store, keyboard.ID, Mapping{Src: capsLock, Dst: DedicatedUsage})
	}

	apply(t, remapper, capsGlobe)
	assertMappings(t, store, 1)
	assertMappings(t, store, 3, capsToGlobe)
}

func TestApplyIsIdempotent(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)

	for range 3 {
		apply(t, remapper, capsGlobe)
	}

	if err := remapper.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	assertMappings(t, store, 3, capsToGlobe)

	if store.writes != 1 {
		t.Errorf("writes = %d, want 1", store.writes)
	}
}

func TestSyncRemapsNewAndResetKeyboards(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)

	apply(t, remapper, capsGlobe)

	store.keyboards[2].Mappings = nil // the system dropped it, e.g. across sleep
	store.keyboards = append(store.keyboards, Keyboard{ID: 4, Vendor: 0x046d, Product: 0xc33f, Location: 4})

	if err := remapper.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	assertMappings(t, store, 3, capsToGlobe)
	assertMappings(t, store, 4, capsToGlobe)
}

func TestSyncSurvivesReconnect(t *testing.T) {
	store := newStore()
	remapper, path := newRemapper(t, store)

	apply(t, remapper, capsGlobe)

	unplugged := store.keyboards[2]
	store.keyboards = store.keyboards[:2]

	if err := remapper.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	// It comes back as a new service, without the mapping.
	unplugged.ID, unplugged.Mappings = 5, nil
	store.keyboards = append(store.keyboards, unplugged)

	if err := remapper.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	assertMappings(t, store, 5, capsToGlobe)

	restore(t, remapper)
	assertMappings(t, store, 5)
	assertNoState(t, path)
}

func TestApplyReplacesAndRestoresSameKeyMapping(t *testing.T) {
	theirs := Mapping{Src: capsLock, Dst: keyboardPage | 0xE0} // caps lock -> left control
	store := newStore(theirs, userMapping)
	remapper, path := newRemapper(t, store)

	apply(t, remapper, capsGlobe)
	assertMappings(t, store, 3, userMapping, capsToGlobe)

	// Re-applying must not lose the displaced entry.
	apply(t, remapper, capsGlobe)
	restore(t, remapper)

	assertMappings(t, store, 3, userMapping, theirs)
	assertNoState(t, path)
}

func TestApplySwitchesKeys(t *testing.T) {
	store := newStore(userMapping)
	remapper, _ := newRemapper(t, store)

	apply(t, remapper, capsGlobe)

	rightCommandFn := config.KeyboardTrigger{Key: "right_command", Mode: config.TriggerModeFn}
	apply(t, remapper, onLogitech(rightCommandFn))

	assertMappings(t, store, 3, userMapping, Mapping{Src: rightCommand, Dst: GlobeUsage})
}

func TestApplyWithoutRemapRemovesEntries(t *testing.T) {
	for name, trigger := range map[string]config.Trigger{
		"none listed":     {},
		"listed as Globe": onLogitech(config.GlobeTrigger),
		"pass-through":    onLogitech(config.KeyboardTrigger{Key: "insert", Mode: config.TriggerModePassThrough}),
	} {
		t.Run(name, func(t *testing.T) {
			store := newStore(userMapping)
			remapper, path := newRemapper(t, store)

			apply(t, remapper, capsGlobe)
			apply(t, remapper, trigger)

			assertMappings(t, store, 3, userMapping)
			assertNoState(t, path)
		})
	}
}

func TestNewCleansUpAfterCrash(t *testing.T) {
	store := newStore(userMapping)
	path := filepath.Join(t.TempDir(), "keymap.json")

	apply(t, New(store, path), capsGlobe)

	// A fresh remapper (the next run) finds the leftover entry and removes it.
	restore(t, New(store, path))

	assertMappings(t, store, 3, userMapping)
	assertNoState(t, path)
}

func TestRestoreCleansUpSystemWideEntry(t *testing.T) {
	theirs := Mapping{Src: capsLock, Dst: keyboardPage | 0x29}
	store := newStore()
	store.systemWide = []Mapping{userMapping, capsToGlobe}

	path := filepath.Join(t.TempDir(), "keymap.json")

	legacy, err := json.Marshal(map[string]any{"src": capsLock, "replaced": theirs})
	if err != nil {
		t.Fatal(err)
	}

	if err = os.WriteFile(path, legacy, stateFileMode); err != nil {
		t.Fatal(err)
	}

	restore(t, New(store, path))

	if !slices.Equal(store.systemWide, []Mapping{userMapping, theirs}) {
		t.Errorf("system-wide mappings = %+v", store.systemWide)
	}

	assertNoState(t, path)
}

func TestRestoreLeavesChangedEntryAlone(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)

	apply(t, remapper, capsGlobe)

	edited := Mapping{Src: capsLock, Dst: keyboardPage | 0x29} // the user remapped it to Escape meanwhile
	store.keyboards[2].Mappings = []Mapping{edited}

	restore(t, remapper)
	assertMappings(t, store, 3, edited)
}

func TestRestoreWithoutStateIsNoOp(t *testing.T) {
	store := newStore(userMapping)
	remapper, _ := newRemapper(t, store)

	restore(t, remapper)
	assertMappings(t, store, 3, userMapping)

	if store.writes != 0 {
		t.Errorf("writes = %d, want 0", store.writes)
	}
}

func TestApplyErrors(t *testing.T) {
	boom := errors.New("boom")

	t.Run("list", func(t *testing.T) {
		store := newStore()
		store.listErr = boom

		remapper, _ := newRemapper(t, store)
		if err := remapper.Apply(capsGlobe); !errors.Is(err, boom) {
			t.Errorf("err = %v, want %v", err, boom)
		}
	})

	t.Run("write keeps state for cleanup", func(t *testing.T) {
		store := newStore()
		store.writeErr = map[uint64]error{3: boom}
		remapper, path := newRemapper(t, store)

		all := everywhere(caps(config.TriggerModeFn))
		if err := remapper.Apply(all); !errors.Is(err, boom) {
			t.Errorf("err = %v, want %v", err, boom)
		}

		assertMappings(t, store, 1, capsToGlobe)

		if _, err := os.Stat(path); err != nil {
			t.Errorf("state file missing after failed write: %v", err)
		}

		store.writeErr = nil

		restore(t, remapper)
		assertMappings(t, store, 1)
		assertNoState(t, path)
	})

	t.Run("corrupt state", func(t *testing.T) {
		remapper, path := newRemapper(t, newStore())
		if err := os.WriteFile(path, []byte("{"), stateFileMode); err != nil {
			t.Fatal(err)
		}

		if err := remapper.Restore(); err == nil {
			t.Error("Restore accepted a corrupt state file")
		}
	})

	t.Run("unknown state version", func(t *testing.T) {
		remapper, path := newRemapper(t, newStore())
		if err := os.WriteFile(path, []byte(`{"version":9}`), stateFileMode); err != nil {
			t.Fatal(err)
		}

		if err := remapper.Restore(); err == nil {
			t.Error("Restore accepted an unknown state version")
		}
	})
}

func TestRunSyncsAndReportsErrorsOnce(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)
	apply(t, remapper, capsGlobe)

	store.listErr = errors.New("boom")

	ctx, cancel := context.WithCancel(t.Context())
	errs := make(chan error, 10)
	done := make(chan struct{})

	go func() {
		remapper.Run(ctx, time.Millisecond, func(err error) { errs <- err })
		close(done)
	}()

	if err := <-errs; err == nil {
		t.Error("Run reported a nil error")
	}

	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if len(errs) != 0 {
		t.Errorf("repeated error reported %d more times", len(errs))
	}
}

func TestKeyboards(t *testing.T) {
	store := newStore(userMapping)
	// A second service of the third-party keyboard, as receivers often expose.
	store.keyboards = append(store.keyboards, Keyboard{
		ID: 6, Vendor: 0x046d, Product: 0xc33f, Location: 3, Mappings: []Mapping{userMapping},
	})
	store.systemWide = []Mapping{userMapping}

	remapper, _ := newRemapper(t, store)
	apply(t, remapper, capsGlobe)

	statuses, systemWide, err := remapper.Keyboards(capsGlobe)
	if err != nil {
		t.Fatalf("Keyboards: %v", err)
	}

	if len(statuses) != 3 {
		t.Fatalf("statuses = %+v, want 3 keyboards", statuses)
	}

	logitech := statuses[2]
	if logitech.Trigger != caps(config.TriggerModeFn) || !logitech.Listed || logitech.Ours == nil ||
		*logitech.Ours != capsToGlobe {
		t.Errorf("logitech = %+v, want our caps lock mapping", logitech)
	}

	if !slices.Equal(logitech.Others, []Mapping{userMapping}) || logitech.Name != "Logitech" {
		t.Errorf("logitech = %+v, want the user's mapping once", logitech)
	}

	if statuses[0].Trigger != config.GlobeTrigger || statuses[0].Ours != nil {
		t.Errorf("built-in = %+v, want out of scope", statuses[0])
	}

	if !slices.Equal(systemWide, []Mapping{userMapping}) {
		t.Errorf("system-wide = %+v", systemWide)
	}
}

func TestUsageName(t *testing.T) {
	tests := map[uint64]string{
		GlobeUsage:          "fn",
		capsLock:            "Caps Lock",
		DedicatedUsage:      "F24",
		keyboardPage | 0x04: "A",
		keyboardPage | 0x1D: "Z",
		keyboardPage | 0x1E: "1",
		keyboardPage | 0x27: "0",
		keyboardPage | 0x3A: "F1",
		keyboardPage | 0x45: "F12",
		keyboardPage | 0x29: "Escape",
		0x0C<<32 | 0xE9:     "0xc000000e9",
	}

	for usage, want := range tests {
		if got := UsageName(usage); got != want {
			t.Errorf("UsageName(%#x) = %q, want %q", usage, got, want)
		}
	}
}
