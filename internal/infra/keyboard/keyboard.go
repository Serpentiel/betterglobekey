// Package keyboard listens for trigger key taps using a CoreGraphics event tap,
// giving the application full control over detection without third-party
// dependencies. The default trigger, the Globe key, surfaces as a
// flags-changed event with key code kVK_Function (63); a release (the fn flag
// clearing) is treated as a press, and a held Shift marks it as a reverse press.
// A key remapped to fn arrives the same way. Dedicated and passed-through
// trigger keys are read per keyboard from raw HID input instead, which also
// records keys for the companion.
package keyboard

/*
#cgo LDFLAGS: -framework AppKit -framework CoreGraphics -framework CoreFoundation -framework IOKit
#include <CoreGraphics/CoreGraphics.h>

#include "decide.h"

// Implemented in keyboard.m.
void bgkSetReverseMask(CGEventFlags mask);
void bgkSetTrigger(bgkTrigger trigger);
void bgkSetRawKeys(const bgkRawKey *keys, int count);
void bgkSetRecording(int on);
int bgkStart(void);
void bgkRun(void);
void bgkStop(void);
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

// onPress is the active press handler. Only one Listener runs at a time.
var onPress func(reverse bool)

// modifierMasks maps configuration modifier names to CoreGraphics event flags.
var modifierMasks = map[string]C.CGEventFlags{
	"shift":   C.kCGEventFlagMaskShift,
	"option":  C.kCGEventFlagMaskAlternate,
	"control": C.kCGEventFlagMaskControl,
	"command": C.kCGEventFlagMaskCommand,
}

// SetReverseModifier sets which modifier key marks a press as reverse. Unknown
// names fall back to Shift. It is safe to call while listening; the event tap
// reads the value live.
func SetReverseModifier(modifier string) {
	mask, ok := modifierMasks[modifier]
	if !ok {
		mask = C.kCGEventFlagMaskShift
	}

	C.bgkSetReverseMask(mask)
}

// Device-dependent modifier flags (NX_DEVICE*KEYMASK), which tell the left and
// right keys of a modifier family apart.
const (
	deviceLeftControl  = 0x00000001
	deviceLeftShift    = 0x00000002
	deviceRightShift   = 0x00000004
	deviceLeftCommand  = 0x00000008
	deviceRightCommand = 0x00000010
	deviceLeftOption   = 0x00000020
	deviceRightOption  = 0x00000040
	deviceRightControl = 0x00002000
)

// Virtual key codes (kVK_* in Carbon's Events.h).
const (
	keyFunction     = 63
	keyCommand      = 55
	keyRightCommand = 54
	keyOption       = 58
	keyRightOption  = 61
	keyControl      = 59
	keyRightControl = 62
	keyShift        = 56
	keyRightShift   = 60
)

// globeTrigger is the Globe key, also used for a key remapped to fn.
var globeTrigger = C.bgkTrigger{kind: C.bgkKindFlag, keycode: keyFunction, pressFlag: C.kCGEventFlagMaskSecondaryFn}

// modifierTrigger describes a modifier key: its key code, its own device flag,
// its family flag, and the device flag of the family's other key.
func modifierTrigger(keycode int64, device, family, otherSide C.uint64_t) C.bgkTrigger {
	return C.bgkTrigger{
		kind:          C.bgkKindFlag,
		keycode:       C.int64_t(keycode),
		pressFlag:     device,
		familyFlag:    family,
		otherSideFlag: otherSide,
	}
}

// modifierTriggers describe the modifier trigger keys, whose family flags keep
// a raw press of the key from reversing itself.
var modifierTriggers = map[string]C.bgkTrigger{
	"right_command": modifierTrigger(keyRightCommand, deviceRightCommand, C.kCGEventFlagMaskCommand, deviceLeftCommand),
	"left_command":  modifierTrigger(keyCommand, deviceLeftCommand, C.kCGEventFlagMaskCommand, deviceRightCommand),
	"right_option":  modifierTrigger(keyRightOption, deviceRightOption, C.kCGEventFlagMaskAlternate, deviceLeftOption),
	"left_option":   modifierTrigger(keyOption, deviceLeftOption, C.kCGEventFlagMaskAlternate, deviceRightOption),
	"right_control": modifierTrigger(keyRightControl, deviceRightControl, C.kCGEventFlagMaskControl, deviceLeftControl),
	"left_control":  modifierTrigger(keyControl, deviceLeftControl, C.kCGEventFlagMaskControl, deviceRightControl),
	"right_shift":   modifierTrigger(keyRightShift, deviceRightShift, C.kCGEventFlagMaskShift, deviceLeftShift),
	"left_shift":    modifierTrigger(keyShift, deviceLeftShift, C.kCGEventFlagMaskShift, deviceRightShift),
}

// noTrigger matches no event, for when the Globe key does not switch.
var noTrigger = C.bgkTrigger{kind: C.bgkKindKey, keycode: C.bgkRawTriggerCode}

// tapTrigger returns the descriptor the event tap watches: the Globe key (a
// key remapped to fn arrives as it), or nothing.
func tapTrigger(globe bool) C.bgkTrigger {
	if globe {
		return globeTrigger
	}

	return noTrigger
}

// SetGlobe sets whether the Globe key switches, dropping any press in progress.
// It is safe to call while listening.
func SetGlobe(globe bool) {
	C.bgkSetTrigger(tapTrigger(globe))
}

// RawKey is a trigger key read from one keyboard's raw HID input.
type RawKey struct {
	// Vendor, Product, and Location identify the keyboard.
	Vendor, Product, Location uint32
	// Usage is the key's usage on the HID Keyboard/Keypad page.
	Usage uint32
	// Key is the trigger key's name (see config.TriggerKeys).
	Key string
}

// rawKey converts key, with the family flags of a modifier.
func rawKey(key RawKey) C.bgkRawKey {
	raw := C.bgkRawKey{
		vendor:   C.uint32_t(key.Vendor),
		product:  C.uint32_t(key.Product),
		location: C.uint32_t(key.Location),
		usage:    C.uint32_t(key.Usage),
	}

	if modifier, ok := modifierTriggers[key.Key]; ok {
		raw.familyFlag = modifier.familyFlag
		raw.otherSideFlag = modifier.otherSideFlag
	}

	return raw
}

// SetRawKeys sets the trigger keys read from each keyboard's raw HID input,
// which needs Input Monitoring. Raw input is read only while there are keys or
// a recording; each call retries opening it, so a grant given later takes
// effect. It is safe to call while listening.
func SetRawKeys(keys []RawKey) {
	raw := make([]C.bgkRawKey, max(len(keys), 1))
	for i, key := range keys {
		raw[i] = rawKey(key)
	}

	C.bgkSetRawKeys(&raw[0], C.int(len(keys)))
}

// RawInputState is what happened to raw input (see OnRawInput).
type RawInputState int

// Raw input states.
const (
	RawInputClosed RawInputState = -1
	RawInputFailed RawInputState = 0
	RawInputOpened RawInputState = 1
)

// rawInputHandler receives raw input state changes.
var rawInputHandler atomic.Pointer[func(state RawInputState, status int32)]

// OnRawInput registers handler to receive raw input being opened, failing to
// open (with the IOReturn status), or closed. It is called on the event tap's
// thread, after every open attempt.
func OnRawInput(handler func(state RawInputState, status int32)) {
	rawInputHandler.Store(&handler)
}

// RecordedKey is a key pressed on a keyboard while recording.
type RecordedKey struct {
	// Vendor, Product, and Location identify the keyboard.
	Vendor, Product, Location uint32
	// Usage is the key's HID usage, its usage page shifted left 32 bits.
	Usage uint64
}

// ErrRecordingReplaced reports that a newer recording took over.
var ErrRecordingReplaced = errors.New("recording replaced by a newer one")

// RawInputError reports that raw input could not be opened.
type RawInputError struct {
	// Status is the IOReturn opening raw input failed with.
	Status int32
}

func (e *RawInputError) Error() string {
	return fmt.Sprintf("reading raw keyboard input failed (%#x)", uint32(e.Status)) //nolint:gosec // an IOReturn
}

// recording is a Record call waiting for a key.
type recording struct {
	accept func(RecordedKey) bool
	keys   chan RecordedKey
	errs   chan error
}

// deliver sends result to results, unless it already holds one.
func deliver[T any](results chan T, result T) {
	select {
	case results <- result:
	default:
	}
}

var (
	recordMu sync.Mutex
	// active is the recording in progress, if any.
	active *recording
)

// Record returns the next key pressed that accept takes, read from raw HID
// input, which needs Input Monitoring. Raw input reports a key before any
// remapping, so a remapped or silenced key records as itself. Only one
// recording runs at a time; a newer one ends this one with ErrRecordingReplaced.
// It fails with a *RawInputError when raw input cannot be opened, and with
// ctx's error when ctx is done first.
func Record(ctx context.Context, accept func(RecordedKey) bool) (RecordedKey, error) {
	r := &recording{accept: accept, keys: make(chan RecordedKey, 1), errs: make(chan error, 1)}

	recordMu.Lock()
	if active != nil {
		deliver(active.errs, ErrRecordingReplaced)
	}

	active = r
	recordMu.Unlock()

	C.bgkSetRecording(1)
	defer stopRecording(r)

	select {
	case key := <-r.keys:
		return key, nil
	case err := <-r.errs:
		return RecordedKey{}, err
	case <-ctx.Done():
		return RecordedKey{}, ctx.Err()
	}
}

// stopRecording ends r, unless a newer recording took over.
func stopRecording(r *recording) {
	recordMu.Lock()
	defer recordMu.Unlock()

	if active == r {
		active = nil

		C.bgkSetRecording(0)
	}
}

// recorded hands a key pressed while recording to the recording in progress.
func recorded(key RecordedKey) {
	recordMu.Lock()
	defer recordMu.Unlock()

	if active != nil && active.accept(key) {
		deliver(active.keys, key)
	}
}

// rawInputFailed ends the recording in progress with err.
func rawInputFailed(err error) {
	recordMu.Lock()
	defer recordMu.Unlock()

	if active != nil {
		deliver(active.errs, err)
	}
}

// Listener delivers trigger key releases via a CoreGraphics event tap.
type Listener struct{}

// New returns a Listener.
func New() *Listener {
	return &Listener{}
}

// Listen installs the event tap and runs its run loop, invoking handler for each
// trigger key release (with reverse set when Shift is held). It blocks until Stop
// is called. If the tap cannot be created (e.g. Accessibility is not granted), it
// returns immediately.
func (*Listener) Listen(handler func(reverse bool)) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	onPress = handler

	if C.bgkStart() == 0 {
		return
	}

	C.bgkRun()
}

// Stop stops the event tap's run loop, unblocking Listen.
func (*Listener) Stop() {
	C.bgkStop()
}
