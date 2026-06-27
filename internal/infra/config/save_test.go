package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
)

func sampleConfig() domainconfig.Config {
	return domainconfig.Config{
		Logger:              domainconfig.Logger{Path: "x.log", RetentionDays: 7, RetentionFiles: 2},
		DoublePressMaxDelay: 300 * time.Millisecond,
		HUD:                 true,
		Collections: []domainconfig.Collection{
			{Name: "primary", Sources: []string{"a", "b"}},
			{Name: "coding", Sources: []string{"c"}},
		},
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	want := sampleConfig()
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.DoublePressMaxDelay != want.DoublePressMaxDelay {
		t.Errorf("delay = %s, want %s", got.DoublePressMaxDelay, want.DoublePressMaxDelay)
	}

	if got.HUD != want.HUD || got.Logger != want.Logger {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if len(got.Collections) != len(want.Collections) {
		t.Fatalf("collections = %d, want %d", len(got.Collections), len(want.Collections))
	}

	for i, collection := range want.Collections {
		if got.Collections[i].Name != collection.Name {
			t.Errorf("collection %d name = %q, want %q", i, got.Collections[i].Name, collection.Name)
		}
	}
}

func TestSaveWritesHeaderAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := Save(path, sampleConfig()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.HasPrefix(string(data), defaultHeader) {
		t.Errorf("saved file is missing the default header")
	}

	if !strings.Contains(string(data), "version: 2") {
		t.Errorf("saved file is missing the schema version")
	}
}

func TestSaveRejectsInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	invalid := sampleConfig()
	invalid.DoublePressMaxDelay = 0

	if err := Save(path, invalid); err == nil {
		t.Fatal("Save accepted an invalid config")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Save wrote a file for an invalid config")
	}
}
