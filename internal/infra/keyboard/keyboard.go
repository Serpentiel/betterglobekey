// Package keyboard listens for Globe (fn) key presses using a CoreGraphics event
// tap, giving the application full control over detection without third-party
// dependencies. The Globe/fn key surfaces as a flags-changed event with key code
// kVK_Function (63); a release (the fn flag clearing) is treated as a press, and
// a held Shift marks it as a reverse press.
package keyboard

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>

extern void goHandleFnRelease(int reverse);

static CFMachPortRef bgkTap = NULL;
static CFRunLoopRef bgkRunLoop = NULL;
static CFRunLoopSourceRef bgkSource = NULL;

// bgkFnKeyCode is kVK_Function, the Globe/fn key.
static const int64_t bgkFnKeyCode = 63;

// fn key state, tracked for edge detection so we fire exactly once per press.
static bool bgkFnDown = false;
// bgkFnReverse latches whether the reverse modifier was held when fn was pressed.
static int bgkFnReverse = 0;
// bgkReverseMask is the modifier flag that marks a press as reverse (default Shift).
static CGEventFlags bgkReverseMask = kCGEventFlagMaskShift;

// bgkSetReverseMask sets the modifier flag used to detect a reverse press.
static void bgkSetReverseMask(CGEventFlags mask) {
	bgkReverseMask = mask;
}

static CGEventRef bgkCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
	// The system disables the tap on timeout or heavy input (e.g. across sleep);
	// re-enable it so the daemon keeps working.
	if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
		if (bgkTap != NULL) {
			CGEventTapEnable(bgkTap, true);
		}
		return event;
	}

	if (type == kCGEventFlagsChanged &&
	    CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode) == bgkFnKeyCode) {
		CGEventFlags flags = CGEventGetFlags(event);
		bool down = (flags & kCGEventFlagMaskSecondaryFn) != 0;

		if (down && !bgkFnDown) {
			// Press: latch the Shift state now (the natural "hold Shift, then tap").
			bgkFnDown = true;
			bgkFnReverse = (flags & bgkReverseMask) != 0 ? 1 : 0;
		} else if (!down && bgkFnDown) {
			// Release: fire once with the latched reverse flag.
			bgkFnDown = false;
			goHandleFnRelease(bgkFnReverse);
		}
		// Any other flags-changed event for this key code is a duplicate/no-op.
	}

	return event;
}

// bgkStart creates and enables the event tap on the current run loop. It returns
// 0 on failure, which typically means Accessibility permission is not granted.
static int bgkStart(void) {
	CGEventMask mask = CGEventMaskBit(kCGEventFlagsChanged);

	bgkTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap,
	                          kCGEventTapOptionListenOnly, mask, bgkCallback, NULL);
	if (bgkTap == NULL) {
		return 0;
	}

	bgkSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, bgkTap, 0);
	bgkRunLoop = CFRunLoopGetCurrent();
	CFRunLoopAddSource(bgkRunLoop, bgkSource, kCFRunLoopCommonModes);
	CGEventTapEnable(bgkTap, true);

	return 1;
}

static void bgkRun(void) {
	CFRunLoopRun();
}

static void bgkStop(void) {
	if (bgkRunLoop != NULL) {
		CFRunLoopStop(bgkRunLoop);
	}
}
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
