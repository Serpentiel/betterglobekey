#include <CoreGraphics/CoreGraphics.h>

// goHandleFnRelease is implemented in Go (see callback.go) and invoked from the
// event-tap callback for each clean Globe key release.
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
// bgkFnUsed records whether another key was pressed while fn was held, so a
// modifier combo (e.g. fn+F1 for brightness) is not treated as a standalone tap.
static int bgkFnUsed = 0;
// bgkReverseMask is the modifier flag that marks a press as reverse (default Shift).
static CGEventFlags bgkReverseMask = kCGEventFlagMaskShift;

// bgkSystemDefined is NSEventTypeSystemDefined, the type of media keys such as
// brightness and volume (fn + F1-F20 on a standard-function-keys keyboard).
static const CGEventType bgkSystemDefined = 14;

// bgkSetReverseMask sets the modifier flag used to detect a reverse press.
void bgkSetReverseMask(CGEventFlags mask) {
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

	if (type == kCGEventFlagsChanged && CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode) == bgkFnKeyCode) {
		CGEventFlags flags = CGEventGetFlags(event);
		bool down = (flags & kCGEventFlagMaskSecondaryFn) != 0;

		if (down && !bgkFnDown) {
			// Press: latch the reverse modifier now (the natural "hold it, then tap").
			bgkFnDown = true;
			bgkFnReverse = (flags & bgkReverseMask) != 0 ? 1 : 0;
			bgkFnUsed = 0;
		} else if (!down && bgkFnDown) {
			// Release: fire only for a clean tap. If another key was pressed while
			// fn was held, it was used as a modifier (e.g. fn+F1) — do nothing.
			bgkFnDown = false;
			if (!bgkFnUsed) {
				goHandleFnRelease(bgkFnReverse);
			}
		}
		// Any other flags-changed event for this key code is a duplicate/no-op.
		return event;
	}

	// A key or media key pressed while fn is held means fn is acting as a
	// modifier, not a standalone tap. Plain modifiers (Shift, etc.) arrive as
	// flags-changed events and are intentionally ignored here.
	if (bgkFnDown && (type == kCGEventKeyDown || type == bgkSystemDefined)) {
		bgkFnUsed = 1;
	}

	return event;
}

// bgkStart creates and enables the event tap on the current run loop. It returns
// 0 on failure, which typically means Accessibility permission is not granted.
int bgkStart(void) {
	CGEventMask mask =
	    CGEventMaskBit(kCGEventFlagsChanged) | CGEventMaskBit(kCGEventKeyDown) | CGEventMaskBit(bgkSystemDefined);

	bgkTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap, kCGEventTapOptionListenOnly, mask, bgkCallback,
	                          NULL);
	if (bgkTap == NULL) {
		return 0;
	}

	bgkSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, bgkTap, 0);
	bgkRunLoop = CFRunLoopGetCurrent();
	CFRunLoopAddSource(bgkRunLoop, bgkSource, kCFRunLoopCommonModes);
	CGEventTapEnable(bgkTap, true);

	return 1;
}

void bgkRun(void) {
	CFRunLoopRun();
}

void bgkStop(void) {
	if (bgkRunLoop != NULL) {
		CFRunLoopStop(bgkRunLoop);
	}
}
