// Package feedback shows the HUD overlay when the input source changes.
package feedback

import "github.com/Serpentiel/betterglobekey/internal/infra/hud"

// Namer resolves an input source ID to its localized display name.
type Namer interface {
	Name(id string) string
}

// HUD is a switcher notifier that displays the input source (with its collection
// as a subtitle) in the on-screen overlay.
type HUD struct {
	namer Namer
}

// NewHUD returns a notifier that shows the HUD on each change.
func NewHUD(namer Namer) HUD {
	return HUD{namer: namer}
}

// Notify shows the localized source name with the collection name as a subtitle.
func (h HUD) Notify(source, collection string) {
	label := h.namer.Name(source)
	if label == "" {
		label = source
	}

	hud.Show(label, collection)
}
