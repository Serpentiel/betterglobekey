// Package inputsource provides access to macOS input sources.
package inputsource

/*
#cgo LDFLAGS: -framework Carbon -framework CoreFoundation

#include <Carbon/Carbon.h>
#include <CoreFoundation/CoreFoundation.h>
*/
import "C"
import "unsafe"

// cfStringToString converts a CFString to a Go string.
func cfStringToString(cfString C.CFStringRef) string {
	length := C.CFStringGetLength(cfString) // nolint:nlreturn

	maxSize := C.CFStringGetMaximumSizeForEncoding(length, C.kCFStringEncodingUTF8)

	buffer := make([]byte, maxSize)

	bufferPtr := (*C.char)(unsafe.Pointer(&buffer[0]))

	C.CFStringGetCString(cfString, bufferPtr, maxSize, C.kCFStringEncodingUTF8) // nolint:nlreturn

	return C.GoString(bufferPtr)
}
