package keyboard

import "C"

// goHandleFnRelease is invoked from the C event-tap callback for each Globe key
// release. It lives in its own file because cgo forbids C definitions in the
// preamble of a file that uses //export.
//
//export goHandleFnRelease
func goHandleFnRelease(reverse C.int) {
	if onPress != nil {
		onPress(reverse != 0)
	}
}
