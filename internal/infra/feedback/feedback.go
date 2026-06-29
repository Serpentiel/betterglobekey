// Package feedback shows the HUD overlay when the input source changes.
package feedback

import "github.com/Serpentiel/betterglobekey/pkg/hud"

// Namer resolves an input source ID to its localized display name.
type Namer interface {
	Name(id string) string
}

// HUD is a switcher notifier that displays the input source (with its collection
// as a subtitle) in the on-screen overlay.
type HUD struct {
	namer          Namer
	showCollection bool
}

// NewHUD returns a notifier that shows the HUD on each change. When
// showCollection is false the collection subtitle is omitted.
func NewHUD(namer Namer, showCollection bool) HUD {
	return HUD{namer: namer, showCollection: showCollection}
}

// Notify shows the localized source name, with the collection name as a subtitle
// when enabled.
func (h HUD) Notify(source, collection string) {
	label := h.namer.Name(source)
	if label == "" {
		label = source
	}

	if !h.showCollection {
		collection = ""
	}

	hud.Show(label, collection)
}
