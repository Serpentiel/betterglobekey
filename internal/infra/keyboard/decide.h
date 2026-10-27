#ifndef BGK_DECIDE_H
#define BGK_DECIDE_H

#include <stdint.h>

// bgkKind is how a trigger key reports its presses.
typedef enum {
	// bgkKindFlag keys (Globe and modifiers) arrive as flags-changed events; the
	// key is down while its press flag is set.
	bgkKindFlag = 0,
	// bgkKindKey keys (F13-F20, Menu) arrive as key-down and key-up events.
	bgkKindKey = 1,
} bgkKind;

// bgkTrigger describes the key that switches input sources.
typedef struct {
	bgkKind kind;
	int64_t keycode;
	// pressFlag is the flag set while a bgkKindFlag key is down: the Secondary Fn
	// flag for Globe, or the key's own device (left/right) flag for a modifier.
	uint64_t pressFlag;
	// familyFlag is the generic modifier flag the key itself sets (e.g. Shift for
	// Right Shift), or 0. A press does not count as reverse because of it alone.
	uint64_t familyFlag;
	// otherSideFlag is the device flag of the same family's other key, which
	// still marks a press as reverse when familyFlag is the reverse modifier.
	uint64_t otherSideFlag;
} bgkTrigger;

// bgkEventType is the subset of event kinds the decision logic distinguishes.
typedef enum {
	bgkEventFlags = 0,
	bgkEventKeyDown = 1,
	bgkEventKeyUp = 2,
	// bgkEventMedia is a media key (volume, brightness, ...) going down.
	bgkEventMedia = 3,
	// bgkEventOther is anything else, such as other system-defined events.
	bgkEventOther = 4,
	// bgkEventReset clears the press state, e.g. after the tap was re-enabled.
	bgkEventReset = 5,
} bgkEventType;

typedef struct {
	bgkEventType type;
	int64_t keycode;
	uint64_t flags;
	int autorepeat;
} bgkEvent;

// bgkState tracks a press of the trigger key.
typedef struct {
	int down;
	// reverse latches whether the reverse modifier was held when the key went down.
	int reverse;
	// used records whether another key was pressed while the trigger was held, so
	// a combination (fn+F1, Right Command+C) is not treated as a standalone tap.
	int used;
} bgkState;

// bgkNoPress is returned by bgkDecide when the event does not complete a tap.
#define bgkNoPress (-1)

// bgkRawTriggerCode is the key code raw presses of a silenced trigger key are
// fed to bgkDecide as; no CoreGraphics key code is negative.
#define bgkRawTriggerCode (-2)

// bgkRawKey is a trigger key silenced on one keyboard: remapped to a usage that
// produces no event, so it is watched in the keyboard's raw HID input instead.
typedef struct {
	uint32_t vendor;
	uint32_t product;
	uint32_t location;
	// usage is the key's usage on the Keyboard/Keypad page.
	uint32_t usage;
} bgkRawKey;

// bgkRawEvent classifies a raw HID value for bgkDecide. own reports that it is a
// silenced trigger key; flags are the modifier flags at the time.
bgkEvent bgkRawEvent(uint32_t page, uint32_t usage, int64_t value, int own, uint64_t flags);

// bgkDecide feeds event into state and returns 1 for a reverse tap, 0 for a
// forward tap, or bgkNoPress.
int bgkDecide(const bgkTrigger *trigger, uint64_t reverseMask, bgkState *state, const bgkEvent *event);

#endif
