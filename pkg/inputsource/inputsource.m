#include <Carbon/Carbon.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

// cfStringCopyUTF8 returns a newly malloc'd UTF-8 C string for s (caller frees),
// or NULL if s is NULL or cannot be converted.
static char *cfStringCopyUTF8(CFStringRef s) {
	if (s == NULL) {
		return NULL;
	}

	CFIndex length = CFStringGetLength(s);
	CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;

	char *buffer = malloc(maxSize);
	if (buffer == NULL) {
		return NULL;
	}

	if (!CFStringGetCString(s, buffer, maxSize, kCFStringEncodingUTF8)) {
		free(buffer);
		return NULL;
	}

	return buffer;
}

// bgkNewFilter creates an empty mutable dictionary for filtering input sources.
static CFMutableDictionaryRef bgkNewFilter(void) {
	return CFDictionaryCreateMutable(kCFAllocatorDefault, 0, &kCFTypeDictionaryKeyCallBacks,
	                                 &kCFTypeDictionaryValueCallBacks);
}

// bgkCopyMatchingSources returns the input sources whose ID equals sourceID
// (caller releases), or NULL.
static CFArrayRef bgkCopyMatchingSources(const char *sourceID) {
	CFStringRef cfID = CFStringCreateWithCString(kCFAllocatorDefault, sourceID, kCFStringEncodingUTF8);
	if (cfID == NULL) {
		return NULL;
	}

	CFMutableDictionaryRef filter = bgkNewFilter();
	CFDictionarySetValue(filter, kTISPropertyInputSourceID, cfID);

	CFArrayRef list = TISCreateInputSourceList(filter, false);

	CFRelease(filter);
	CFRelease(cfID);

	return list;
}

// bgkAllInputSources returns the IDs of all enabled, selectable keyboard layouts
// and input modes as a newline-separated, malloc'd UTF-8 string (caller frees),
// or NULL.
char *bgkAllInputSources(void) {
	CFMutableDictionaryRef filter = bgkNewFilter();
	CFDictionarySetValue(filter, kTISPropertyInputSourceIsEnabled, kCFBooleanTrue);
	CFDictionarySetValue(filter, kTISPropertyInputSourceIsSelectCapable, kCFBooleanTrue);

	CFArrayRef list = TISCreateInputSourceList(filter, false);
	CFRelease(filter);
	if (list == NULL) {
		return NULL;
	}

	CFMutableStringRef joined = CFStringCreateMutable(kCFAllocatorDefault, 0);

	for (CFIndex i = 0, count = CFArrayGetCount(list); i < count; i++) {
		TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(list, i);

		CFStringRef type = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceType);
		if (CFStringCompare(type, kTISTypeKeyboardLayout, 0) != kCFCompareEqualTo &&
		    CFStringCompare(type, kTISTypeKeyboardInputMode, 0) != kCFCompareEqualTo) {
			continue;
		}

		CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
		if (sourceID == NULL) {
			continue;
		}

		if (CFStringGetLength(joined) > 0) {
			CFStringAppendCString(joined, "\n", kCFStringEncodingUTF8);
		}
		CFStringAppend(joined, sourceID);
	}

	CFRelease(list);

	char *result = cfStringCopyUTF8(joined);
	CFRelease(joined);

	return result;
}

// bgkCurrentInputSource returns the current keyboard input source's ID as a
// malloc'd UTF-8 string (caller frees), or NULL.
char *bgkCurrentInputSource(void) {
	TISInputSourceRef source = TISCopyCurrentKeyboardInputSource();
	if (source == NULL) {
		return NULL;
	}

	CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
	char *result = cfStringCopyUTF8(sourceID);

	CFRelease(source);

	return result;
}

// bgkInputSourceName returns the localized display name for the input source with
// the given ID as a malloc'd UTF-8 string (caller frees), or NULL if unavailable.
char *bgkInputSourceName(const char *sourceID) {
	CFArrayRef list = bgkCopyMatchingSources(sourceID);
	if (list == NULL) {
		return NULL;
	}

	char *result = NULL;
	if (CFArrayGetCount(list) > 0) {
		TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(list, 0);
		CFStringRef name = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyLocalizedName);
		result = cfStringCopyUTF8(name);
	}

	CFRelease(list);

	return result;
}

// bgkSelectInputSource selects the input source with the given ID, if present.
void bgkSelectInputSource(const char *sourceID) {
	CFArrayRef list = bgkCopyMatchingSources(sourceID);
	if (list == NULL) {
		return;
	}

	if (CFArrayGetCount(list) > 0) {
		TISSelectInputSource((TISInputSourceRef)CFArrayGetValueAtIndex(list, 0));
	}

	CFRelease(list);
}
