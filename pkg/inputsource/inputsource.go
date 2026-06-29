// Package inputsource provides access to macOS input sources.
// nolint:nlreturn // nlreturn misreports against the cgo string-bridge calls (CString/GoString/free).
package inputsource

/*
#cgo LDFLAGS: -framework Carbon -framework CoreFoundation

#include <stdlib.h>

// Implemented in inputsource.m.
char *bgkAllInputSources(void);
char *bgkCurrentInputSource(void);
char *bgkInputSourceName(const char *sourceID);
void bgkSelectInputSource(const char *sourceID);
*/
import "C"

import (
	"slices"
	"strings"
	"unsafe"
)

// All returns the IDs of all available input sources, sorted.
func All() []string {
	raw := C.bgkAllInputSources()
	if raw == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(raw))

	joined := C.GoString(raw)
	if joined == "" {
		return nil
	}

	sources := strings.Split(joined, "\n")
	slices.Sort(sources)

	return sources
}

// Current returns the current input source's ID.
func Current() string {
	raw := C.bgkCurrentInputSource()
	if raw == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(raw))

	return C.GoString(raw)
}

// Name returns the localized display name for the input source with the given
// ID, or an empty string if it is not available.
func Name(id string) string {
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))

	raw := C.bgkInputSourceName(cID)
	if raw == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(raw))

	return C.GoString(raw)
}

// Select selects the input source with the specified ID.
func Select(id string) {
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))

	C.bgkSelectInputSource(cID)
}
