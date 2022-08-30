// Package inputsource provides access to macOS input sources.
// nolint:nlreturn // nlreturn does not seems to work properly with this package.
package inputsource

/*
#cgo LDFLAGS: -framework Carbon -framework CoreFoundation

#include <Carbon/Carbon.h>
#include <CoreFoundation/CoreFoundation.h>
*/
import "C"
import (
	"unsafe"

	"golang.org/x/exp/slices"
)

// All returns a slice of all available input sources' IDs.
func All() []string {
	filter := C.CFDictionaryCreateMutable(C.CFAllocatorRef(0), 0, nil, nil)

	// nolint:govet
	C.CFDictionarySetValue(
		filter, unsafe.Pointer(C.kTISPropertyInputSourceIsEnabled), unsafe.Pointer(C.kCFBooleanTrue),
	)

	// nolint:govet
	C.CFDictionarySetValue(
		filter, unsafe.Pointer(C.kTISPropertyInputSourceIsSelectCapable), unsafe.Pointer(C.kCFBooleanTrue),
	)

	defer C.CFRelease(C.CFTypeRef(filter))

	sourceList := C.TISCreateInputSourceList(C.CFDictionaryRef(filter), 0)

	defer C.CFRelease(C.CFTypeRef(sourceList))

	var result []string

	for i := C.CFIndex(0); i < C.CFArrayGetCount(sourceList); i++ {
		inputSource := C.TISInputSourceRef(C.CFArrayGetValueAtIndex(sourceList, i))

		t := C.CFStringRef(C.TISGetInputSourceProperty(inputSource, C.kTISPropertyInputSourceType))
		if C.CFStringCompare(t, C.kTISTypeKeyboardLayout, 0) != C.kCFCompareEqualTo &&
			C.CFStringCompare(t, C.kTISTypeKeyboardInputMode, 0) != C.kCFCompareEqualTo {
			continue
		}

		id := C.CFStringRef(C.TISGetInputSourceProperty(inputSource, C.kTISPropertyInputSourceID))

		result = append(result, cfStringToString(id))
	}

	slices.Sort(result)

	return result
}

// Current returns current input source's ID.
func Current() string {
	inputSource := C.TISCopyCurrentKeyboardInputSource()

	defer C.CFRelease(C.CFTypeRef(inputSource))

	return cfStringToString(C.CFStringRef(C.TISGetInputSourceProperty(inputSource, C.kTISPropertyInputSourceID)))
}

// Select selects input source with specified ID.
func Select(id string) {
	cID := C.CString(id)

	defer C.free(unsafe.Pointer(cID))

	cfID := C.CFStringCreateWithBytes(0, (*C.uchar)(unsafe.Pointer(cID)), C.long(len(id)), C.kCFStringEncodingUTF8, 0)

	defer C.CFRelease(C.CFTypeRef(cfID))

	filter := C.CFDictionaryCreateMutable(C.CFAllocatorRef(0), 0, nil, nil)

	// nolint:govet
	C.CFDictionarySetValue(filter, unsafe.Pointer(C.kTISPropertyInputSourceID), unsafe.Pointer(cfID))

	defer C.CFRelease(C.CFTypeRef(filter))

	sourceList := C.TISCreateInputSourceList(C.CFDictionaryRef(filter), 0)

	defer C.CFRelease(C.CFTypeRef(sourceList))

	inputSource := C.TISInputSourceRef(C.CFArrayGetValueAtIndex(sourceList, 0))

	C.TISSelectInputSource(inputSource)
}
