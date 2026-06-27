package config_test

import (
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
)

func validConfig() config.Config {
	return config.Config{
		Logger:              config.Logger{Path: "betterglobekey.log", RetentionDays: 30, RetentionFiles: 3},
		DoublePressMaxDelay: 250 * time.Millisecond,
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
		"non-positive delay":    func(c *config.Config) { c.DoublePressMaxDelay = 0 },
		"empty logger path":     func(c *config.Config) { c.Logger.Path = "" },
		"negative retention":    func(c *config.Config) { c.Logger.RetentionFiles = -1 },
		"empty collection name": func(c *config.Config) { c.Collections[0].Name = "" },
		"collection no sources": func(c *config.Config) { c.Collections[0].Sources = nil },
		"duplicate collections": func(c *config.Config) { c.Collections = append(c.Collections, c.Collections[0]) },
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
