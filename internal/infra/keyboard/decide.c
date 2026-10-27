#include "decide.h"

// bgkIsReverse reports whether flags hold the reverse modifier, ignoring the
// flag the trigger key sets itself.
static int bgkIsReverse(const bgkTrigger *trigger, uint64_t reverseMask, uint64_t flags) {
	if ((flags & reverseMask) == 0) {
		return 0;
	}

	if (trigger->familyFlag != 0 && trigger->familyFlag == reverseMask) {
		return (flags & trigger->otherSideFlag) != 0 ? 1 : 0;
	}

	return 1;
}

static void bgkPress(const bgkTrigger *trigger, uint64_t reverseMask, bgkState *state, uint64_t flags) {
	state->down = 1;
	state->reverse = bgkIsReverse(trigger, reverseMask, flags);
	state->used = 0;
}

static int bgkRelease(bgkState *state) {
	state->down = 0;

	return state->used ? bgkNoPress : state->reverse;
}

int bgkDecide(const bgkTrigger *trigger, uint64_t reverseMask, bgkState *state, const bgkEvent *event) {
	if (event->type == bgkEventReset) {
		*state = (bgkState){0};

		return bgkNoPress;
	}

	int own = event->keycode == trigger->keycode;

	if (trigger->kind == bgkKindFlag && own && event->type == bgkEventFlags) {
		int down = (event->flags & trigger->pressFlag) != 0;

		if (down && !state->down) {
			bgkPress(trigger, reverseMask, state, event->flags);
		} else if (!down && state->down) {
			return bgkRelease(state);
		}

		// Any other flags-changed event for this key is a duplicate.
		return bgkNoPress;
	}

	if (trigger->kind == bgkKindKey && own && (event->type == bgkEventKeyDown || event->type == bgkEventKeyUp)) {
		if (event->type == bgkEventKeyDown && !event->autorepeat && !state->down) {
			bgkPress(trigger, reverseMask, state, event->flags);
		} else if (event->type == bgkEventKeyUp && state->down) {
			return bgkRelease(state);
		}

		return bgkNoPress;
	}

	// A key or media key pressed while the trigger is held makes it a combination.
	// Plain modifiers arrive as flags-changed events and are intentionally ignored.
	if (state->down && (event->type == bgkEventKeyDown || event->type == bgkEventMedia)) {
		state->used = 1;
	}

	return bgkNoPress;
}

// HID usage pages and Keyboard/Keypad page usages bgkRawEvent distinguishes.
#define bgkPageKeyboard 0x07
#define bgkPageConsumer 0x0C
#define bgkUsageFirstKey 0x04
#define bgkUsageFirstModifier 0xE0

bgkEvent bgkRawEvent(uint32_t page, uint32_t usage, int64_t value, int own, uint64_t flags) {
	bgkEvent event = {.type = bgkEventOther, .keycode = -1, .flags = flags};

	if (page == bgkPageKeyboard && own) {
		event.type = value != 0 ? bgkEventKeyDown : bgkEventKeyUp;
		event.keycode = bgkRawTriggerCode;
	} else if (page == bgkPageKeyboard && usage >= bgkUsageFirstKey && usage < bgkUsageFirstModifier && value != 0) {
		// Any other key going down; modifiers, like flags-changed events, do not count.
		event.type = bgkEventKeyDown;
		event.keycode = usage;
	} else if (page == bgkPageConsumer && value != 0) {
		event.type = bgkEventMedia;
	}

	return event;
}
