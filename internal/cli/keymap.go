package cli

import (
	"github.com/Serpentiel/betterglobekey/internal/infra/keymap"
	"github.com/Serpentiel/betterglobekey/pkg/hidkeymap"
)

// hidStore adapts pkg/hidkeymap to the remapper's Store port.
type hidStore struct{}

func (hidStore) Mappings() ([]keymap.Mapping, error) {
	mappings, err := hidkeymap.Mappings()
	if err != nil {
		return nil, err
	}

	converted := make([]keymap.Mapping, len(mappings))
	for i, mapping := range mappings {
		converted[i] = keymap.Mapping{Src: mapping.Src, Dst: mapping.Dst}
	}

	return converted, nil
}

func (hidStore) SetMappings(mappings []keymap.Mapping) error {
	converted := make([]hidkeymap.Mapping, len(mappings))
	for i, mapping := range mappings {
		converted[i] = hidkeymap.Mapping{Src: mapping.Src, Dst: mapping.Dst}
	}

	return hidkeymap.SetMappings(converted)
}
