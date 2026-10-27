package keyboard

/*
#include "decide.h"
*/
import "C"

import "github.com/Serpentiel/betterglobekey/internal/domain/config"

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

func newDecider(trigger config.Trigger, reverseModifier string) *decider {
	return &decider{trigger: triggerFor(trigger), reverseMask: modifierMasks[reverseModifier]}
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
