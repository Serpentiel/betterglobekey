package cli

import (
	"errors"

	"github.com/Serpentiel/betterglobekey/internal/infra/keyboard"
	"github.com/Serpentiel/betterglobekey/internal/infra/keymap"
	"github.com/Serpentiel/betterglobekey/pkg/hidkeymap"
)

// watchDedicatedKeys points the event tap's raw input at the keys the remapper
// silenced for a dedicated trigger.
func watchDedicatedKeys(applied []keymap.Applied) {
	keys := make([]keyboard.DedicatedKey, 0, len(applied))

	for _, a := range applied {
		if a.Mapping.Dst != keymap.DedicatedUsage {
			continue
		}

		keys = append(keys, keyboard.DedicatedKey{
			Vendor:   a.Vendor,
			Product:  a.Product,
			Location: a.Location,
			Usage:    keymap.KeyUsage(a.Mapping.Src),
		})
	}

	keyboard.SetDedicatedKeys(keys)
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
