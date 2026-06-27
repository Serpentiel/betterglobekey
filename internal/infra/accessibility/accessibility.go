// Package accessibility reports and requests the macOS Accessibility permission
// that the global event tap requires.
package accessibility

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include <ApplicationServices/ApplicationServices.h>

static int bgkTrusted(void) {
	return AXIsProcessTrusted() ? 1 : 0;
}

static int bgkPrompt(void) {
	const void *keys[] = { kAXTrustedCheckOptionPrompt };
	const void *values[] = { kCFBooleanTrue };
	CFDictionaryRef options = CFDictionaryCreate(
		NULL, keys, values, 1,
		&kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);

	int trusted = AXIsProcessTrustedWithOptions(options) ? 1 : 0;

	CFRelease(options);

	return trusted;
}
*/
import "C"

// Trusted reports whether the process is trusted for Accessibility, which is
// required for the global keyboard event tap to receive events.
func Trusted() bool {
	return C.bgkTrusted() != 0
}

// Prompt reports whether the process is trusted, presenting the system
// Accessibility prompt if it is not yet granted.
func Prompt() bool {
	return C.bgkPrompt() != 0
}
