// Package inputsource provides access to macOS input sources.
// nolint:nlreturn // nlreturn does not seem to work properly with this cgo package.
package inputsource

/*
#cgo LDFLAGS: -framework Carbon -framework CoreFoundation

#include <Carbon/Carbon.h>
#include <CoreFoundation/CoreFoundation.h>
*/
import "C"
import (
	"slices"
	"unsafe"
)

// All returns a slice of all available input sources' IDs.
func All() []string {
	filter := newFilter()

	// nolint:govet
	C.CFDictionarySetValue(
		filter, unsafe.Pointer(C.kTISPropertyInputSourceIsEnabled), unsafe.Pointer(C.kCFBooleanTrue),
	)

	// nolint:govet
	C.CFDictionarySetValue(
		filter, unsafe.Pointer(C.kTISPropertyInputSourceIsSelectCapable), unsafe.Pointer(C.kCFBooleanTrue),
	)

	var result []string

	withInputSources(filter, func(list C.CFArrayRef) {
		for i := C.CFIndex(0); i < C.CFArrayGetCount(list); i++ {
			inputSource := C.TISInputSourceRef(C.CFArrayGetValueAtIndex(list, i))

			t := C.CFStringRef(C.TISGetInputSourceProperty(inputSource, C.kTISPropertyInputSourceType))
			if C.CFStringCompare(t, C.kTISTypeKeyboardLayout, 0) != C.kCFCompareEqualTo &&
				C.CFStringCompare(t, C.kTISTypeKeyboardInputMode, 0) != C.kCFCompareEqualTo {
				continue
			}

			id := C.CFStringRef(C.TISGetInputSourceProperty(inputSource, C.kTISPropertyInputSourceID))

			result = append(result, cfStringToString(id))
		}
	})

	slices.Sort(result)

	return result
}

// Current returns current input source's ID.
func Current() string {
	inputSource := C.TISCopyCurrentKeyboardInputSource()

	defer C.CFRelease(C.CFTypeRef(inputSource))

	return cfStringToString(C.CFStringRef(C.TISGetInputSourceProperty(inputSource, C.kTISPropertyInputSourceID)))
}

// Name returns the localized display name for the input source with the given
// ID, or an empty string if it is not available.
func Name(id string) string {
	cID := C.CString(id)

	defer C.free(unsafe.Pointer(cID))

	cfID := C.CFStringCreateWithBytes(0, (*C.uchar)(unsafe.Pointer(cID)), C.long(len(id)), C.kCFStringEncodingUTF8, 0)

	defer C.CFRelease(C.CFTypeRef(cfID))

	filter := newFilter()

	// nolint:govet
	C.CFDictionarySetValue(filter, unsafe.Pointer(C.kTISPropertyInputSourceID), unsafe.Pointer(cfID))

	var name string

	withInputSources(filter, func(list C.CFArrayRef) {
		if C.CFArrayGetCount(list) == 0 {
			return
		}

		inputSource := C.TISInputSourceRef(C.CFArrayGetValueAtIndex(list, 0))

		localized := C.CFStringRef(C.TISGetInputSourceProperty(inputSource, C.kTISPropertyLocalizedName))
		if localized != 0 {
			name = cfStringToString(localized)
		}
	})

	return name
}

// Select selects input source with specified ID.
func Select(id string) {
	cID := C.CString(id)

	defer C.free(unsafe.Pointer(cID))

	cfID := C.CFStringCreateWithBytes(0, (*C.uchar)(unsafe.Pointer(cID)), C.long(len(id)), C.kCFStringEncodingUTF8, 0)

	defer C.CFRelease(C.CFTypeRef(cfID))

	filter := newFilter()

	// nolint:govet
	C.CFDictionarySetValue(filter, unsafe.Pointer(C.kTISPropertyInputSourceID), unsafe.Pointer(cfID))

	withInputSources(filter, func(list C.CFArrayRef) {
		if C.CFArrayGetCount(list) == 0 {
			return
		}

		C.TISSelectInputSource(C.TISInputSourceRef(C.CFArrayGetValueAtIndex(list, 0)))
	})
}

// newFilter creates an empty mutable CFDictionary used to filter input sources.
func newFilter() C.CFMutableDictionaryRef {
	return C.CFDictionaryCreateMutable(C.CFAllocatorRef(0), 0, nil, nil)
}

// withInputSources invokes fn with the input sources matching filter, releasing
// both the filter and the resulting list afterwards.
func withInputSources(filter C.CFMutableDictionaryRef, fn func(C.CFArrayRef)) {
	defer C.CFRelease(C.CFTypeRef(filter))

	list := C.TISCreateInputSourceList(C.CFDictionaryRef(filter), 0)

	defer C.CFRelease(C.CFTypeRef(list))

	fn(list)
}

// cfStringToString converts a CFString to a Go string.
func cfStringToString(cfString C.CFStringRef) string {
	length := C.CFStringGetLength(cfString)

	maxSize := C.CFStringGetMaximumSizeForEncoding(length, C.kCFStringEncodingUTF8)

	buffer := make([]byte, maxSize)

	bufferPtr := (*C.char)(unsafe.Pointer(&buffer[0]))

	C.CFStringGetCString(cfString, bufferPtr, maxSize, C.kCFStringEncodingUTF8)

	return C.GoString(bufferPtr)
}
