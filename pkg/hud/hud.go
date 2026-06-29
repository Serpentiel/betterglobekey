// Package hud shows a brief, system-bezel-style overlay naming the active input
// source. It owns an AppKit application that must run on the process's main
// thread; Show may be called from any goroutine.
// nolint:nlreturn // nlreturn misreports against cgo calls in this package.
package hud

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#include <stdlib.h>

void hudRun(void);
void hudStop(void);
void hudShow(const char *text, const char *subtitle);
void hudSetDuration(double seconds);
*/
import "C"

import (
	"time"
	"unsafe"
)

// Run starts the AppKit run loop and blocks until Stop is called. It must be
// invoked on the main thread (see runtime.LockOSThread in main).
func Run() {
	C.hudRun()
}

// Stop ends the AppKit run loop, unblocking Run.
func Stop() {
	C.hudStop()
}

// SetDuration sets how long the HUD stays fully visible before fading out.
func SetDuration(d time.Duration) {
	C.hudSetDuration(C.double(d.Seconds()))
}

// Show displays the given text and subtitle in the HUD, fading it out shortly
// after. An empty subtitle is omitted.
func Show(text, subtitle string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	cSubtitle := C.CString(subtitle)
	defer C.free(unsafe.Pointer(cSubtitle))

	C.hudShow(cText, cSubtitle)
}
