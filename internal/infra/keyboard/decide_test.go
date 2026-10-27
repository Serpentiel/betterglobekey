package keyboard

import (
	"testing"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

// CoreGraphics flags (kCGEventFlagMask*) and the extra kVK_* key codes the cases use.
const (
	flagShift   = 0x00020000
	flagCommand = 0x00100000
	flagFn      = 0x00800000

	keyC  = 8
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

	rightCommandDown = flags(keyRightCommand, flagCommand|deviceRightCommand)
	rightCommandUp   = flags(keyRightCommand, 0)
)

func TestDecide(t *testing.T) {
	globe := config.Trigger{Key: config.TriggerGlobe, Mode: config.TriggerModeFn}
	remapped := config.Trigger{Key: config.TriggerCapsLock, Mode: config.TriggerModeFn}
	dedicated := config.Trigger{Key: config.TriggerCapsLock, Mode: config.TriggerModeDedicated}
	rightCommand := config.Trigger{Key: "right_command", Mode: config.TriggerModePassThrough}
	rightShift := config.Trigger{Key: "right_shift", Mode: config.TriggerModePassThrough}
	f13 := config.Trigger{Key: "f13", Mode: config.TriggerModePassThrough}

	tests := []struct {
		name    string
		trigger config.Trigger
		events  []tapEvent
		want    []press // one outcome per tap completed, in order
	}{
		{"globe tap", globe, []tapEvent{globeDown, globeUp}, []press{{fired: true}}},
		{
			"globe tap with shift reverses", globe,
			[]tapEvent{flags(keyFunction, flagFn|flagShift), globeUp},
			[]press{{fired: true, reverse: true}},
		},
		{"remapped key arrives as globe", remapped, []tapEvent{globeDown, globeUp}, []press{{fired: true}}},
		{"globe does nothing in dedicated mode", dedicated, []tapEvent{globeDown, globeUp}, nil},
		{"f20 does nothing in dedicated mode", dedicated, []tapEvent{keyDown(keyF20, 0), keyUp(keyF20, 0)}, nil},
		{"globe + f1 media", globe, []tapEvent{globeDown, media, globeUp}, nil},
		{"globe + f1 key", globe, []tapEvent{globeDown, keyDown(keyF1, flagFn), keyUp(keyF1, flagFn), globeUp}, nil},
		{
			"globe ignores duplicate flags", globe,
			[]tapEvent{globeDown, globeDown, globeUp, globeUp},
			[]press{{fired: true}},
		},
		{"right command tap", rightCommand, []tapEvent{rightCommandDown, rightCommandUp}, []press{{fired: true}}},
		{
			"right command + c", rightCommand,
			[]tapEvent{rightCommandDown, keyDown(keyC, flagCommand), keyUp(keyC, flagCommand), rightCommandUp},
			nil,
		},
		{"right command + volume", rightCommand, []tapEvent{rightCommandDown, media, rightCommandUp}, nil},
		{
			"right command + fn + volume", rightCommand,
			[]tapEvent{
				rightCommandDown, flags(keyFunction, flagCommand|deviceRightCommand|flagFn), media, rightCommandUp,
			},
			nil,
		},
		{
			"right command watches only its own device flag", rightCommand,
			[]tapEvent{flags(55, flagCommand|deviceLeftCommand), flags(55, 0)},
			nil,
		},
		{"globe is not watched while passing through", rightCommand, []tapEvent{globeDown, globeUp}, nil},
		{
			"right shift trigger is not reverse by itself", rightShift,
			[]tapEvent{flags(keyRightShift, flagShift|deviceRightShift), flags(keyRightShift, 0)},
			[]press{{fired: true}},
		},
		{
			"right shift trigger with left shift reverses", rightShift,
			[]tapEvent{
				flags(keyShift, flagShift|deviceLeftShift),
				flags(keyRightShift, flagShift|deviceLeftShift|deviceRightShift),
				flags(keyRightShift, flagShift|deviceLeftShift),
			},
			[]press{{fired: true, reverse: true}},
		},
		{
			"f13 tap fires once despite autorepeat", f13,
			[]tapEvent{
				keyDown(keyF13, 0),
				{kind: eventKeyDown, keycode: keyF13, autorepeat: true},
				{kind: eventKeyDown, keycode: keyF13, autorepeat: true},
				keyUp(keyF13, 0),
			},
			[]press{{fired: true}},
		},
		{"f13 + c", f13, []tapEvent{keyDown(keyF13, 0), keyDown(keyC, 0), keyUp(keyF13, 0)}, nil},
		{"stray system event while held", globe, []tapEvent{globeDown, other, globeUp}, []press{{fired: true}}},
		{
			"modifier press while held", globe,
			[]tapEvent{globeDown, flags(keyRightShift, flagFn|flagShift), globeUp},
			[]press{{fired: true}},
		},
		{"reset mid-press", globe, []tapEvent{globeDown, reset, globeUp}, nil},
		{
			"tap after a combination", globe,
			[]tapEvent{globeDown, media, globeUp, globeDown, globeUp},
			[]press{{fired: true}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := newDecider(test.trigger, "shift")

			var got []press

			for _, event := range test.events {
				if result := d.feed(event); result.fired {
					got = append(got, result)
				}
			}

			if len(got) != len(test.want) {
				t.Fatalf("presses = %+v, want %+v", got, test.want)
			}

			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("press %d = %+v, want %+v", i, got[i], test.want[i])
				}
			}
		})
	}
}

// HID usage pages and usages the raw cases use.
const (
	pageKeyboard  = 0x07
	pageConsumer  = 0x0C
	usageCapsLock = 0x39
	usageA        = 0x04
	usageLShift   = 0xE1
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

	tests := []struct {
		name   string
		values []rawValue
		want   []press
	}{
		{"tap", []rawValue{capsLock(true, true, 0), capsLock(false, true, 0)}, []press{{fired: true}}},
		{
			"tap with shift reverses",
			[]rawValue{capsLock(true, true, flagShift), capsLock(false, true, flagShift)},
			[]press{{fired: true, reverse: true}},
		},
		{"key while held", []rawValue{capsLock(true, true, 0), keyA, capsLock(false, true, 0)}, nil},
		{"media key while held", []rawValue{capsLock(true, true, 0), volumeUp, capsLock(false, true, 0)}, nil},
		{
			"modifier and rollover while held",
			[]rawValue{capsLock(true, true, 0), shiftDown, rollover, capsLock(false, true, 0)},
			[]press{{fired: true}},
		},
		{"same key on another keyboard", []rawValue{capsLock(true, false, 0), capsLock(false, false, 0)}, nil},
		{
			"same key on another keyboard while held",
			[]rawValue{capsLock(true, true, 0), capsLock(true, false, 0), capsLock(false, true, 0)},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := newRawDecider("shift")

			var got []press

			for _, value := range test.values {
				if result := d.feedRaw(value); result.fired {
					got = append(got, result)
				}
			}

			if len(got) != len(test.want) {
				t.Fatalf("presses = %+v, want %+v", got, test.want)
			}

			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("press %d = %+v, want %+v", i, got[i], test.want[i])
				}
			}
		})
	}
}

func TestTriggerForFallsBackToGlobe(t *testing.T) {
	for _, trigger := range []config.Trigger{
		{Key: config.TriggerGlobe, Mode: config.TriggerModePassThrough},
		{Key: config.TriggerGlobe, Mode: config.TriggerModeDedicated},
		{Key: "right_command", Mode: config.TriggerModeFn},
		{Key: "unknown", Mode: config.TriggerModePassThrough},
	} {
		if got := triggerFor(trigger); got != globeTrigger {
			t.Errorf("triggerFor(%+v) is not the Globe trigger", trigger)
		}
	}
}

func TestPassThroughTriggersCoverKeys(t *testing.T) {
	for _, key := range config.TriggerKeys {
		_, ok := passThroughTriggers[key]
		want := key != config.TriggerGlobe && key != config.TriggerCapsLock

		if ok != want {
			t.Errorf("passThroughTriggers[%q] present = %v, want %v", key, ok, want)
		}
	}
}
