// Package keyboard listens for Globe (fn) key presses using a CoreGraphics event
// tap, giving the application full control over detection without third-party
// dependencies. The Globe/fn key surfaces as a flags-changed event with key code
// kVK_Function (63); a release (the fn flag clearing) is treated as a press, and
// a held Shift marks it as a reverse press.
package keyboard

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>

// Implemented in keyboard.m.
void bgkSetReverseMask(CGEventFlags mask);
int bgkStart(void);
void bgkRun(void);
void bgkStop(void);
*/
import "C"

import "runtime"

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

// Listener delivers Globe (fn) key releases via a CoreGraphics event tap.
type Listener struct{}

// New returns a Listener.
func New() *Listener {
	return &Listener{}
}

// Listen installs the event tap and runs its run loop, invoking handler for each
// Globe key release (with reverse set when Shift is held). It blocks until Stop
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
