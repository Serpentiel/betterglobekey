// Package hidkeymap reads and writes the system-wide HID key remapping table,
// the UserKeyMapping property that `hidutil property --set` edits. Mappings act
// below every event tap and last until they are replaced or the system restarts.
package hidkeymap

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <stdint.h>

// Implemented in hidkeymap.m.
int bgkKeymapGet(uint64_t *src, uint64_t *dst, int capacity);
int bgkKeymapSet(const uint64_t *src, const uint64_t *dst, int count);
*/
import "C"

import "errors"

// Mapping remaps the HID usage Src (usage page << 32 | usage) to Dst.
type Mapping struct {
	Src uint64
	Dst uint64
}

// errUnavailable reports that the HID event system could not be reached.
var errUnavailable = errors.New("HID event system unavailable")

// errRejected reports that the HID event system refused the new mappings.
var errRejected = errors.New("HID event system rejected the key mappings")

// initialCapacity is the first buffer size tried when reading mappings.
const initialCapacity = 32

// Mappings returns the current system-wide key mappings.
func Mappings() ([]Mapping, error) {
	capacity := initialCapacity

	for {
		src := make([]C.uint64_t, capacity)
		dst := make([]C.uint64_t, capacity)

		count := int(C.bgkKeymapGet(&src[0], &dst[0], C.int(capacity)))
		if count < 0 {
			return nil, errUnavailable
		}

		if count > capacity {
			capacity = count

			continue
		}

		mappings := make([]Mapping, count)
		for i := range mappings {
			mappings[i] = Mapping{Src: uint64(src[i]), Dst: uint64(dst[i])}
		}

		return mappings, nil
	}
}

// SetMappings replaces the system-wide key mappings.
func SetMappings(mappings []Mapping) error {
	// Keep one element so the slices have an address even when clearing.
	src := make([]C.uint64_t, max(len(mappings), 1))
	dst := make([]C.uint64_t, max(len(mappings), 1))

	for i, mapping := range mappings {
		src[i] = C.uint64_t(mapping.Src)
		dst[i] = C.uint64_t(mapping.Dst)
	}

	if C.bgkKeymapSet(&src[0], &dst[0], C.int(len(mappings))) == 0 {
		return errRejected
	}

	return nil
}
