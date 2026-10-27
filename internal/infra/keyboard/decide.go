package keyboard

/*
#include "decide.h"
*/
import "C"

// eventType mirrors bgkEventType for tests, which cannot use cgo directly.
type eventType int

const (
	eventFlags   eventType = C.bgkEventFlags
	eventKeyDown eventType = C.bgkEventKeyDown
	eventKeyUp   eventType = C.bgkEventKeyUp
	eventMedia   eventType = C.bgkEventMedia
	eventOther   eventType = C.bgkEventOther
	eventReset   eventType = C.bgkEventReset
)

// tapEvent is a Go view of bgkEvent.
type tapEvent struct {
	kind       eventType
	keycode    int64
	flags      uint64
	autorepeat bool
}

// press is the outcome of feeding an event: whether a tap completed, and its direction.
type press struct {
	fired   bool
	reverse bool
}

// decider runs bgkDecide over a sequence of events for one trigger.
type decider struct {
	trigger     C.bgkTrigger
	reverseMask C.uint64_t
	state       C.bgkState
}

// newDecider returns a decider for the event tap, watching the Globe key or nothing.
func newDecider(globe bool, reverseModifier string) *decider {
	return &decider{trigger: tapTrigger(globe), reverseMask: modifierMasks[reverseModifier]}
}

func (d *decider) feed(event tapEvent) press {
	autorepeat := C.int(0)
	if event.autorepeat {
		autorepeat = 1
	}

	raw := C.bgkEvent{
		_type:      C.bgkEventType(event.kind),
		keycode:    C.int64_t(event.keycode),
		flags:      C.uint64_t(event.flags),
		autorepeat: autorepeat,
	}

	result := C.bgkDecide(&d.trigger, d.reverseMask, &d.state, &raw)
	if result == C.bgkNoPress {
		return press{}
	}

	return press{fired: true, reverse: result != 0}
}

// rawValue is a Go view of a raw HID value fed through bgkRawEvent.
type rawValue struct {
	page, usage uint32
	value       int64
	own         bool
	flags       uint64
}

// newRawDecider returns a decider for raw presses of the trigger key named key.
func newRawDecider(key, reverseModifier string) *decider {
	raw := rawKey(RawKey{Key: key})

	return &decider{
		trigger: C.bgkTrigger{
			kind:          C.bgkKindKey,
			keycode:       C.bgkRawTriggerCode,
			familyFlag:    raw.familyFlag,
			otherSideFlag: raw.otherSideFlag,
		},
		reverseMask: modifierMasks[reverseModifier],
	}
}

func (d *decider) feedRaw(value rawValue) press {
	own := C.int(0)
	if value.own {
		own = 1
	}

	raw := C.bgkRawEvent(
		C.uint32_t(value.page), C.uint32_t(value.usage), C.int64_t(value.value), own, C.uint64_t(value.flags),
	)

	result := C.bgkDecide(&d.trigger, d.reverseMask, &d.state, &raw)
	if result == C.bgkNoPress {
		return press{}
	}

	return press{fired: true, reverse: result != 0}
}
