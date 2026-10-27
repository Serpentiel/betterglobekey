#include <IOKit/hidsystem/IOHIDEventSystemClient.h>

// bgkKeymapProperty is the HID event system property hidutil calls UserKeyMapping.
#define bgkKeymapProperty CFSTR("UserKeyMapping")
#define bgkKeymapSrc CFSTR("HIDKeyboardModifierMappingSrc")
#define bgkKeymapDst CFSTR("HIDKeyboardModifierMappingDst")

// bgkKeymapGet copies up to capacity system-wide key mappings into src and dst.
// It returns the total number of mappings, which may exceed capacity, or -1 if
// the HID event system is unavailable.
int bgkKeymapGet(uint64_t *src, uint64_t *dst, int capacity) {
	IOHIDEventSystemClientRef client = IOHIDEventSystemClientCreateSimpleClient(kCFAllocatorDefault);
	if (client == NULL) {
		return -1;
	}

	int count = 0;
	CFTypeRef value = IOHIDEventSystemClientCopyProperty(client, bgkKeymapProperty);

	if (value != NULL && CFGetTypeID(value) == CFArrayGetTypeID()) {
		CFArrayRef mappings = (CFArrayRef)value;

		for (CFIndex i = 0; i < CFArrayGetCount(mappings); i++) {
			CFTypeRef entry = CFArrayGetValueAtIndex(mappings, i);
			if (CFGetTypeID(entry) != CFDictionaryGetTypeID()) {
				continue;
			}

			CFTypeRef from = CFDictionaryGetValue((CFDictionaryRef)entry, bgkKeymapSrc);
			CFTypeRef to = CFDictionaryGetValue((CFDictionaryRef)entry, bgkKeymapDst);
			if (from == NULL || to == NULL || CFGetTypeID(from) != CFNumberGetTypeID() ||
			    CFGetTypeID(to) != CFNumberGetTypeID()) {
				continue;
			}

			if (count < capacity) {
				CFNumberGetValue((CFNumberRef)from, kCFNumberSInt64Type, &src[count]);
				CFNumberGetValue((CFNumberRef)to, kCFNumberSInt64Type, &dst[count]);
			}
			count++;
		}
	}

	if (value != NULL) {
		CFRelease(value);
	}
	CFRelease(client);

	return count;
}

// bgkKeymapSet replaces the system-wide key mappings with the count pairs in src
// and dst. It returns 1 on success and 0 on failure.
int bgkKeymapSet(const uint64_t *src, const uint64_t *dst, int count) {
	IOHIDEventSystemClientRef client = IOHIDEventSystemClientCreateSimpleClient(kCFAllocatorDefault);
	if (client == NULL) {
		return 0;
	}

	CFMutableArrayRef mappings = CFArrayCreateMutable(kCFAllocatorDefault, count, &kCFTypeArrayCallBacks);

	for (int i = 0; i < count; i++) {
		CFNumberRef from = CFNumberCreate(kCFAllocatorDefault, kCFNumberSInt64Type, &src[i]);
		CFNumberRef to = CFNumberCreate(kCFAllocatorDefault, kCFNumberSInt64Type, &dst[i]);
		const void *keys[] = {bgkKeymapSrc, bgkKeymapDst};
		const void *values[] = {from, to};
		CFDictionaryRef entry = CFDictionaryCreate(kCFAllocatorDefault, keys, values, 2, &kCFTypeDictionaryKeyCallBacks,
		                                           &kCFTypeDictionaryValueCallBacks);

		CFArrayAppendValue(mappings, entry);

		CFRelease(entry);
		CFRelease(to);
		CFRelease(from);
	}

	int ok = IOHIDEventSystemClientSetProperty(client, bgkKeymapProperty, mappings) ? 1 : 0;

	CFRelease(mappings);
	CFRelease(client);

	return ok;
}
