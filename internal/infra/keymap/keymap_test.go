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

func caps(mode string, keyboards config.Keyboards) config.Trigger {
	if keyboards.ApplyTo == "" {
		keyboards.ApplyTo = config.KeyboardsWithoutGlobe
	}

	return config.Trigger{Key: config.TriggerCapsLock, Mode: mode, Keyboards: keyboards}
}

var capsGlobe = caps(config.TriggerModeFn, config.Keyboards{})

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

func TestDedicatedUsageIsNoTriggerKey(t *testing.T) {
	for key, usage := range Usages {
		if usage == DedicatedUsage {
			t.Errorf("DedicatedUsage is the usage of trigger key %q", key)
		}
	}
}

func TestObserveReportsAppliedMappings(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)

	var got []Applied

	remapper.Observe(func(applied []Applied) { got = applied })

	apply(t, remapper, caps(config.TriggerModeDedicated, config.Keyboards{}))

	silenced := Mapping{Src: capsLock, Dst: DedicatedUsage}

	want := []Applied{{Vendor: 0x046d, Product: 0xc33f, Location: 3, Mapping: silenced}}
	if !slices.Equal(got, want) {
		t.Errorf("applied = %+v, want %+v", got, want)
	}

	restore(t, remapper)

	if len(got) != 0 {
		t.Errorf("applied after restore = %+v, want none", got)
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

func TestApplyScopes(t *testing.T) {
	tests := []struct {
		name      string
		keyboards config.Keyboards
		remapped  []uint64
	}{
		{"all", config.Keyboards{ApplyTo: config.KeyboardsAll}, []uint64{1, 2, 3}},
		{
			"apple keyboard included",
			config.Keyboards{Overrides: map[string]bool{"004c:029c": true}},
			[]uint64{2, 3},
		},
		{
			"third-party keyboard excluded",
			config.Keyboards{ApplyTo: config.KeyboardsAll, Overrides: map[string]bool{"046d:c33f": false}},
			[]uint64{1, 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newStore()
			remapper, _ := newRemapper(t, store)

			apply(t, remapper, caps(config.TriggerModeFn, test.keyboards))

			for _, keyboard := range store.keyboards {
				want := []Mapping(nil)
				if slices.Contains(test.remapped, keyboard.ID) {
					want = []Mapping{capsToGlobe}
				}

				assertMappings(t, store, keyboard.ID, want...)
			}
		})
	}
}

func TestApplyDedicatedMapsToDedicatedKey(t *testing.T) {
	store := newStore()
	remapper, _ := newRemapper(t, store)

	apply(t, remapper, caps(config.TriggerModeDedicated, config.Keyboards{ApplyTo: config.KeyboardsAll}))

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
	store.keyboards = append(store.keyboards, Keyboard{ID: 4, Vendor: 0x1234, Product: 0x5678, Location: 4})

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
	apply(t, remapper, config.Trigger{
		Key:       "right_command",
		Mode:      config.TriggerModeFn,
		Keyboards: config.Keyboards{ApplyTo: config.KeyboardsWithoutGlobe},
	})

	assertMappings(t, store, 3, userMapping, Mapping{Src: rightCommand, Dst: GlobeUsage})
}

func TestApplyWithoutRemapRemovesEntries(t *testing.T) {
	for name, trigger := range map[string]config.Trigger{
		"globe":        {Key: config.TriggerGlobe, Mode: config.TriggerModeFn},
		"pass-through": {Key: "right_command", Mode: config.TriggerModePassThrough},
		"out of scope": caps(config.TriggerModeFn, config.Keyboards{Overrides: map[string]bool{"046d:c33f": false}}),
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

		all := caps(config.TriggerModeFn, config.Keyboards{ApplyTo: config.KeyboardsAll})
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

	t.Run("unknown key", func(t *testing.T) {
		remapper, _ := newRemapper(t, newStore())
		if err := remapper.Apply(config.Trigger{Key: "fn", Mode: config.TriggerModeFn}); err == nil {
			t.Error("Apply accepted an unknown key")
		}
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
	if !logitech.InScope || logitech.Ours == nil || *logitech.Ours != capsToGlobe {
		t.Errorf("logitech = %+v, want our caps lock mapping", logitech)
	}

	if !slices.Equal(logitech.Others, []Mapping{userMapping}) || logitech.Name != "Logitech" {
		t.Errorf("logitech = %+v, want the user's mapping once", logitech)
	}

	if statuses[0].InScope || statuses[0].Ours != nil {
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
