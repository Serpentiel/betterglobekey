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
	defaultLogLevel         = "info"
	defaultDoublePressDelay = "250ms"
	defaultReverseModifier  = "shift"
	defaultHUDDuration      = "900ms"
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

// schemaV2 is the current configuration format. Collections are an ordered list,
// durations are Go duration strings, and the behavior groups expose their
// options. Boolean toggles are pointers so an absent value can default to true.
type schemaV2 struct {
	Version int `yaml:"version"`
	Logger  struct {
		Path      string `yaml:"path"`
		Level     string `yaml:"level"`
		Retention struct {
			Days  int `yaml:"days"`
			Files int `yaml:"files"`
		} `yaml:"retention"`
	} `yaml:"logger"`
	DoublePress struct {
		Enabled      *bool  `yaml:"enabled"`
		MaximumDelay string `yaml:"maximum_delay"`
	} `yaml:"double_press"`
	Reverse struct {
		Enabled  *bool  `yaml:"enabled"`
		Modifier string `yaml:"modifier"`
	} `yaml:"reverse"`
	HUD struct {
		Enabled        *bool  `yaml:"enabled"`
		Duration       string `yaml:"duration"`
		ShowCollection *bool  `yaml:"show_collection"`
	} `yaml:"hud"`
	Collections []collectionV2 `yaml:"collections"`
}

// collectionV2 is a named, ordered set of input source IDs in the v2 schema.
type collectionV2 struct {
	Name    string   `yaml:"name"`
	Sources []string `yaml:"sources"`
}

// boolPtr returns a pointer to b, for optional boolean schema fields.
func boolPtr(b bool) *bool {
	return &b
}

// boolOr dereferences p, falling back to def when p is nil.
func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}

	return *p
}
