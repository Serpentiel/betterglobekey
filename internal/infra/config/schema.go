// Package config loads, migrates, and persists the on-disk configuration file,
// translating it into the pure domain configuration model.
package config

// Configuration schema versions.
const (
	versionV1      = 1
	versionV2      = 2
	currentVersion = versionV2
)

// Defaults applied when generating a fresh configuration file.
const (
	defaultRetentionDays    = 30
	defaultRetentionFiles   = 3
	defaultDoublePressDelay = "250ms"
	defaultCollectionName   = "primary"
)

// schemaV1 is the legacy, unversioned configuration format. Input sources were
// stored as an unordered map and the double-press delay as an integer of
// milliseconds.
type schemaV1 struct {
	Logger struct {
		Path   string `yaml:"path"`
		Retain struct {
			Days   int `yaml:"days"`
			Copies int `yaml:"copies"`
		} `yaml:"retain"`
	} `yaml:"logger"`
	DoublePress struct {
		MaximumDelay int `yaml:"maximum_delay"`
	} `yaml:"double_press"`
	InputSources map[string][]string `yaml:"input_sources"`
}

// schemaV2 is the current configuration format. Collections are an ordered list
// and the double-press delay is a Go duration string (e.g. "250ms").
type schemaV2 struct {
	Version int `yaml:"version"`
	Logger  struct {
		Path      string `yaml:"path"`
		Retention struct {
			Days  int `yaml:"days"`
			Files int `yaml:"files"`
		} `yaml:"retention"`
	} `yaml:"logger"`
	DoublePress struct {
		MaximumDelay string `yaml:"maximum_delay"`
	} `yaml:"double_press"`
	// HUD is a pointer so an absent value can default to enabled.
	HUD         *bool          `yaml:"hud"`
	Collections []collectionV2 `yaml:"collections"`
}

// collectionV2 is a named, ordered set of input source IDs in the v2 schema.
type collectionV2 struct {
	Name    string   `yaml:"name"`
	Sources []string `yaml:"sources"`
}
