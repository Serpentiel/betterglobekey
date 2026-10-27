// Package inputmonitoring reports the macOS Input Monitoring permission that
// reading a keyboard's raw HID input requires.
package inputmonitoring

/*
#cgo LDFLAGS: -framework IOKit

#include <IOKit/hidsystem/IOHIDLib.h>

static int bgkInputMonitoringGranted(void) {
	return IOHIDCheckAccess(kIOHIDRequestTypeListenEvent) == kIOHIDAccessTypeGranted;
}
*/
import "C"

// Granted reports whether the process may read raw keyboard input.
func Granted() bool {
	return C.bgkInputMonitoringGranted() != 0
}
