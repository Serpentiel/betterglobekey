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

// goRawInput is implemented in Go (see callback.go) and reports raw input being
// opened (state 1, status 0), failing to open (state 0, status the IOReturn),
// or closed (state -1).
extern void goRawInput(int state, int status);

// goRecordedKey is implemented in Go (see callback.go) and receives each key
// pressed while recording.
extern void goRecordedKey(uint32_t vendor, uint32_t product, uint32_t location, uint32_t page, uint32_t usage);

static CFMachPortRef bgkTap = NULL;
static CFRunLoopRef bgkRunLoop = NULL;
static CFRunLoopSourceRef bgkSource = NULL;

// bgkFnKeyCode is kVK_Function, the Globe key.
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

// bgkMaxRawKeys caps the trigger keys read from raw HID input.
#define bgkMaxRawKeys 64

// bgkRawKeys are the trigger keys read from raw input, at most one per
// keyboard. bgkRawState tracks their presses, and bgkRawTrigger is the
// descriptor raw presses are decided against, with the pressed key's flags.
static bgkRawKey bgkRawKeys[bgkMaxRawKeys];
static int bgkRawKeyCount = 0;
static bgkState bgkRawState = {0};
static const bgkTrigger bgkRawTrigger = {.kind = bgkKindKey, .keycode = bgkRawTriggerCode};

// bgkRecording reports keys pressed in raw input to goRecordedKey.
static int bgkRecording = 0;

// bgkHID reads raw keyboard input while there are raw keys or a recording,
// which needs Input Monitoring. It lives on the tap's run loop.
static IOHIDManagerRef bgkHID = NULL;
static int bgkHIDOpen = 0;

// bgkHIDSignal is a run-loop source on the tap's run loop that runs bgkSyncHID
// there when another thread changes the raw keys.
static CFRunLoopSourceRef bgkHIDSignal = NULL;

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

// bgkFindRawKey returns the raw trigger key with usage on the identified
// keyboard, or NULL. The caller holds bgkLock.
static const bgkRawKey *bgkFindRawKey(uint32_t vendor, uint32_t product, uint32_t location, uint32_t usage) {
	for (int i = 0; i < bgkRawKeyCount; i++) {
		const bgkRawKey *key = &bgkRawKeys[i];
		if (key->usage == usage && key->vendor == vendor && key->product == product && key->location == location) {
			return key;
		}
	}

	return NULL;
}

// bgkHIDValue handles raw input: presses of a raw trigger key, any other key or
// media key that makes such a press a combination, and keys being recorded.
static void bgkHIDValue(void *context, IOReturn status, void *sender, IOHIDValueRef value) {
	IOHIDElementRef element = IOHIDValueGetElement(value);
	IOHIDDeviceRef device = IOHIDElementGetDevice(element);
	uint32_t page = IOHIDElementGetUsagePage(element);
	uint32_t usage = IOHIDElementGetUsage(element);
	CFIndex pressed = IOHIDValueGetIntegerValue(value);
	uint64_t flags = CGEventSourceFlagsState(kCGEventSourceStateHIDSystemState);

	uint32_t vendor = 0, product = 0, location = 0;
	if (device != NULL) {
		vendor = bgkDeviceNumber(device, CFSTR(kIOHIDVendorIDKey));
		product = bgkDeviceNumber(device, CFSTR(kIOHIDProductIDKey));
		location = bgkDeviceNumber(device, CFSTR(kIOHIDLocationIDKey));
	}

	pthread_mutex_lock(&bgkLock);
	const bgkRawKey *key =
	    page == kHIDPage_KeyboardOrKeypad && device != NULL ? bgkFindRawKey(vendor, product, location, usage) : NULL;

	bgkTrigger trigger = bgkRawTrigger;
	if (key != NULL) {
		trigger.familyFlag = key->familyFlag;
		trigger.otherSideFlag = key->otherSideFlag;
	}

	bgkEvent event = bgkRawEvent(page, usage, pressed, key != NULL, flags);
	int reverse = bgkRawKeyCount > 0 ? bgkDecide(&trigger, bgkReverseMask, &bgkRawState, &event) : bgkNoPress;
	int record = bgkRecording && device != NULL && pressed != 0 && bgkIsRecordable(page, usage);
	pthread_mutex_unlock(&bgkLock);

	if (reverse != bgkNoPress) {
		goHandleTriggerRelease(reverse);
	}

	if (record) {
		goRecordedKey(vendor, product, location, page, usage);
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

// bgkSyncHID opens the HID manager while there are raw keys or a recording, and
// closes it otherwise. It runs on the tap's run loop, and retries an open that
// failed for want of Input Monitoring each time the raw keys are set.
static void bgkSyncHID(void) {
	pthread_mutex_lock(&bgkLock);
	int want = bgkRawKeyCount > 0 || bgkRecording;
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

		IOReturn status = IOHIDManagerOpen(bgkHID, kIOHIDOptionsTypeNone);
		bgkHIDOpen = status == kIOReturnSuccess;
		goRawInput(bgkHIDOpen, bgkHIDOpen ? 0 : status);
	} else if (!want && bgkHIDOpen) {
		IOHIDManagerClose(bgkHID, kIOHIDOptionsTypeNone);
		bgkHIDOpen = 0;
		goRawInput(-1, 0);
	}
}

// bgkHIDPerform is bgkHIDSignal's handler.
static void bgkHIDPerform(void *info) {
	bgkSyncHID();
}

// bgkSignalHID runs bgkSyncHID on the tap's run loop, once it runs.
static void bgkSignalHID(void) {
	pthread_mutex_lock(&bgkLock);
	pthread_mutex_unlock(&bgkLock);

	bgkSignalHID();
}

// bgkSetRecording turns reporting pressed keys to goRecordedKey on or off,
// opening raw input while it is on.
void bgkSetRecording(int on) {
	pthread_mutex_lock(&bgkLock);
	bgkRecording = on;
	pthread_mutex_unlock(&bgkLock);

	bgkSignalHID();
}

// bgkSetRawKeys sets the raw trigger keys, dropping a raw press in progress
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
	pthread_mutex_unlock(&bgkLock);

	bgkSignalHID();
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

	CFRunLoopSourceContext context = {.version = 0, .perform = bgkHIDPerform};
	CFRunLoopSourceRef signal = CFRunLoopSourceCreate(kCFAllocatorDefault, 0, &context);
	CFRunLoopAddSource(CFRunLoopGetCurrent(), signal, kCFRunLoopCommonModes);

	pthread_mutex_lock(&bgkLock);
	bgkRunLoop = CFRunLoopGetCurrent();
	bgkHIDSignal = signal;
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
