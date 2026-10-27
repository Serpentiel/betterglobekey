// Package hidkeymap reads and writes HID key remapping tables: the system-wide
// UserKeyMapping property that `hidutil property --set` edits, and each
// keyboard's own, which `hidutil property --matching` edits. Mappings act below
// every event tap; a keyboard's own last until it disconnects, the system-wide
// one until the system restarts.
package hidkeymap

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include "hidkeymap.h"
*/
import "C"

import (
	"errors"
	"fmt"
)

// Mapping remaps the HID usage Src (usage page << 32 | usage) to Dst.
type Mapping struct {
	Src uint64
	Dst uint64
}

// Keyboard is a connected keyboard service.
type Keyboard struct {
	// ID is the service's registry entry ID, unique while it is connected.
	ID       uint64
	Vendor   uint32
	Product  uint32
	Location uint32
	// Name is the product name, or empty.
	Name string
	// Transport is the connection, such as USB or Bluetooth, or empty.
	Transport string
	BuiltIn   bool
}

// errUnavailable reports that the HID event system could not be reached.
var errUnavailable = errors.New("HID event system unavailable")

// errRejected reports that the HID event system refused the new mappings.
var errRejected = errors.New("HID event system rejected the key mappings")

// ErrNotFound reports that a keyboard is no longer connected.
var ErrNotFound = errors.New("keyboard not connected")

// initialCapacity is the first buffer size tried when reading a list.
const initialCapacity = 32

// readMappings reads mappings with get, growing its buffers until they fit.
func readMappings(get func(src, dst *C.uint64_t, capacity C.int) C.int) ([]Mapping, error) {
	capacity := initialCapacity

	for {
		src := make([]C.uint64_t, capacity)
		dst := make([]C.uint64_t, capacity)

		count := int(get(&src[0], &dst[0], C.int(capacity)))

		switch {
		case count == C.bgkNotFound:
			return nil, ErrNotFound
		case count < 0:
			return nil, errUnavailable
		case count > capacity:
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

// writeMappings converts mappings and passes them to set.
func writeMappings(mappings []Mapping, set func(src, dst *C.uint64_t, count C.int) C.int) error {
	// Keep one element so the slices have an address even when clearing.
	src := make([]C.uint64_t, max(len(mappings), 1))
	dst := make([]C.uint64_t, max(len(mappings), 1))

	for i, mapping := range mappings {
		src[i] = C.uint64_t(mapping.Src)
		dst[i] = C.uint64_t(mapping.Dst)
	}

	switch set(&src[0], &dst[0], C.int(len(mappings))) {
	case 1:
		return nil
	case C.bgkNotFound:
		return ErrNotFound
	case C.bgkUnavailable:
		return errUnavailable
	default:
		return errRejected
	}
}

// Mappings returns the current system-wide key mappings.
func Mappings() ([]Mapping, error) {
	return readMappings(func(src, dst *C.uint64_t, capacity C.int) C.int {
		return C.bgkKeymapGet(src, dst, capacity)
	})
}

// SetMappings replaces the system-wide key mappings.
func SetMappings(mappings []Mapping) error {
	return writeMappings(mappings, func(src, dst *C.uint64_t, count C.int) C.int {
		return C.bgkKeymapSet(src, dst, count)
	})
}

// Keyboards returns the connected keyboard services.
func Keyboards() ([]Keyboard, error) {
	capacity := initialCapacity

	for {
		raw := make([]C.bgkKeyboard, capacity)

		count := int(C.bgkKeyboardsList(&raw[0], C.int(capacity)))
		if count < 0 {
			return nil, errUnavailable
		}

		if count > capacity {
			capacity = count

			continue
		}

		keyboards := make([]Keyboard, count)
		for i := range keyboards {
			keyboards[i] = Keyboard{
				ID:        uint64(raw[i].id),
				Vendor:    uint32(raw[i].vendor),
				Product:   uint32(raw[i].product),
				Location:  uint32(raw[i].location),
				Name:      C.GoString(&raw[i].name[0]),
				Transport: C.GoString(&raw[i].transport[0]),
				BuiltIn:   raw[i].builtIn != 0,
			}
		}

		return keyboards, nil
	}
}

// KeyboardMappings returns the key mappings of the keyboard with registry ID id.
func KeyboardMappings(id uint64) ([]Mapping, error) {
	mappings, err := readMappings(func(src, dst *C.uint64_t, capacity C.int) C.int {
		return C.bgkKeyboardKeymapGet(C.uint64_t(id), src, dst, capacity)
	})
	if err != nil {
		return nil, fmt.Errorf("keyboard %#x: %w", id, err)
	}

	return mappings, nil
}

// SetKeyboardMappings replaces the key mappings of the keyboard with registry ID id.
func SetKeyboardMappings(id uint64, mappings []Mapping) error {
	err := writeMappings(mappings, func(src, dst *C.uint64_t, count C.int) C.int {
		return C.bgkKeyboardKeymapSet(C.uint64_t(id), src, dst, count)
	})
	if err != nil {
		return fmt.Errorf("keyboard %#x: %w", id, err)
	}

	return nil
}
