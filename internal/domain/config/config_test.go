package config_test

import (
	"slices"
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

func validConfig() config.Config {
	return config.Config{
		Logger:      config.Logger{Path: "betterglobekey.log", Level: "info", RetentionDays: 30, RetentionFiles: 3},
		DoublePress: config.DoublePress{Enabled: true, MaxDelay: 250 * time.Millisecond},
		Reverse:     config.Reverse{Enabled: true, Modifier: "shift"},
		HUD:         config.HUD{Enabled: true, Duration: 900 * time.Millisecond, ShowCollection: true},
		Collections: []config.Collection{
			{Name: "primary", Sources: []string{"com.apple.keylayout.US"}},
		},
	}
}

func TestValidateAcceptsValidConfig(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAllowsNoCollections(t *testing.T) {
	cfg := validConfig()
	cfg.Collections = nil

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// keyboardTrigger returns a trigger setting key and mode for one keyboard.
func keyboardTrigger(key, mode string) config.Trigger {
	return config.Trigger{Keyboards: map[string][]config.KeyboardTrigger{"046d:c33f": {{Key: key, Mode: mode}}}}
}

func TestValidateRejectsInvalidConfigs(t *testing.T) {
	tests := map[string]func(*config.Config){
		"non-positive delay":    func(c *config.Config) { c.DoublePress.MaxDelay = 0 },
		"non-positive duration": func(c *config.Config) { c.HUD.Duration = 0 },
		"invalid level":         func(c *config.Config) { c.Logger.Level = "trace" },
		"invalid modifier":      func(c *config.Config) { c.Reverse.Modifier = "fn" },
		"empty logger path":     func(c *config.Config) { c.Logger.Path = "" },
		"invalid trigger key":   func(c *config.Config) { c.Trigger = keyboardTrigger("fn", "dedicated") },
		"empty trigger key":     func(c *config.Config) { c.Trigger = keyboardTrigger("", "dedicated") },
		"invalid trigger mode":  func(c *config.Config) { c.Trigger = keyboardTrigger("insert", "swallow") },
		"fn trigger mode":       func(c *config.Config) { c.Trigger = keyboardTrigger("insert", "fn") },
		"caps lock pass-through": func(c *config.Config) {
			c.Trigger = keyboardTrigger(config.TriggerCapsLock, config.TriggerModePassThrough)
		},
		"invalid keyboard id": func(c *config.Config) {
			c.Trigger.Keyboards = map[string][]config.KeyboardTrigger{"05AC:29C": {config.GlobeTrigger}}
		},
		"duplicate trigger key": func(c *config.Config) {
			insert := config.KeyboardTrigger{Key: "insert", Mode: config.TriggerModeDedicated}
			c.Trigger.Keyboards = map[string][]config.KeyboardTrigger{"046d:c33f": {insert, insert}}
		},
		"negative retention":    func(c *config.Config) { c.Logger.RetentionFiles = -1 },
		"empty collection name": func(c *config.Config) { c.Collections[0].Name = "" },
		"collection no sources": func(c *config.Config) { c.Collections[0].Sources = nil },
		"duplicate collections": func(c *config.Config) { c.Collections = append(c.Collections, c.Collections[0]) },
		"empty source":          func(c *config.Config) { c.Collections[0].Sources = []string{""} },
		"duplicate sources":     func(c *config.Config) { c.Collections[0].Sources = []string{"a", "a"} },
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validConfig()
			mutate(&cfg)

			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected validation error for %s", name)
			}
		})
	}
}

func TestValidateAcceptsEveryTriggerKey(t *testing.T) {
	for _, key := range config.TriggerKeys {
		cfg := validConfig()
		cfg.Trigger = keyboardTrigger(key, config.TriggerModeDedicated)

		if err := cfg.Validate(); err != nil {
			t.Errorf("trigger %q: unexpected error: %v", key, err)
		}
	}
}

func TestValidateAcceptsSeveralTriggerKeys(t *testing.T) {
	cfg := validConfig()
	cfg.Trigger.Keyboards = map[string][]config.KeyboardTrigger{
		"05ac:0281": {
			config.GlobeTrigger,
			{Key: config.TriggerCapsLock, Mode: config.TriggerModeDedicated},
			{Key: "left_command", Mode: config.TriggerModePassThrough},
		},
		"046d:c33f": {},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKeyboardTriggerModes(t *testing.T) {
	key := func(key, mode string) config.KeyboardTrigger { return config.KeyboardTrigger{Key: key, Mode: mode} }

	tests := []struct {
		trigger                               config.KeyboardTrigger
		remapped, dedicated, passThrough, raw bool
	}{
		{config.KeyboardTrigger{}, false, false, false, false},
		{config.GlobeTrigger, false, false, false, false},
		{key(config.TriggerCapsLock, "fn"), false, false, false, false},
		{key(config.TriggerCapsLock, config.TriggerModeDedicated), true, true, false, true},
		{key("insert", config.TriggerModePassThrough), false, false, true, true},
	}

	for _, test := range tests {
		trigger := test.trigger
		got := []bool{trigger.Remapped(), trigger.Dedicated(), trigger.PassThrough(), trigger.Raw()}
		want := []bool{test.remapped, test.dedicated, test.passThrough, test.raw}

		if !slices.Equal(got, want) {
			t.Errorf("%+v: remapped, dedicated, pass-through, raw = %v, want %v", trigger, got, want)
		}
	}
}

func TestTriggerFor(t *testing.T) {
	insert := config.KeyboardTrigger{Key: "insert", Mode: config.TriggerModePassThrough}
	trigger := config.Trigger{Keyboards: map[string][]config.KeyboardTrigger{
		"046d:c33f": {insert},
		"05ac:0342": {config.GlobeTrigger, insert},
		"05ac:0343": {},
	}}

	tests := []struct {
		name     string
		id       string
		hasGlobe bool
		want     []config.KeyboardTrigger
		listed   bool
	}{
		{"listed", "046d:c33f", false, []config.KeyboardTrigger{insert}, true},
		{"listed with the Globe key", "05ac:0342", true, []config.KeyboardTrigger{config.GlobeTrigger, insert}, true},
		{"listed without keys", "05ac:0343", true, []config.KeyboardTrigger{}, true},
		{"with Globe key", "05ac:0281", true, []config.KeyboardTrigger{config.GlobeTrigger}, false},
		{"without Globe key", "1234:5678", false, nil, false},
	}

	for _, test := range tests {
		got, listed := trigger.For(test.id, test.hasGlobe)
		if !slices.Equal(got, test.want) || listed != test.listed {
			t.Errorf("%s: For = %+v, %v; want %+v, %v", test.name, got, listed, test.want, test.listed)
		}
	}
}

func TestSwitchesWithGlobe(t *testing.T) {
	insert := config.KeyboardTrigger{Key: "insert", Mode: config.TriggerModeDedicated}

	if config.SwitchesWithGlobe([]config.KeyboardTrigger{insert}) || config.SwitchesWithGlobe(nil) {
		t.Error("SwitchesWithGlobe without the Globe key = true")
	}

	if !config.SwitchesWithGlobe([]config.KeyboardTrigger{insert, config.GlobeTrigger}) {
		t.Error("SwitchesWithGlobe with the Globe key = false")
	}
}

func TestKeyboardID(t *testing.T) {
	if got := config.KeyboardID(0x5ac, 0x29c); got != "05ac:029c" {
		t.Errorf("KeyboardID = %q, want 05ac:029c", got)
	}
}
