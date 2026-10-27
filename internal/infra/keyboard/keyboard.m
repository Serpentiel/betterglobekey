#import <AppKit/AppKit.h>
#include <CoreGraphics/CoreGraphics.h>
#include <pthread.h>

#include "decide.h"

// goHandleTriggerRelease is implemented in Go (see callback.go) and invoked from
// the event-tap callback for each clean tap of the trigger key.
extern void goHandleTriggerRelease(int reverse);

static CFMachPortRef bgkTap = NULL;
static CFRunLoopRef bgkRunLoop = NULL;
static CFRunLoopSourceRef bgkSource = NULL;

// bgkFnKeyCode is kVK_Function, the Globe/fn key.
static const int64_t bgkFnKeyCode = 63;

// bgkSystemDefined is NSEventTypeSystemDefined, the type of media keys such as
// brightness and volume.
static const CGEventType bgkSystemDefined = 14;

// bgkAuxControlButtons is NX_SUBTYPE_AUX_CONTROL_BUTTONS, the media-key subtype.
static const short bgkAuxControlButtons = 8;

// bgkKeyStateDown is NX_KEYDOWN as carried in a media event's data1.
static const long bgkKeyStateDown = 0x0A;

// bgkLock guards the trigger, reverse mask, and press state, which the daemon
// reconfigures from another thread while the tap callback reads them.
static pthread_mutex_t bgkLock = PTHREAD_MUTEX_INITIALIZER;

// bgkTriggerKey is the key that switches input sources (default Globe).
static bgkTrigger bgkTriggerKey = {
    .kind = bgkKindFlag, .keycode = bgkFnKeyCode, .pressFlag = kCGEventFlagMaskSecondaryFn};

// bgkReverseMask is the modifier flag that marks a press as reverse (default Shift).
static uint64_t bgkReverseMask = kCGEventFlagMaskShift;

static bgkState bgkPressState = {0};

// bgkSetReverseMask sets the modifier flag used to detect a reverse press.
void bgkSetReverseMask(CGEventFlags mask) {
	pthread_mutex_lock(&bgkLock);
	bgkReverseMask = mask;
	pthread_mutex_unlock(&bgkLock);
}

// bgkSetTrigger sets the trigger key and drops any press in progress.
void bgkSetTrigger(bgkTrigger trigger) {
	pthread_mutex_lock(&bgkLock);
	bgkTriggerKey = trigger;
	bgkPressState = (bgkState){0};
	pthread_mutex_unlock(&bgkLock);
}

// bgkIsMediaKeyDown reports whether a system-defined event is a media key press.
static int bgkIsMediaKeyDown(CGEventRef event) {
	@autoreleasepool {
		NSEvent *nsEvent = [NSEvent eventWithCGEvent:event];
		if (nsEvent == nil || nsEvent.subtype != bgkAuxControlButtons) {
			return 0;
		}

		return ((nsEvent.data1 & 0xFF00) >> 8) == bgkKeyStateDown ? 1 : 0;
	}
}

static int bgkFeed(bgkEvent event) {
	pthread_mutex_lock(&bgkLock);
	int result = bgkDecide(&bgkTriggerKey, bgkReverseMask, &bgkPressState, &event);
	pthread_mutex_unlock(&bgkLock);

	return result;
}

static CGEventRef bgkCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
	// The system disables the tap on timeout or heavy input (e.g. across sleep);
	// re-enable it so the daemon keeps working, and forget a press whose release
	// may have been missed meanwhile.
	if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
		if (bgkTap != NULL) {
			CGEventTapEnable(bgkTap, true);
		}
		bgkFeed((bgkEvent){.type = bgkEventReset});
		return event;
	}

	bgkEvent decided = {
	    .keycode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode),
	    .flags = CGEventGetFlags(event),
	};

	switch (type) {
	case kCGEventFlagsChanged:
		decided.type = bgkEventFlags;
		break;
	case kCGEventKeyDown:
		decided.type = bgkEventKeyDown;
		decided.autorepeat = CGEventGetIntegerValueField(event, kCGKeyboardEventAutorepeat) != 0;
		break;
	case kCGEventKeyUp:
		decided.type = bgkEventKeyUp;
		break;
	default:
		decided.type = type == bgkSystemDefined && bgkIsMediaKeyDown(event) ? bgkEventMedia : bgkEventOther;
		decided.keycode = -1;
		break;
	}

	int reverse = bgkFeed(decided);
	if (reverse != bgkNoPress) {
		goHandleTriggerRelease(reverse);
	}

	return event;
}

// bgkStart creates and enables the event tap on the current run loop. It returns
// 0 on failure, which typically means Accessibility permission is not granted.
int bgkStart(void) {
	CGEventMask mask = CGEventMaskBit(kCGEventFlagsChanged) | CGEventMaskBit(kCGEventKeyDown) |
	                   CGEventMaskBit(kCGEventKeyUp) | CGEventMaskBit(bgkSystemDefined);

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
