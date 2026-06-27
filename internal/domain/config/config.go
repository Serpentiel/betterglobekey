// Package config defines the application's configuration model and its validation rules.
//
// The types here are pure: they describe a fully-resolved configuration and know
// nothing about how it is stored, parsed, or migrated. Loading and migrating
// on-disk configuration is the responsibility of the infrastructure layer.
package config

import (
	"fmt"
	"time"
)

// Config is the fully-resolved application configuration.
type Config struct {
	// Logger holds the logging configuration.
	Logger Logger
	// DoublePressMaxDelay is the maximum interval between two Globe key presses
	// for them to be treated as a double press.
	DoublePressMaxDelay time.Duration
	// Collections are the ordered input source collections the Globe key cycles through.
	Collections []Collection
}

// Logger holds the logging configuration.
type Logger struct {
	// Path is the path to the log file.
	Path string
	// RetentionDays is the number of days to retain rotated log files.
	RetentionDays int
	// RetentionFiles is the number of rotated log files to retain.
	RetentionFiles int
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
	if c.DoublePressMaxDelay <= 0 {
		return fmt.Errorf("double press max delay must be positive, got %s", c.DoublePressMaxDelay)
	}

	if c.Logger.Path == "" {
		return fmt.Errorf("logger path must not be empty")
	}

	if c.Logger.RetentionDays < 0 || c.Logger.RetentionFiles < 0 {
		return fmt.Errorf("logger retention values must not be negative")
	}

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
	}

	return nil
}
