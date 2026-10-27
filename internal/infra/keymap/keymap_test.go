package keymap

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

// fakeStore is an in-memory mapping table.
type fakeStore struct {
	mappings []Mapping
	writes   int
	readErr  error
	writeErr error
}

func (s *fakeStore) Mappings() ([]Mapping, error) {
	return slices.Clone(s.mappings), s.readErr
}

func (s *fakeStore) SetMappings(mappings []Mapping) error {
	if s.writeErr != nil {
		return s.writeErr
	}

	s.mappings = slices.Clone(mappings)
	s.writes++

	return nil
}

var (
	capsLock     = Usages[config.TriggerCapsLock]
	rightCommand = Usages["right_command"]
	userMapping  = Mapping{Src: keyboardPage | 0x04, Dst: keyboardPage | 0x05} // a -> b
)

func newRemapper(t *testing.T, store *fakeStore) (*Remapper, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "keymap.json")

	return New(store, path), path
}

func assertMappings(t *testing.T, store *fakeStore, want ...Mapping) {
	t.Helper()

	if !slices.Equal(store.mappings, want) {
		t.Errorf("mappings = %+v, want %+v", store.mappings, want)
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

func TestApplyAddsMappingAndKeepsOthers(t *testing.T) {
	store := &fakeStore{mappings: []Mapping{userMapping}}
	remapper, path := newRemapper(t, store)

	if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	assertMappings(t, store, userMapping, Mapping{Src: capsLock, Dst: GlobeUsage})

	if _, err := os.Stat(path); err != nil {
		t.Errorf("state file not written: %v", err)
	}
}

func TestApplyIsIdempotent(t *testing.T) {
	store := &fakeStore{}
	remapper, _ := newRemapper(t, store)

	for range 3 {
		if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	}

	assertMappings(t, store, Mapping{Src: capsLock, Dst: GlobeUsage})

	if store.writes != 1 {
		t.Errorf("writes = %d, want 1", store.writes)
	}
}

func TestApplyReaddsDroppedMapping(t *testing.T) {
	store := &fakeStore{}
	remapper, _ := newRemapper(t, store)

	if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	store.mappings = nil // the system forgot it, e.g. across a restart

	if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	assertMappings(t, store, Mapping{Src: capsLock, Dst: GlobeUsage})
}

func TestApplyReplacesAndRestoresSameKeyMapping(t *testing.T) {
	theirs := Mapping{Src: capsLock, Dst: keyboardPage | 0xE0} // caps lock -> left control
	store := &fakeStore{mappings: []Mapping{theirs, userMapping}}
	remapper, path := newRemapper(t, store)

	if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	assertMappings(t, store, userMapping, Mapping{Src: capsLock, Dst: GlobeUsage})

	// Re-applying must not lose the displaced entry.
	if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if err := remapper.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	assertMappings(t, store, userMapping, theirs)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("state file left behind: %v", err)
	}
}

func TestApplySwitchesKeys(t *testing.T) {
	store := &fakeStore{mappings: []Mapping{userMapping}}
	remapper, _ := newRemapper(t, store)

	if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if err := remapper.Apply(config.Trigger{Key: "right_command"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	assertMappings(t, store, userMapping, Mapping{Src: rightCommand, Dst: GlobeUsage})
}

func TestApplyWithoutRemapRemovesEntry(t *testing.T) {
	for name, trigger := range map[string]config.Trigger{
		"globe":        {Key: config.TriggerGlobe},
		"pass-through": {Key: "right_command", PassThrough: true},
	} {
		t.Run(name, func(t *testing.T) {
			store := &fakeStore{mappings: []Mapping{userMapping}}
			remapper, _ := newRemapper(t, store)

			if err := remapper.Apply(config.Trigger{Key: "right_command"}); err != nil {
				t.Fatalf("Apply: %v", err)
			}

			if err := remapper.Apply(trigger); err != nil {
				t.Fatalf("Apply: %v", err)
			}

			assertMappings(t, store, userMapping)
		})
	}
}

func TestNewCleansUpAfterCrash(t *testing.T) {
	store := &fakeStore{mappings: []Mapping{userMapping}}
	path := filepath.Join(t.TempDir(), "keymap.json")

	if err := New(store, path).Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// A fresh remapper (the next run) finds the leftover entry and removes it.
	if err := New(store, path).Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	assertMappings(t, store, userMapping)
}

func TestRestoreLeavesChangedEntryAlone(t *testing.T) {
	store := &fakeStore{}
	remapper, _ := newRemapper(t, store)

	if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	edited := Mapping{Src: capsLock, Dst: keyboardPage | 0x29} // the user remapped it to Escape meanwhile
	store.mappings = []Mapping{edited}

	if err := remapper.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	assertMappings(t, store, edited)
}

func TestRestoreWithoutStateIsNoOp(t *testing.T) {
	store := &fakeStore{mappings: []Mapping{userMapping}}
	remapper, _ := newRemapper(t, store)

	if err := remapper.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	assertMappings(t, store, userMapping)

	if store.writes != 0 {
		t.Errorf("writes = %d, want 0", store.writes)
	}
}

func TestApplyErrors(t *testing.T) {
	boom := errors.New("boom")

	t.Run("read", func(t *testing.T) {
		remapper, _ := newRemapper(t, &fakeStore{readErr: boom})
		if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); !errors.Is(err, boom) {
			t.Errorf("err = %v, want %v", err, boom)
		}
	})

	t.Run("write keeps state for cleanup", func(t *testing.T) {
		store := &fakeStore{writeErr: boom}
		remapper, path := newRemapper(t, store)

		if err := remapper.Apply(config.Trigger{Key: config.TriggerCapsLock}); !errors.Is(err, boom) {
			t.Errorf("err = %v, want %v", err, boom)
		}

		if _, err := os.Stat(path); err != nil {
			t.Errorf("state file missing after failed write: %v", err)
		}

		store.writeErr = nil

		if err := remapper.Restore(); err != nil {
			t.Fatalf("Restore: %v", err)
		}

		assertMappings(t, store)
	})

	t.Run("unknown key", func(t *testing.T) {
		remapper, _ := newRemapper(t, &fakeStore{})
		if err := remapper.Apply(config.Trigger{Key: "fn"}); err == nil {
			t.Error("Apply accepted an unknown key")
		}
	})

	t.Run("corrupt state", func(t *testing.T) {
		remapper, path := newRemapper(t, &fakeStore{})
		if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := remapper.Restore(); err == nil {
			t.Error("Restore accepted a corrupt state file")
		}
	})
}
