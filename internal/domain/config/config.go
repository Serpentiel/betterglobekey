// Package config defines the application's configuration model and its validation rules.
//
// The types here are pure: they describe a fully-resolved configuration and know
// nothing about how it is stored, parsed, or migrated. Loading and migrating
// on-disk configuration is the responsibility of the infrastructure layer.
package config

import (
	"fmt"
	"slices"
	"time"
)

// LogLevels are the accepted logger levels, from most to least verbose.
var LogLevels = []string{"debug", "info", "warn", "error"}

// ReverseModifiers are the accepted modifier keys for a reverse press.
var ReverseModifiers = []string{"shift", "option", "control", "command"}

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

	if c.Logger.Path == "" {
		return fmt.Errorf("logger path must not be empty")
	}

	if c.Logger.RetentionDays < 0 || c.Logger.RetentionFiles < 0 {
		return fmt.Errorf("logger retention values must not be negative")
	}

	return c.validateCollections()
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
