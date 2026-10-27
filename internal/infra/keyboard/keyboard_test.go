package keyboard

import (
	"testing"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
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

func TestSetTrigger(_ *testing.T) {
	SetTrigger(config.Trigger{Key: "right_command", Mode: config.TriggerModePassThrough})
	SetTrigger(config.Trigger{Key: config.TriggerCapsLock, Mode: config.TriggerModeDedicated})
	SetTrigger(config.Trigger{Key: config.TriggerGlobe})
}
