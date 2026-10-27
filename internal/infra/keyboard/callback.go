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

// goRawInput is invoked from C when raw input is opened, fails to open, or is
// closed (see RawInputState).
//
//export goRawInput
func goRawInput(state, status C.int) {
	if RawInputState(state) == RawInputFailed {
		rawInputFailed(&RawInputError{Status: int32(status)})
	}

	if handler := rawInputHandler.Load(); handler != nil {
		(*handler)(RawInputState(state), int32(status))
	}
}

// goRecordedKey is invoked from C for each key pressed while recording.
//
//export goRecordedKey
func goRecordedKey(vendor, product, location, page, usage C.uint32_t) {
	recorded(RecordedKey{
		Vendor:   uint32(vendor),
		Product:  uint32(product),
		Location: uint32(location),
		Usage:    uint64(page)<<32 | uint64(usage),
	})
}
