package config_test

import (
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
		Trigger: config.Trigger{
			Key:       config.TriggerGlobe,
			Mode:      config.TriggerModeFn,
			Keyboards: config.Keyboards{ApplyTo: config.KeyboardsWithoutGlobe},
		},
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

func TestValidateRejectsInvalidConfigs(t *testing.T) {
	tests := map[string]func(*config.Config){
		"non-positive delay":    func(c *config.Config) { c.DoublePress.MaxDelay = 0 },
		"non-positive duration": func(c *config.Config) { c.HUD.Duration = 0 },
		"invalid level":         func(c *config.Config) { c.Logger.Level = "trace" },
		"invalid modifier":      func(c *config.Config) { c.Reverse.Modifier = "fn" },
		"empty logger path":     func(c *config.Config) { c.Logger.Path = "" },
		"invalid trigger key":   func(c *config.Config) { c.Trigger.Key = "fn" },
		"empty trigger key":     func(c *config.Config) { c.Trigger.Key = "" },
		"invalid trigger mode":  func(c *config.Config) { c.Trigger.Mode = "swallow" },
		"caps lock pass-through": func(c *config.Config) {
			c.Trigger.Key, c.Trigger.Mode = config.TriggerCapsLock, config.TriggerModePassThrough
		},
		"dedicated key as dedicated trigger": func(c *config.Config) {
			c.Trigger.Key, c.Trigger.Mode = config.TriggerDedicatedKey, config.TriggerModeDedicated
		},
		"invalid keyboard scope": func(c *config.Config) { c.Trigger.Keyboards.ApplyTo = "apple" },
		"invalid keyboard id": func(c *config.Config) {
			c.Trigger.Keyboards.Overrides = map[string]bool{"05AC:29C": true}
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
		cfg.Trigger.Key = key

		if err := cfg.Validate(); err != nil {
			t.Errorf("trigger %q: unexpected error: %v", key, err)
		}
	}
}

func TestTriggerModes(t *testing.T) {
	tests := []struct {
		trigger                          config.Trigger
		remapped, dedicated, passThrough bool
	}{
		{config.Trigger{Key: config.TriggerGlobe, Mode: config.TriggerModeFn}, false, false, false},
		{config.Trigger{Key: config.TriggerGlobe, Mode: config.TriggerModeDedicated}, false, false, false},
		{config.Trigger{Key: config.TriggerGlobe, Mode: config.TriggerModePassThrough}, false, false, false},
		{config.Trigger{Key: config.TriggerCapsLock, Mode: config.TriggerModeFn}, true, false, false},
		{config.Trigger{Key: config.TriggerCapsLock, Mode: config.TriggerModeDedicated}, true, true, false},
		{config.Trigger{Key: "right_command", Mode: config.TriggerModePassThrough}, false, false, true},
	}

	for _, test := range tests {
		trigger := test.trigger
		if got := trigger.Remapped(); got != test.remapped {
			t.Errorf("%+v.Remapped() = %v, want %v", trigger, got, test.remapped)
		}

		if got := trigger.Dedicated(); got != test.dedicated {
			t.Errorf("%+v.Dedicated() = %v, want %v", trigger, got, test.dedicated)
		}

		if got := trigger.PassThrough(); got != test.passThrough {
			t.Errorf("%+v.PassThrough() = %v, want %v", trigger, got, test.passThrough)
		}
	}
}

func TestKeyboardsIncludes(t *testing.T) {
	withoutGlobe := config.Keyboards{
		ApplyTo:   config.KeyboardsWithoutGlobe,
		Overrides: map[string]bool{"05ac:029c": true, "046d:c33f": false},
	}
	all := config.Keyboards{ApplyTo: config.KeyboardsAll, Overrides: map[string]bool{"046d:c33f": false}}

	tests := []struct {
		name      string
		keyboards config.Keyboards
		id        string
		hasGlobe  bool
		want      bool
	}{
		{"without globe: third-party", withoutGlobe, "1234:5678", false, true},
		{"without globe: apple", withoutGlobe, "05ac:0281", true, false},
		{"without globe: apple included", withoutGlobe, "05ac:029c", true, true},
		{"without globe: third-party excluded", withoutGlobe, "046d:c33f", false, false},
		{"all: apple", all, "05ac:0281", true, true},
		{"all: excluded", all, "046d:c33f", false, false},
	}

	for _, test := range tests {
		if got := test.keyboards.Includes(test.id, test.hasGlobe); got != test.want {
			t.Errorf("%s: Includes = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestKeyboardID(t *testing.T) {
	if got := config.KeyboardID(0x5ac, 0x29c); got != "05ac:029c" {
		t.Errorf("KeyboardID = %q, want 05ac:029c", got)
	}
}
