package cli

import (
	"context"
	"errors"
	"fmt"

	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	"github.com/Serpentiel/betterglobekey/internal/infra/control"
	"github.com/Serpentiel/betterglobekey/internal/infra/keyboard"
	"github.com/Serpentiel/betterglobekey/internal/infra/keymap"
	"github.com/Serpentiel/betterglobekey/internal/infra/logging"
	"github.com/Serpentiel/betterglobekey/pkg/hidkeymap"
)

// ioReturnNotPermitted is kIOReturnNotPermitted, what opening raw input fails
// with while Input Monitoring is not granted.
const ioReturnNotPermitted = 0xE00002E2

// logRawInput returns a raw input handler that logs it opening and closing, and
// each new reason it fails to open. Raw input is retried on every remapper sync,
// so a failure repeats until its cause is fixed.
func logRawInput(logger *logging.Logger) func(keyboard.RawInputState, int32) {
	var lastFailure uint32

	return func(state keyboard.RawInputState, status int32) {
		switch state {
		case keyboard.RawInputOpened:
			lastFailure = 0

			logger.Info("reading raw keyboard input for the dedicated trigger key")
		case keyboard.RawInputClosed:
			logger.Info("stopped reading raw keyboard input")
		case keyboard.RawInputFailed:
			code := uint32(status) //nolint:gosec // an IOReturn, reinterpreted as unsigned
			if code == lastFailure {
				return
			}

			lastFailure = code

			reason := "unknown"
			if code == ioReturnNotPermitted {
				reason = "Input Monitoring not granted"
			}

			logger.Error("reading raw keyboard input failed; the dedicated trigger key is ignored",
				"status", fmt.Sprintf("%#x", code), "reason", reason)
		}
	}
}

// watchTriggers points the event tap at every connected keyboard's trigger:
// the Globe key while any keyboard switches with it (its own, or a key
// remapped to fn), and raw input for dedicated and passed-through keys.
func watchTriggers(resolved []keymap.Resolved) {
	globe := len(resolved) == 0
	keys := make([]keyboard.RawKey, 0, len(resolved))

	for _, r := range resolved {
		trigger := r.Trigger

		switch {
		case trigger.Globe():
			globe = true
		case trigger.Fn():
			globe = globe || r.Applied
		case trigger.PassThrough() || (trigger.Dedicated() && r.Applied):
			keys = append(keys, keyboard.RawKey{
				Vendor:   r.Vendor,
				Product:  r.Product,
				Location: r.Location,
				Usage:    keymap.KeyUsage(keymap.Usages[trigger.Key]),
				Key:      trigger.Key,
			})
		}
	}

	keyboard.SetGlobe(globe)
	keyboard.SetRawKeys(keys)
}

// readsRawInput reports whether any trigger is read from raw input, which needs
// Input Monitoring.
func readsRawInput(trigger domainconfig.Trigger) bool {
	for _, keyboard := range trigger.Keyboards {
		if keyboard.Raw() {
			return true
		}
	}

	return false
}

// keyboardsPort adapts the remapper and raw key recording to the control
// server's Keyboards port.
type keyboardsPort struct {
	remapper *keymap.Remapper
}

func (p keyboardsPort) Keyboards(trigger domainconfig.Trigger) ([]keymap.Status, []keymap.Mapping, error) {
	return p.remapper.Keyboards(trigger)
}

func (keyboardsPort) Record(ctx context.Context, keyboardID string) (control.RecordedKey, error) {
	key, err := keyboard.Record(ctx, func(key keyboard.RecordedKey) bool {
		return keyboardID == "" || domainconfig.KeyboardID(key.Vendor, key.Product) == keyboardID
	})

	var rawErr *keyboard.RawInputError
	if errors.As(err, &rawErr) {
		code := uint32(rawErr.Status) //nolint:gosec // an IOReturn, reinterpreted as unsigned

		return control.RecordedKey{}, fmt.Errorf("%w (%#x)", control.ErrRawInput, code)
	}

	if err != nil {
		return control.RecordedKey{}, err
	}

	return control.RecordedKey{KeyboardID: domainconfig.KeyboardID(key.Vendor, key.Product), Usage: key.Usage}, nil
}

// hidStore adapts pkg/hidkeymap to the remapper's Store port.
type hidStore struct{}

func (hidStore) Keyboards() ([]keymap.Keyboard, error) {
	keyboards, err := hidkeymap.Keyboards()
	if err != nil {
		return nil, err
	}

	converted := make([]keymap.Keyboard, 0, len(keyboards))

	for _, keyboard := range keyboards {
		mappings, mappingsErr := hidkeymap.KeyboardMappings(keyboard.ID)
		if errors.Is(mappingsErr, hidkeymap.ErrNotFound) {
			continue // disconnected since it was listed
		}

		if mappingsErr != nil {
			return nil, mappingsErr
		}

		converted = append(converted, keymap.Keyboard{
			ID:        keyboard.ID,
			Vendor:    keyboard.Vendor,
			Product:   keyboard.Product,
			Location:  keyboard.Location,
			Name:      keyboard.Name,
			Transport: keyboard.Transport,
			BuiltIn:   keyboard.BuiltIn,
			Mappings:  fromHID(mappings),
		})
	}

	return converted, nil
}

func (hidStore) SetKeyboardMappings(id uint64, mappings []keymap.Mapping) error {
	return hidkeymap.SetKeyboardMappings(id, toHID(mappings))
}

func (hidStore) Mappings() ([]keymap.Mapping, error) {
	mappings, err := hidkeymap.Mappings()
	if err != nil {
		return nil, err
	}

	return fromHID(mappings), nil
}

func (hidStore) SetMappings(mappings []keymap.Mapping) error {
	return hidkeymap.SetMappings(toHID(mappings))
}

func fromHID(mappings []hidkeymap.Mapping) []keymap.Mapping {
	converted := make([]keymap.Mapping, len(mappings))
	for i, mapping := range mappings {
		converted[i] = keymap.Mapping{Src: mapping.Src, Dst: mapping.Dst}
	}

	return converted
}

func toHID(mappings []keymap.Mapping) []hidkeymap.Mapping {
	converted := make([]hidkeymap.Mapping, len(mappings))
	for i, mapping := range mappings {
		converted[i] = hidkeymap.Mapping{Src: mapping.Src, Dst: mapping.Dst}
	}

	return converted
}
