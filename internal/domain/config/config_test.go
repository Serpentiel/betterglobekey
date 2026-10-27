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
		Trigger:     config.Trigger{Key: config.TriggerGlobe},
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
		"caps lock pass-through": func(c *config.Config) {
			c.Trigger = config.Trigger{Key: config.TriggerCapsLock, PassThrough: true}
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

func TestTriggerRemapped(t *testing.T) {
	tests := []struct {
		trigger config.Trigger
		want    bool
	}{
		{config.Trigger{Key: config.TriggerGlobe}, false},
		{config.Trigger{Key: config.TriggerGlobe, PassThrough: true}, false},
		{config.Trigger{Key: config.TriggerCapsLock}, true},
		{config.Trigger{Key: "right_command"}, true},
		{config.Trigger{Key: "right_command", PassThrough: true}, false},
	}

	for _, test := range tests {
		if got := test.trigger.Remapped(); got != test.want {
			t.Errorf("%+v.Remapped() = %v, want %v", test.trigger, got, test.want)
		}
	}
}
