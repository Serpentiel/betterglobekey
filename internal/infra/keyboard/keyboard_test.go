package keyboard

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestSetReverseModifier exercises the modifier-name lookup, including the
// fallback for unknown names. It only updates an internal event-flag mask, so it
// has no observable side effects and never installs the event tap.
func TestSetReverseModifier(_ *testing.T) {
	for _, name := range []string{"shift", "option", "control", "command", "unknown", ""} {
		SetReverseModifier(name)
	}
}

func TestNewReturnsListener(t *testing.T) {
	if New() == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetGlobe(_ *testing.T) {
	SetGlobe(false)
	SetGlobe(true)
}

func TestSetRawKeys(_ *testing.T) {
	// The listener is not running, so this only records the keys.
	SetRawKeys([]RawKey{{Vendor: 0x046d, Product: 0xc33f, Location: 1, Usage: 0xE5, Key: "right_shift"}})
	SetRawKeys(nil)
}

func TestRecord(t *testing.T) {
	logitech := func(key RecordedKey) bool { return key.Vendor == 0x046d }

	// The listener is not running, so keys arrive only through recorded.
	keys := make(chan RecordedKey)

	go func() {
		key, err := Record(t.Context(), logitech)
		if err != nil {
			t.Errorf("Record: %v", err)
		}

		keys <- key
	}()

	want := RecordedKey{Vendor: 0x046d, Product: 0xc33f, Location: 1, Usage: 0x7_0000_0049}

	for {
		recorded(RecordedKey{Vendor: 0x05ac, Usage: 0x7_0000_0004}) // another keyboard
		recorded(want)

		select {
		case got := <-keys:
			if got != want {
				t.Errorf("recorded %+v, want %+v", got, want)
			}

			return
		case <-time.After(time.Millisecond):
		}
	}
}

func TestRecordErrors(t *testing.T) {
	t.Run("raw input", func(t *testing.T) {
		errs := make(chan error)

		go func() {
			_, err := Record(t.Context(), func(RecordedKey) bool { return true })
			errs <- err
		}()

		failure := &RawInputError{Status: -0x1FFFFD1E}

		for {
			rawInputFailed(failure)

			select {
			case err := <-errs:
				if !errors.Is(err, failure) || err.Error() != "reading raw keyboard input failed (0xe00002e2)" {
					t.Errorf("err = %v, want %v", err, failure)
				}

				return
			case <-time.After(time.Millisecond):
			}
		}
	})

	t.Run("context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		if _, err := Record(ctx, func(RecordedKey) bool { return true }); !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want %v", err, context.Canceled)
		}
	})
}
