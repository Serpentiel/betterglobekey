#include <ApplicationServices/ApplicationServices.h>

// bgkTrusted reports whether the process is trusted for Accessibility, which is
// required for the global keyboard event tap to receive events.
int bgkTrusted(void) {
	return AXIsProcessTrusted() ? 1 : 0;
}

// bgkPrompt reports whether the process is trusted, presenting the system
// Accessibility prompt if it is not yet granted.
int bgkPrompt(void) {
	const void *keys[] = {kAXTrustedCheckOptionPrompt};
	const void *values[] = {kCFBooleanTrue};
	CFDictionaryRef options =
	    CFDictionaryCreate(NULL, keys, values, 1, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);

	int trusted = AXIsProcessTrustedWithOptions(options) ? 1 : 0;

	CFRelease(options);

	return trusted;
}
