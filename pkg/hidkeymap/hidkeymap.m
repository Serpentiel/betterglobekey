#include "hidkeymap.h"

#include <IOKit/hid/IOHIDKeys.h>
#include <IOKit/hid/IOHIDUsageTables.h>
#include <IOKit/hidsystem/IOHIDEventSystemClient.h>
#include <IOKit/hidsystem/IOHIDServiceClient.h>

// bgkKeymapProperty is the HID property hidutil calls UserKeyMapping.
#define bgkKeymapProperty CFSTR("UserKeyMapping")
#define bgkKeymapSrc CFSTR("HIDKeyboardModifierMappingSrc")
#define bgkKeymapDst CFSTR("HIDKeyboardModifierMappingDst")

// bgkReadMappings copies up to capacity mappings from a UserKeyMapping value into
// src and dst, returning the total number of mappings, which may exceed capacity.
static int bgkReadMappings(CFTypeRef value, uint64_t *src, uint64_t *dst, int capacity) {
	if (value == NULL || CFGetTypeID(value) != CFArrayGetTypeID()) {
		return 0;
	}

	int count = 0;
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

	return count;
}

// bgkBuildMappings returns a new UserKeyMapping value holding the count pairs in
// src and dst. The caller releases it.
static CFArrayRef bgkBuildMappings(const uint64_t *src, const uint64_t *dst, int count) {
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

	return mappings;
}

// bgkKeymapGet copies up to capacity system-wide key mappings into src and dst.
// It returns the total number of mappings, which may exceed capacity, or
// bgkUnavailable if the HID event system is unavailable.
int bgkKeymapGet(uint64_t *src, uint64_t *dst, int capacity) {
	IOHIDEventSystemClientRef client = IOHIDEventSystemClientCreateSimpleClient(kCFAllocatorDefault);
	if (client == NULL) {
		return bgkUnavailable;
	}

	CFTypeRef value = IOHIDEventSystemClientCopyProperty(client, bgkKeymapProperty);
	int count = bgkReadMappings(value, src, dst, capacity);

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

	CFArrayRef mappings = bgkBuildMappings(src, dst, count);
	int ok = IOHIDEventSystemClientSetProperty(client, bgkKeymapProperty, mappings) ? 1 : 0;

	CFRelease(mappings);
	CFRelease(client);

	return ok;
}

// bgkNumber reads a service property as an unsigned integer, or 0.
static uint64_t bgkNumber(IOHIDServiceClientRef service, CFStringRef key) {
	uint64_t result = 0;
	CFTypeRef value = IOHIDServiceClientCopyProperty(service, key);

	if (value != NULL && CFGetTypeID(value) == CFNumberGetTypeID()) {
		CFNumberGetValue((CFNumberRef)value, kCFNumberSInt64Type, &result);
	} else if (value != NULL && CFGetTypeID(value) == CFBooleanGetTypeID()) {
		result = CFBooleanGetValue((CFBooleanRef)value) ? 1 : 0;
	}

	if (value != NULL) {
		CFRelease(value);
	}

	return result;
}

// bgkString copies a service string property into out, or an empty string.
static void bgkString(IOHIDServiceClientRef service, CFStringRef key, char *out, CFIndex size) {
	out[0] = '\0';
	CFTypeRef value = IOHIDServiceClientCopyProperty(service, key);

	if (value != NULL && CFGetTypeID(value) == CFStringGetTypeID()) {
		if (!CFStringGetCString((CFStringRef)value, out, size, kCFStringEncodingUTF8)) {
			out[0] = '\0';
		}
	}

	if (value != NULL) {
		CFRelease(value);
	}
}

// bgkRegistryID returns a service's registry entry ID, or 0.
static uint64_t bgkRegistryID(IOHIDServiceClientRef service) {
	uint64_t id = 0;
	CFTypeRef value = IOHIDServiceClientGetRegistryID(service);

	if (value != NULL && CFGetTypeID(value) == CFNumberGetTypeID()) {
		CFNumberGetValue((CFNumberRef)value, kCFNumberSInt64Type, &id);
	}

	return id;
}

// bgkKeyboardServices returns the connected keyboard services, or NULL. The
// caller releases the array.
static CFArrayRef bgkKeyboardServices(IOHIDEventSystemClientRef client) {
	CFArrayRef services = IOHIDEventSystemClientCopyServices(client);
	if (services == NULL) {
		return NULL;
	}

	CFMutableArrayRef keyboards = CFArrayCreateMutable(kCFAllocatorDefault, 0, &kCFTypeArrayCallBacks);

	for (CFIndex i = 0; i < CFArrayGetCount(services); i++) {
		IOHIDServiceClientRef service = (IOHIDServiceClientRef)CFArrayGetValueAtIndex(services, i);
		if (IOHIDServiceClientConformsTo(service, kHIDPage_GenericDesktop, kHIDUsage_GD_Keyboard)) {
			CFArrayAppendValue(keyboards, service);
		}
	}

	CFRelease(services);

	return keyboards;
}

// bgkKeyboardsList fills up to capacity entries of out with the connected
// keyboards. It returns the total number of keyboards, which may exceed
// capacity, or bgkUnavailable.
int bgkKeyboardsList(bgkKeyboard *out, int capacity) {
	IOHIDEventSystemClientRef client = IOHIDEventSystemClientCreateSimpleClient(kCFAllocatorDefault);
	if (client == NULL) {
		return bgkUnavailable;
	}

	CFArrayRef keyboards = bgkKeyboardServices(client);
	if (keyboards == NULL) {
		CFRelease(client);
		return bgkUnavailable;
	}

	int count = (int)CFArrayGetCount(keyboards);

	for (int i = 0; i < count && i < capacity; i++) {
		IOHIDServiceClientRef service = (IOHIDServiceClientRef)CFArrayGetValueAtIndex(keyboards, i);
		bgkKeyboard *keyboard = &out[i];

		keyboard->id = bgkRegistryID(service);
		keyboard->vendor = (uint32_t)bgkNumber(service, CFSTR(kIOHIDVendorIDKey));
		keyboard->product = (uint32_t)bgkNumber(service, CFSTR(kIOHIDProductIDKey));
		keyboard->location = (uint32_t)bgkNumber(service, CFSTR(kIOHIDLocationIDKey));
		keyboard->builtIn = bgkNumber(service, CFSTR(kIOHIDBuiltInKey)) != 0;
		bgkString(service, CFSTR(kIOHIDProductKey), keyboard->name, sizeof(keyboard->name));
		bgkString(service, CFSTR(kIOHIDTransportKey), keyboard->transport, sizeof(keyboard->transport));
	}

	CFRelease(keyboards);
	CFRelease(client);

	return count;
}

// bgkFindKeyboard returns the keyboard service with registry ID id from
// keyboards, or NULL.
static IOHIDServiceClientRef bgkFindKeyboard(CFArrayRef keyboards, uint64_t id) {
	for (CFIndex i = 0; i < CFArrayGetCount(keyboards); i++) {
		IOHIDServiceClientRef service = (IOHIDServiceClientRef)CFArrayGetValueAtIndex(keyboards, i);
		if (bgkRegistryID(service) == id) {
			return service;
		}
	}

	return NULL;
}

// bgkKeyboardKeymapGet copies up to capacity of a keyboard's key mappings into
// src and dst. It returns the total number of mappings, which may exceed
// capacity, bgkUnavailable, or bgkNotFound.
int bgkKeyboardKeymapGet(uint64_t id, uint64_t *src, uint64_t *dst, int capacity) {
	IOHIDEventSystemClientRef client = IOHIDEventSystemClientCreateSimpleClient(kCFAllocatorDefault);
	if (client == NULL) {
		return bgkUnavailable;
	}

	CFArrayRef keyboards = bgkKeyboardServices(client);
	if (keyboards == NULL) {
		CFRelease(client);
		return bgkUnavailable;
	}

	int count = bgkNotFound;
	IOHIDServiceClientRef service = bgkFindKeyboard(keyboards, id);

	if (service != NULL) {
		CFTypeRef value = IOHIDServiceClientCopyProperty(service, bgkKeymapProperty);
		count = bgkReadMappings(value, src, dst, capacity);

		if (value != NULL) {
			CFRelease(value);
		}
	}

	CFRelease(keyboards);
	CFRelease(client);

	return count;
}

// bgkKeyboardKeymapSet replaces a keyboard's key mappings with the count pairs in
// src and dst. It returns 1 on success, 0 on failure, bgkUnavailable, or
// bgkNotFound.
int bgkKeyboardKeymapSet(uint64_t id, const uint64_t *src, const uint64_t *dst, int count) {
	IOHIDEventSystemClientRef client = IOHIDEventSystemClientCreateSimpleClient(kCFAllocatorDefault);
	if (client == NULL) {
		return bgkUnavailable;
	}

	CFArrayRef keyboards = bgkKeyboardServices(client);
	if (keyboards == NULL) {
		CFRelease(client);
		return bgkUnavailable;
	}

	int result = bgkNotFound;
	IOHIDServiceClientRef service = bgkFindKeyboard(keyboards, id);

	if (service != NULL) {
		CFArrayRef mappings = bgkBuildMappings(src, dst, count);
		result = IOHIDServiceClientSetProperty(service, bgkKeymapProperty, mappings) ? 1 : 0;
		CFRelease(mappings);
	}

	CFRelease(keyboards);
	CFRelease(client);

	return result;
}
