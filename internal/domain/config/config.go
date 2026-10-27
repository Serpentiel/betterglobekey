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

// TriggerGlobe is the default trigger key: the Globe (fn) key itself.
const TriggerGlobe = "globe"

// TriggerCapsLock is the Caps Lock key, which can only be remapped, not passed through.
const TriggerCapsLock = "caps_lock"

// TriggerDedicatedKey is the key a dedicated trigger is remapped to. F20 has a
// key code but no Apple keyboard has the key and macOS binds nothing to it;
// F21-F24 would be rarer still, but macOS defines no key code for them.
const TriggerDedicatedKey = "f20"

// TriggerKeys are the accepted trigger keys.
var TriggerKeys = []string{
	TriggerGlobe,
	TriggerCapsLock,
	"right_command", "right_option", "right_control", "right_shift",
	"left_command", "left_option", "left_control", "left_shift",
	"f13", "f14", "f15", "f16", "f17", "f18", "f19", "f20",
	"menu",
}

// Trigger modes: how a trigger key other than Globe is detected.
const (
	// TriggerModeGlobe remaps the key to Globe, so it gains Globe's shortcuts and
	// Globe keeps switching too.
	TriggerModeGlobe = "globe"
	// TriggerModeDedicated remaps the key to TriggerDedicatedKey, so it is the
	// only key that switches and Globe keeps just its macOS functions.
	TriggerModeDedicated = "dedicated"
	// TriggerModePassThrough keeps the key's normal function and watches it
	// directly, switching only when it is tapped on its own.
	TriggerModePassThrough = "pass_through"
)

// TriggerModes are the accepted trigger modes.
var TriggerModes = []string{TriggerModeGlobe, TriggerModeDedicated, TriggerModePassThrough}

// Keyboard scopes: which keyboards a remapped trigger key applies to.
const (
	// KeyboardsWithoutGlobe remaps the key on keyboards that lack a Globe key.
	KeyboardsWithoutGlobe = "without_globe"
	// KeyboardsAll remaps the key on every keyboard.
	KeyboardsAll = "all"
)

// KeyboardScopes are the accepted keyboard scopes.
var KeyboardScopes = []string{KeyboardsWithoutGlobe, KeyboardsAll}

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
	// Trigger selects the key that switches input sources.
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

// Trigger selects the key that switches input sources.
type Trigger struct {
	// Key is the trigger key (see TriggerKeys).
	Key string
	// Mode is how a key other than Globe is detected (see TriggerModes). It has
	// no effect for the Globe key.
	Mode string
	// Keyboards selects the keyboards a remapped key applies to.
	Keyboards Keyboards
}

// Keyboards selects the keyboards a remapped trigger key applies to.
type Keyboards struct {
	// ApplyTo is the default scope (see KeyboardScopes).
	ApplyTo string
	// Overrides include (true) or exclude (false) keyboards by keyboard ID
	// (see KeyboardID), whatever ApplyTo says.
	Overrides map[string]bool
}

// Includes reports whether the keyboard with the given ID, which has or lacks a
// Globe key, is in scope.
func (k Keyboards) Includes(id string, hasGlobe bool) bool {
	if include, ok := k.Overrides[id]; ok {
		return include
	}

	return k.ApplyTo == KeyboardsAll || !hasGlobe
}

// Remapped reports whether the trigger key is remapped at the HID level.
func (t Trigger) Remapped() bool {
	return t.Key != TriggerGlobe && t.Mode != TriggerModePassThrough
}

// Dedicated reports whether the trigger key is remapped to TriggerDedicatedKey,
// leaving Globe out of switching.
func (t Trigger) Dedicated() bool {
	return t.Key != TriggerGlobe && t.Mode == TriggerModeDedicated
}

// PassThrough reports whether the trigger key keeps its function and is watched directly.
func (t Trigger) PassThrough() bool {
	return t.Key != TriggerGlobe && t.Mode == TriggerModePassThrough
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

// validate checks the trigger key, mode, and keyboard scope.
func (t Trigger) validate() error {
	if !slices.Contains(TriggerKeys, t.Key) {
		return fmt.Errorf("trigger key must be one of %v, got %q", TriggerKeys, t.Key)
	}

	if !slices.Contains(TriggerModes, t.Mode) {
		return fmt.Errorf("trigger mode must be one of %v, got %q", TriggerModes, t.Mode)
	}

	if t.Key == TriggerCapsLock && t.Mode == TriggerModePassThrough {
		return fmt.Errorf("trigger key %q cannot be passed through", TriggerCapsLock)
	}

	if t.Key == TriggerDedicatedKey && t.Mode == TriggerModeDedicated {
		return fmt.Errorf("trigger key %q is the dedicated key itself; pass it through instead", TriggerDedicatedKey)
	}

	if !slices.Contains(KeyboardScopes, t.Keyboards.ApplyTo) {
		return fmt.Errorf("trigger keyboards apply_to must be one of %v, got %q", KeyboardScopes, t.Keyboards.ApplyTo)
	}

	for id := range t.Keyboards.Overrides {
		if !keyboardIDPattern.MatchString(id) {
			return fmt.Errorf("trigger keyboard override %q must be vendor:product in hex, like 05ac:029c", id)
		}
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
