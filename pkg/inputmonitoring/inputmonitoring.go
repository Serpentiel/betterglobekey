// Package inputmonitoring reports the macOS Input Monitoring permission that
// reading a keyboard's raw HID input requires.
package inputmonitoring

/*
#cgo LDFLAGS: -framework IOKit

#include <IOKit/hidsystem/IOHIDLib.h>

static int bgkInputMonitoringGranted(void) {
	return IOHIDCheckAccess(kIOHIDRequestTypeListenEvent) == kIOHIDAccessTypeGranted;
}

static int bgkInputMonitoringRequest(void) {
	return IOHIDRequestAccess(kIOHIDRequestTypeListenEvent);
}
*/
import "C"

// Granted reports whether the process may read raw keyboard input.
func Granted() bool {
	return C.bgkInputMonitoringGranted() != 0
}

// Request asks for Input Monitoring: macOS prompts the first time, and lists
// the process under Input Monitoring either way. It reports whether access is
// granted.
func Request() bool {
	return C.bgkInputMonitoringRequest() != 0
}
