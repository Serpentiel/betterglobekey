package keyboard

import "C"

// goHandleTriggerRelease is invoked from the C event-tap callback for each clean
// tap of the trigger key. It lives in its own file because cgo forbids C
// definitions in the preamble of a file that uses //export.
//
//export goHandleTriggerRelease
func goHandleTriggerRelease(reverse C.int) {
	if onPress != nil {
		onPress(reverse != 0)
	}
}
