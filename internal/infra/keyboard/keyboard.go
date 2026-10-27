// Package keyboard listens for trigger key taps using a CoreGraphics event tap,
// giving the application full control over detection without third-party
// dependencies. The default trigger, the Globe/fn key, surfaces as a
// flags-changed event with key code kVK_Function (63); a release (the fn flag
// clearing) is treated as a press, and a held Shift marks it as a reverse press.
// A passed-through trigger key is watched the same way: modifiers through their
// device flag, other keys through key-down and key-up events. A dedicated
// trigger is remapped to F20 and watched as that key, leaving Globe alone.
package keyboard

/*
#cgo LDFLAGS: -framework AppKit -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>

#include "decide.h"

// Implemented in keyboard.m.
void bgkSetReverseMask(CGEventFlags mask);
void bgkSetTrigger(bgkTrigger trigger);
int bgkStart(void);
void bgkRun(void);
void bgkStop(void);
*/
import "C"

import (
	"runtime"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

// onPress is the active press handler. Only one Listener runs at a time.
var onPress func(reverse bool)

// modifierMasks maps configuration modifier names to CoreGraphics event flags.
var modifierMasks = map[string]C.CGEventFlags{
	"shift":   C.kCGEventFlagMaskShift,
	"option":  C.kCGEventFlagMaskAlternate,
	"control": C.kCGEventFlagMaskControl,
	"command": C.kCGEventFlagMaskCommand,
}

// SetReverseModifier sets which modifier key marks a press as reverse. Unknown
// names fall back to Shift. It is safe to call while listening; the event tap
// reads the value live.
func SetReverseModifier(modifier string) {
	mask, ok := modifierMasks[modifier]
	if !ok {
		mask = C.kCGEventFlagMaskShift
	}

	C.bgkSetReverseMask(mask)
}

// Device-dependent modifier flags (NX_DEVICE*KEYMASK), which tell the left and
// right keys of a modifier family apart.
const (
	deviceLeftControl  = 0x00000001
	deviceLeftShift    = 0x00000002
	deviceRightShift   = 0x00000004
	deviceLeftCommand  = 0x00000008
	deviceRightCommand = 0x00000010
	deviceLeftOption   = 0x00000020
	deviceRightOption  = 0x00000040
	deviceRightControl = 0x00002000
)

// Virtual key codes (kVK_* in Carbon's Events.h).
const (
	keyFunction       = 63
	keyCommand        = 55
	keyRightCommand   = 54
	keyOption         = 58
	keyRightOption    = 61
	keyControl        = 59
	keyRightControl   = 62
	keyShift          = 56
	keyRightShift     = 60
	keyF13            = 105
	keyF14            = 107
	keyF15            = 113
	keyF16            = 106
	keyF17            = 64
	keyF18            = 79
	keyF19            = 80
	keyF20            = 90
	keyContextualMenu = 110
)

// globeTrigger is the Globe (fn) key, also used for a key remapped to Globe.
var globeTrigger = C.bgkTrigger{kind: C.bgkKindFlag, keycode: keyFunction, pressFlag: C.kCGEventFlagMaskSecondaryFn}

// modifierTrigger describes a modifier key: its key code, its own device flag,
// its family flag, and the device flag of the family's other key.
func modifierTrigger(keycode int64, device, family, otherSide C.uint64_t) C.bgkTrigger {
	return C.bgkTrigger{
		kind:          C.bgkKindFlag,
		keycode:       C.int64_t(keycode),
		pressFlag:     device,
		familyFlag:    family,
		otherSideFlag: otherSide,
	}
}

// keyTrigger describes a regular key reported through key-down and key-up events.
func keyTrigger(keycode int64) C.bgkTrigger {
	return C.bgkTrigger{kind: C.bgkKindKey, keycode: C.int64_t(keycode)}
}

// passThroughTriggers maps the trigger keys that can be watched directly to their
// descriptors.
var passThroughTriggers = map[string]C.bgkTrigger{
	"right_command": modifierTrigger(keyRightCommand, deviceRightCommand, C.kCGEventFlagMaskCommand, deviceLeftCommand),
	"left_command":  modifierTrigger(keyCommand, deviceLeftCommand, C.kCGEventFlagMaskCommand, deviceRightCommand),
	"right_option":  modifierTrigger(keyRightOption, deviceRightOption, C.kCGEventFlagMaskAlternate, deviceLeftOption),
	"left_option":   modifierTrigger(keyOption, deviceLeftOption, C.kCGEventFlagMaskAlternate, deviceRightOption),
	"right_control": modifierTrigger(keyRightControl, deviceRightControl, C.kCGEventFlagMaskControl, deviceLeftControl),
	"left_control":  modifierTrigger(keyControl, deviceLeftControl, C.kCGEventFlagMaskControl, deviceRightControl),
	"right_shift":   modifierTrigger(keyRightShift, deviceRightShift, C.kCGEventFlagMaskShift, deviceLeftShift),
	"left_shift":    modifierTrigger(keyShift, deviceLeftShift, C.kCGEventFlagMaskShift, deviceRightShift),
	"f13":           keyTrigger(keyF13),
	"f14":           keyTrigger(keyF14),
	"f15":           keyTrigger(keyF15),
	"f16":           keyTrigger(keyF16),
	"f17":           keyTrigger(keyF17),
	"f18":           keyTrigger(keyF18),
	"f19":           keyTrigger(keyF19),
	"f20":           keyTrigger(keyF20),
	"menu":          keyTrigger(keyContextualMenu),
}

// dedicatedTrigger is the key a dedicated trigger is remapped to.
var dedicatedTrigger = keyTrigger(keyF20)

// triggerFor returns the descriptor the event tap watches for trigger: the key
// itself when it is passed through, the dedicated key when it is remapped to
// that, otherwise Globe (a key remapped to Globe arrives as Globe).
func triggerFor(trigger config.Trigger) C.bgkTrigger {
	switch {
	case trigger.Dedicated():
		return dedicatedTrigger
	case trigger.PassThrough():
		if descriptor, ok := passThroughTriggers[trigger.Key]; ok {
			return descriptor
		}
	}

	return globeTrigger
}

// SetTrigger sets which key the event tap watches, dropping any press in
// progress. It is safe to call while listening.
func SetTrigger(trigger config.Trigger) {
	C.bgkSetTrigger(triggerFor(trigger))
}

// Listener delivers trigger key releases via a CoreGraphics event tap.
type Listener struct{}

// New returns a Listener.
func New() *Listener {
	return &Listener{}
}

// Listen installs the event tap and runs its run loop, invoking handler for each
// trigger key release (with reverse set when Shift is held). It blocks until Stop
// is called. If the tap cannot be created (e.g. Accessibility is not granted), it
// returns immediately.
func (*Listener) Listen(handler func(reverse bool)) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	onPress = handler

	if C.bgkStart() == 0 {
		return
	}

	C.bgkRun()
}

// Stop stops the event tap's run loop, unblocking Listen.
func (*Listener) Stop() {
	C.bgkStop()
}
