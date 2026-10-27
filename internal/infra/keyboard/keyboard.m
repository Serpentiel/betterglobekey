#import <AppKit/AppKit.h>
#include <CoreGraphics/CoreGraphics.h>
#include <IOKit/hid/IOHIDKeys.h>
#include <IOKit/hid/IOHIDManager.h>
#include <IOKit/hid/IOHIDUsageTables.h>
#include <IOKit/hidsystem/IOHIDLib.h>
#include <pthread.h>
#include <string.h>

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

// bgkMaxRawKeys caps the silenced trigger keys watched in raw HID input.
#define bgkMaxRawKeys 64

// bgkRawKeys are the silenced trigger keys, one per keyboard; none outside
// dedicated mode. bgkRawState tracks their presses, and bgkRawTrigger is the
// descriptor raw presses are decided against.
static bgkRawKey bgkRawKeys[bgkMaxRawKeys];
static int bgkRawKeyCount = 0;
static bgkState bgkRawState = {0};
static const bgkTrigger bgkRawTrigger = {.kind = bgkKindKey, .keycode = bgkRawTriggerCode};

// bgkHID reads raw keyboard input while there are raw keys, which needs Input
// Monitoring. It lives on the tap's run loop.
static IOHIDManagerRef bgkHID = NULL;
static int bgkHIDOpen = 0;

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

// bgkDeviceNumber reads an integer property of a HID device, or 0.
static uint32_t bgkDeviceNumber(IOHIDDeviceRef device, CFStringRef key) {
	int64_t number = 0;
	CFTypeRef value = IOHIDDeviceGetProperty(device, key);

	if (value != NULL && CFGetTypeID(value) == CFNumberGetTypeID()) {
		CFNumberGetValue((CFNumberRef)value, kCFNumberSInt64Type, &number);
	}

	return (uint32_t)number;
}

// bgkIsRawKey reports whether usage on device is a silenced trigger key. The
// caller holds bgkLock.
static int bgkIsRawKey(IOHIDDeviceRef device, uint32_t usage) {
	if (device == NULL || bgkRawKeyCount == 0) {
		return 0;
	}

	uint32_t vendor = bgkDeviceNumber(device, CFSTR(kIOHIDVendorIDKey));
	uint32_t product = bgkDeviceNumber(device, CFSTR(kIOHIDProductIDKey));
	uint32_t location = bgkDeviceNumber(device, CFSTR(kIOHIDLocationIDKey));

	for (int i = 0; i < bgkRawKeyCount; i++) {
		const bgkRawKey *key = &bgkRawKeys[i];
		if (key->usage == usage && key->vendor == vendor && key->product == product && key->location == location) {
			return 1;
		}
	}

	return 0;
}

// bgkHIDValue handles raw input: presses of a silenced trigger key, and any
// other key or media key that makes such a press a combination.
static void bgkHIDValue(void *context, IOReturn status, void *sender, IOHIDValueRef value) {
	IOHIDElementRef element = IOHIDValueGetElement(value);
	uint32_t page = IOHIDElementGetUsagePage(element);
	uint32_t usage = IOHIDElementGetUsage(element);
	CFIndex pressed = IOHIDValueGetIntegerValue(value);
	uint64_t flags = CGEventSourceFlagsState(kCGEventSourceStateHIDSystemState);

	pthread_mutex_lock(&bgkLock);
	int own = page == kHIDPage_KeyboardOrKeypad && bgkIsRawKey(IOHIDElementGetDevice(element), usage);
	bgkEvent event = bgkRawEvent(page, usage, pressed, own, flags);
	int reverse = bgkRawKeyCount > 0 ? bgkDecide(&bgkRawTrigger, bgkReverseMask, &bgkRawState, &event) : bgkNoPress;
	pthread_mutex_unlock(&bgkLock);

	if (reverse != bgkNoPress) {
		goHandleTriggerRelease(reverse);
	}
}

// bgkMatching returns a matching dictionary for devices of a usage page and usage.
static CFDictionaryRef bgkMatching(int page, int usage) {
	CFNumberRef pageNumber = CFNumberCreate(kCFAllocatorDefault, kCFNumberIntType, &page);
	CFNumberRef usageNumber = CFNumberCreate(kCFAllocatorDefault, kCFNumberIntType, &usage);
	const void *keys[] = {CFSTR(kIOHIDDeviceUsagePageKey), CFSTR(kIOHIDDeviceUsageKey)};
	const void *values[] = {pageNumber, usageNumber};
	CFDictionaryRef matching = CFDictionaryCreate(kCFAllocatorDefault, keys, values, 2, &kCFTypeDictionaryKeyCallBacks,
	                                              &kCFTypeDictionaryValueCallBacks);

	CFRelease(usageNumber);
	CFRelease(pageNumber);

	return matching;
}

// bgkSyncHID opens the HID manager while there are raw keys and closes it
// otherwise. It runs on the tap's run loop, and retries an open that failed for
// want of Input Monitoring each time the raw keys are set.
static void bgkSyncHID(void) {
	pthread_mutex_lock(&bgkLock);
	int want = bgkRawKeyCount > 0;
	pthread_mutex_unlock(&bgkLock);

	if (want && !bgkHIDOpen) {
		if (bgkHID == NULL) {
			bgkHID = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);

			CFDictionaryRef keyboards = bgkMatching(kHIDPage_GenericDesktop, kHIDUsage_GD_Keyboard);
			CFDictionaryRef consumer = bgkMatching(kHIDPage_Consumer, kHIDUsage_Csmr_ConsumerControl);
			const void *matches[] = {keyboards, consumer};
			CFArrayRef matching = CFArrayCreate(kCFAllocatorDefault, matches, 2, &kCFTypeArrayCallBacks);

			IOHIDManagerSetDeviceMatchingMultiple(bgkHID, matching);
			IOHIDManagerRegisterInputValueCallback(bgkHID, bgkHIDValue, NULL);
			IOHIDManagerScheduleWithRunLoop(bgkHID, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);

			CFRelease(matching);
			CFRelease(consumer);
			CFRelease(keyboards);
		}

		// Shows the Input Monitoring prompt the first time; a no-op after.
		IOHIDRequestAccess(kIOHIDRequestTypeListenEvent);
		bgkHIDOpen = IOHIDManagerOpen(bgkHID, kIOHIDOptionsTypeNone) == kIOReturnSuccess;
	} else if (!want && bgkHIDOpen) {
		IOHIDManagerClose(bgkHID, kIOHIDOptionsTypeNone);
		bgkHIDOpen = 0;
	}
}

// bgkSetRawKeys sets the silenced trigger keys, dropping a raw press in progress
// when they change, and opens or closes raw input to match.
void bgkSetRawKeys(const bgkRawKey *keys, int count) {
	if (keys == NULL || count < 0) {
		count = 0;
	} else if (count > bgkMaxRawKeys) {
		count = bgkMaxRawKeys;
	}

	pthread_mutex_lock(&bgkLock);
	if (count != bgkRawKeyCount || (count > 0 && memcmp(keys, bgkRawKeys, sizeof(bgkRawKey) * count) != 0)) {
		if (count > 0) {
			memcpy(bgkRawKeys, keys, sizeof(bgkRawKey) * count);
		}
		bgkRawKeyCount = count;
		bgkRawState = (bgkState){0};
	}
	CFRunLoopRef runLoop = bgkRunLoop;
	pthread_mutex_unlock(&bgkLock);

	if (runLoop != NULL) {
		CFRunLoopPerformBlock(runLoop, kCFRunLoopCommonModes, ^{
		  bgkSyncHID();
		});
		CFRunLoopWakeUp(runLoop);
	}
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

		pthread_mutex_lock(&bgkLock);
		bgkRawState = (bgkState){0};
		pthread_mutex_unlock(&bgkLock);

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

	pthread_mutex_lock(&bgkLock);
	bgkRunLoop = CFRunLoopGetCurrent();
	pthread_mutex_unlock(&bgkLock);

	CFRunLoopAddSource(bgkRunLoop, bgkSource, kCFRunLoopCommonModes);
	CGEventTapEnable(bgkTap, true);
	bgkSyncHID();

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
