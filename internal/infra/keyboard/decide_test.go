package keyboard

import "testing"

// CoreGraphics flags (kCGEventFlagMask*) and the extra kVK_* key codes the cases use.
const (
	flagShift = 0x00020000
	flagFn    = 0x00800000

	keyF1 = 122
)

func flags(keycode int64, value uint64) tapEvent {
	return tapEvent{kind: eventFlags, keycode: keycode, flags: value}
}

func keyDown(keycode int64, value uint64) tapEvent {
	return tapEvent{kind: eventKeyDown, keycode: keycode, flags: value}
}

func keyUp(keycode int64, value uint64) tapEvent {
	return tapEvent{kind: eventKeyUp, keycode: keycode, flags: value}
}

var (
	media = tapEvent{kind: eventMedia, keycode: -1}
	other = tapEvent{kind: eventOther, keycode: -1}
	reset = tapEvent{kind: eventReset}

	globeDown = flags(keyFunction, flagFn)
	globeUp   = flags(keyFunction, 0)
)

func TestDecide(t *testing.T) {
	tests := []struct {
		name   string
		globe  bool
		events []tapEvent
		want   []press // one outcome per tap completed, in order
	}{
		{"globe tap", true, []tapEvent{globeDown, globeUp}, []press{{fired: true}}},
		{
			"globe tap with shift reverses", true,
			[]tapEvent{flags(keyFunction, flagFn|flagShift), globeUp},
			[]press{{fired: true, reverse: true}},
		},
		{"globe not watched", false, []tapEvent{globeDown, globeUp}, nil},
		{"globe + f1 media", true, []tapEvent{globeDown, media, globeUp}, nil},
		{"globe + f1 key", true, []tapEvent{globeDown, keyDown(keyF1, flagFn), keyUp(keyF1, flagFn), globeUp}, nil},
		{
			"globe ignores duplicate flags", true,
			[]tapEvent{globeDown, globeDown, globeUp, globeUp},
			[]press{{fired: true}},
		},
		{"stray system event while held", true, []tapEvent{globeDown, other, globeUp}, []press{{fired: true}}},
		{
			"modifier press while held", true,
			[]tapEvent{globeDown, flags(keyRightShift, flagFn|flagShift), globeUp},
			[]press{{fired: true}},
		},
		{"reset mid-press", true, []tapEvent{globeDown, reset, globeUp}, nil},
		{
			"tap after a combination", true,
			[]tapEvent{globeDown, media, globeUp, globeDown, globeUp},
			[]press{{fired: true}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := newDecider(test.globe, "shift")

			var got []press

			for _, event := range test.events {
				if result := d.feed(event); result.fired {
					got = append(got, result)
				}
			}

			assertPresses(t, got, test.want)
		})
	}
}

func assertPresses(t *testing.T, got, want []press) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("presses = %+v, want %+v", got, want)
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("press %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// HID usage pages and usages the raw cases use.
const (
	pageKeyboard  = 0x07
	pageConsumer  = 0x0C
	usageCapsLock = 0x39
	usageA        = 0x04
	usageLShift   = 0xE1
	usageRShift   = 0xE5
	usageVolumeUp = 0xE9
)

func capsLock(down, own bool, flags uint64) rawValue {
	value := int64(0)
	if down {
		value = 1
	}

	return rawValue{page: pageKeyboard, usage: usageCapsLock, value: value, own: own, flags: flags}
}

func TestDecideRaw(t *testing.T) {
	keyA := rawValue{page: pageKeyboard, usage: usageA, value: 1}
	shiftDown := rawValue{page: pageKeyboard, usage: usageLShift, value: 1, flags: flagShift}
	volumeUp := rawValue{page: pageConsumer, usage: usageVolumeUp, value: 1}
	rollover := rawValue{page: pageKeyboard, usage: 0xFFFFFFFF, value: 1}

	rightShift := func(value int64, flags uint64) rawValue {
		return rawValue{page: pageKeyboard, usage: usageRShift, value: value, own: true, flags: flags}
	}
	rightShiftDown := rightShift(1, flagShift|deviceRightShift)
	rightShiftUp := rightShift(0, 0)
	leftShiftHeld := uint64(flagShift | deviceLeftShift | deviceRightShift)

	tests := []struct {
		name   string
		key    string
		values []rawValue
		want   []press
	}{
		{"tap", "caps_lock", []rawValue{capsLock(true, true, 0), capsLock(false, true, 0)}, []press{{fired: true}}},
		{
			"tap with shift reverses", "caps_lock",
			[]rawValue{capsLock(true, true, flagShift), capsLock(false, true, flagShift)},
			[]press{{fired: true, reverse: true}},
		},
		{"key while held", "caps_lock", []rawValue{capsLock(true, true, 0), keyA, capsLock(false, true, 0)}, nil},
		{
			"media key while held", "caps_lock",
			[]rawValue{capsLock(true, true, 0), volumeUp, capsLock(false, true, 0)}, nil,
		},
		{
			"modifier and rollover while held", "caps_lock",
			[]rawValue{capsLock(true, true, 0), shiftDown, rollover, capsLock(false, true, 0)},
			[]press{{fired: true}},
		},
		{
			"same key on another keyboard", "caps_lock",
			[]rawValue{capsLock(true, false, 0), capsLock(false, false, 0)}, nil,
		},
		{
			"same key on another keyboard while held", "caps_lock",
			[]rawValue{capsLock(true, true, 0), capsLock(true, false, 0), capsLock(false, true, 0)},
			nil,
		},
		{
			"right shift is not reverse by itself", "right_shift",
			[]rawValue{rightShiftDown, rightShiftUp}, []press{{fired: true}},
		},
		{
			"right shift with left shift held reverses", "right_shift",
			[]rawValue{rightShift(1, leftShiftHeld), rightShiftUp},
			[]press{{fired: true, reverse: true}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := newRawDecider(test.key, "shift")

			var got []press

			for _, value := range test.values {
				if result := d.feedRaw(value); result.fired {
					got = append(got, result)
				}
			}

			assertPresses(t, got, test.want)
		})
	}
}

func TestRawKeyFamilies(t *testing.T) {
	if raw := rawKey(RawKey{Key: "right_shift"}); raw.familyFlag != flagShift || raw.otherSideFlag != deviceLeftShift {
		t.Errorf("right shift = %+v, want the shift family", raw)
	}

	if raw := rawKey(RawKey{Key: "f13"}); raw.familyFlag != 0 || raw.otherSideFlag != 0 {
		t.Errorf("f13 = %+v, want no family", raw)
	}
}
