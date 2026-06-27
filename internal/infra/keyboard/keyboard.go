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
		// Fire on release: the fn flag is no longer set.
		if ((flags & kCGEventFlagMaskSecondaryFn) == 0) {
			goHandleFnRelease((flags & kCGEventFlagMaskShift) != 0 ? 1 : 0);
		}
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
