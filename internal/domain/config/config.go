// Package config defines the application's configuration model and its validation rules.
//
// The types here are pure: they describe a fully-resolved configuration and know
// nothing about how it is stored, parsed, or migrated. Loading and migrating
// on-disk configuration is the responsibility of the infrastructure layer.
package config

import (
	"fmt"
	"regexp"
	"slices"
	"time"
)

// LogLevels are the accepted logger levels, from most to least verbose.
var LogLevels = []string{"debug", "info", "warn", "error"}

// ReverseModifiers are the accepted modifier keys for a reverse press.
var ReverseModifiers = []string{"shift", "option", "control", "command"}

// TriggerGlobe is the default trigger key: the Globe key itself.
const TriggerGlobe = "globe"

// TriggerCapsLock is the Caps Lock key, which can only be remapped, not passed through.
const TriggerCapsLock = "caps_lock"

// TriggerKeys are the accepted trigger keys.
var TriggerKeys = []string{
	TriggerGlobe,
	TriggerCapsLock,
	"right_command", "right_option", "right_control", "right_shift",
	"left_command", "left_option", "left_control", "left_shift",
	"f13", "f14", "f15", "f16", "f17", "f18", "f19", "f20",
	"menu", "insert", "print_screen", "scroll_lock", "pause",
}

// Trigger modes: how a trigger key other than the Globe key is detected.
const (
	// TriggerModeFn remaps the key to fn, the Globe key's usage, so it switches
	// input sources and works in fn shortcuts (fn+arrows, fn+F-keys, ...).
	TriggerModeFn = "fn"
	// TriggerModeDedicated silences the key, remapping it to one macOS has no key
	// code for, and reads it from the keyboard's raw input instead, which needs
	// Input Monitoring.
	TriggerModeDedicated = "dedicated"
	// TriggerModePassThrough keeps the key's normal function and reads it from
	// the keyboard's raw input, switching only when it is tapped on its own.
	TriggerModePassThrough = "pass_through"
)

// TriggerModes are the accepted trigger modes.
var TriggerModes = []string{TriggerModeFn, TriggerModeDedicated, TriggerModePassThrough}

// keyboardIDPattern matches a keyboard ID: vendor and product IDs as four
// lower-case hex digits each.
var keyboardIDPattern = regexp.MustCompile(`^[0-9a-f]{4}:[0-9a-f]{4}$`)

// KeyboardID formats a keyboard's vendor and product IDs as a keyboard ID.
func KeyboardID(vendor, product uint32) string {
	return fmt.Sprintf("%04x:%04x", vendor, product)
}

// Config is the fully-resolved application configuration.
type Config struct {
	// Logger holds the logging configuration.
	Logger Logger
	// DoublePress configures the double-press (collection switching) behavior.
	DoublePress DoublePress
	// Reverse configures the modifier that inverts a press.
	Reverse Reverse
	// HUD configures the on-screen overlay shown when the input source changes.
	HUD HUD
	// Trigger selects the key that switches input sources on each keyboard.
	Trigger Trigger
	// Collections are the ordered input source collections the Globe key cycles through.
	Collections []Collection
}

// Logger holds the logging configuration.
type Logger struct {
	// Path is the path to the log file.
	Path string
	// Level is the minimum level to log (see LogLevels).
	Level string
	// RetentionDays is the number of days to retain rotated log files.
	RetentionDays int
	// RetentionFiles is the number of rotated log files to retain.
	RetentionFiles int
}

// DoublePress configures the double-press gesture.
type DoublePress struct {
	// Enabled turns collection switching (double press) on or off.
	Enabled bool
	// MaxDelay is the maximum interval between two presses for them to be treated
	// as a double press.
	MaxDelay time.Duration
}

// Reverse configures the modifier that inverts a press (go back instead of forward).
type Reverse struct {
	// Enabled turns the reverse modifier on or off.
	Enabled bool
	// Modifier is the key held to reverse a press (see ReverseModifiers).
	Modifier string
}

// GlobeTrigger is the trigger of a keyboard with a Globe key of its own, unless
// it is listed in Trigger.Keyboards.
var GlobeTrigger = KeyboardTrigger{Key: TriggerGlobe, Mode: TriggerModeFn}

// Trigger selects the key that switches input sources on each keyboard.
type Trigger struct {
	// Keyboards are the triggers of single keyboards, by keyboard ID (see
	// KeyboardID). Any other keyboard uses GlobeTrigger if it has a Globe key,
	// and has no trigger otherwise.
	Keyboards map[string]KeyboardTrigger
}

// For returns the trigger of the keyboard with the given ID, which has or lacks
// a Globe key, and whether it is listed in Keyboards. The zero KeyboardTrigger
// is no trigger.
func (t Trigger) For(id string, hasGlobe bool) (KeyboardTrigger, bool) {
	if trigger, ok := t.Keyboards[id]; ok {
		return trigger, true
	}

	if hasGlobe {
		return GlobeTrigger, false
	}

	return KeyboardTrigger{}, false
}

// KeyboardTrigger is one keyboard's trigger key and how it is detected.
type KeyboardTrigger struct {
	// Key is the trigger key (see TriggerKeys).
	Key string
	// Mode is how a key other than the Globe key is detected (see
	// TriggerModes). It has no effect for the Globe key.
	Mode string
}

// None reports whether there is no trigger key.
func (k KeyboardTrigger) None() bool {
	return k.Key == ""
}

// Globe reports whether the trigger is the Globe key itself.
func (k KeyboardTrigger) Globe() bool {
	return k.Key == TriggerGlobe
}

// other reports whether the trigger is a key other than the Globe key.
func (k KeyboardTrigger) other() bool {
	return !k.None() && !k.Globe()
}

// Remapped reports whether the trigger key is remapped at the HID level.
func (k KeyboardTrigger) Remapped() bool {
	return k.other() && k.Mode != TriggerModePassThrough
}

// Fn reports whether the trigger key is remapped to fn, so it arrives as the
// Globe key.
func (k KeyboardTrigger) Fn() bool {
	return k.other() && k.Mode == TriggerModeFn
}

// Dedicated reports whether the trigger key is silenced and read from raw input.
func (k KeyboardTrigger) Dedicated() bool {
	return k.other() && k.Mode == TriggerModeDedicated
}

// PassThrough reports whether the trigger key keeps its function and is read
// from raw input.
func (k KeyboardTrigger) PassThrough() bool {
	return k.other() && k.Mode == TriggerModePassThrough
}

// Raw reports whether the trigger key is read from the keyboard's raw input,
// which needs Input Monitoring.
func (k KeyboardTrigger) Raw() bool {
	return k.Dedicated() || k.PassThrough()
}

// HUD configures the on-screen overlay.
type HUD struct {
	// Enabled shows or hides the overlay.
	Enabled bool
	// Duration is how long the overlay stays fully visible before fading.
	Duration time.Duration
	// ShowCollection includes the collection name as a subtitle.
	ShowCollection bool
}

// Collection is a named, ordered set of input source IDs.
type Collection struct {
	// Name uniquely identifies the collection.
	Name string
	// Sources are the input source IDs, in cycling order.
	Sources []string
}

// Validate reports whether the configuration is internally consistent.
func (c Config) Validate() error {
	if c.DoublePress.MaxDelay <= 0 {
		return fmt.Errorf("double press max delay must be positive, got %s", c.DoublePress.MaxDelay)
	}

	if c.HUD.Duration <= 0 {
		return fmt.Errorf("hud duration must be positive, got %s", c.HUD.Duration)
	}

	if !slices.Contains(LogLevels, c.Logger.Level) {
		return fmt.Errorf("logger level must be one of %v, got %q", LogLevels, c.Logger.Level)
	}

	if !slices.Contains(ReverseModifiers, c.Reverse.Modifier) {
		return fmt.Errorf("reverse modifier must be one of %v, got %q", ReverseModifiers, c.Reverse.Modifier)
	}

	if err := c.Trigger.validate(); err != nil {
		return err
	}

	if c.Logger.Path == "" {
		return fmt.Errorf("logger path must not be empty")
	}

	if c.Logger.RetentionDays < 0 || c.Logger.RetentionFiles < 0 {
		return fmt.Errorf("logger retention values must not be negative")
	}

	return c.validateCollections()
}

// validate checks every keyboard's trigger.
func (t Trigger) validate() error {
	for id, trigger := range t.Keyboards {
		if !keyboardIDPattern.MatchString(id) {
			return fmt.Errorf("trigger keyboard %q must be vendor:product in hex, like 05ac:029c", id)
		}

		if err := trigger.validate(fmt.Sprintf("trigger for keyboard %s", id)); err != nil {
			return err
		}
	}

	return nil
}

// validate checks the trigger key and mode; name names the trigger in errors.
func (k KeyboardTrigger) validate(name string) error {
	if !slices.Contains(TriggerKeys, k.Key) {
		return fmt.Errorf("%s key must be one of %v, got %q", name, TriggerKeys, k.Key)
	}

	if !slices.Contains(TriggerModes, k.Mode) {
		return fmt.Errorf("%s mode must be one of %v, got %q", name, TriggerModes, k.Mode)
	}

	if k.Key == TriggerCapsLock && k.Mode == TriggerModePassThrough {
		return fmt.Errorf("%s key %q cannot be passed through", name, TriggerCapsLock)
	}

	return nil
}

// validateCollections checks that collection names are present and unique and that
// each collection holds non-empty, non-duplicate input sources.
func (c Config) validateCollections() error {
	seen := make(map[string]struct{}, len(c.Collections))

	for _, collection := range c.Collections {
		if collection.Name == "" {
			return fmt.Errorf("collection name must not be empty")
		}

		if _, ok := seen[collection.Name]; ok {
			return fmt.Errorf("duplicate collection name %q", collection.Name)
		}

		seen[collection.Name] = struct{}{}

		if len(collection.Sources) == 0 {
			return fmt.Errorf("collection %q must have at least one input source", collection.Name)
		}

		if err := validateSources(collection); err != nil {
			return err
		}
	}

	return nil
}

// validateSources reports whether a collection's input sources are non-empty and
// free of duplicates.
func validateSources(collection Collection) error {
	seen := make(map[string]struct{}, len(collection.Sources))

	for _, id := range collection.Sources {
		if id == "" {
			return fmt.Errorf("collection %q has an empty input source", collection.Name)
		}

		if _, ok := seen[id]; ok {
			return fmt.Errorf("collection %q lists input source %q more than once", collection.Name, id)
		}

		seen[id] = struct{}{}
	}

	return nil
}
