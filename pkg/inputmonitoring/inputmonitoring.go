// Package inputmonitoring obtains the macOS Input Monitoring access that
// reading a keyboard's raw HID input requires. macOS grants it without asking
// to a process that holds Accessibility, once the process requests it.
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

// Ensure requests Input Monitoring unless it is granted, and reports whether
// it is. It is granted at once with Accessibility, and refused without it.
func Ensure() bool {
	return Granted() || C.bgkInputMonitoringRequest() != 0
}
