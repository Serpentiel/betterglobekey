package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadMissingFileErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml")); err == nil {
		t.Fatal("Load accepted a missing file")
	}
}

func TestWriteDefaultWithoutSources(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := WriteDefault(path, nil); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.Contains(string(data), "version: 2") {
		t.Errorf("default file missing version; got:\n%s", data)
	}

	if strings.Contains(string(data), "name: primary") {
		t.Errorf("expected no seeded collection without sources; got:\n%s", data)
	}
}

func TestWriteWithBackupErrors(t *testing.T) {
	t.Run("backup write fails", func(t *testing.T) {
		// The parent directory does not exist, so writing the backup fails first.
		path := filepath.Join(t.TempDir(), "missing", "config.yaml")
		if err := writeWithBackup(path, []byte("old"), []byte("new"), 1); err == nil {
			t.Fatal("writeWithBackup succeeded despite an unwritable backup path")
		}
	})

	t.Run("migrated write fails", func(t *testing.T) {
		// path is itself a directory: the backup (path + suffix) writes fine, but
		// writing the migrated document over a directory fails.
		dir := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}

		if err := writeWithBackup(dir, []byte("old"), []byte("new"), 1); err == nil {
			t.Fatal("writeWithBackup succeeded despite an unwritable target path")
		}
	})
}

func TestToDomainRejectsInvalidHUDDuration(t *testing.T) {
	var schema schemaV2

	schema.DoublePress.MaximumDelay = "250ms"
	schema.HUD.Duration = "not-a-duration"

	if _, err := toDomain(schema); err == nil {
		t.Fatal("toDomain accepted an invalid hud.duration")
	}
}

func TestApplySchemaDefaults(t *testing.T) {
	var empty schemaV2
	applySchemaDefaults(&empty)

	if empty.Logger.Level != defaultLogLevel {
		t.Errorf("level = %q, want %q", empty.Logger.Level, defaultLogLevel)
	}

	if empty.DoublePress.MaximumDelay != defaultDoublePressDelay {
		t.Errorf("delay = %q, want %q", empty.DoublePress.MaximumDelay, defaultDoublePressDelay)
	}

	if empty.Reverse.Modifier != defaultReverseModifier {
		t.Errorf("modifier = %q, want %q", empty.Reverse.Modifier, defaultReverseModifier)
	}

	if empty.HUD.Duration != defaultHUDDuration {
		t.Errorf("duration = %q, want %q", empty.HUD.Duration, defaultHUDDuration)
	}

	// Populated fields must be left untouched.
	set := schemaV2{}
	set.Logger.Level = "debug"
	set.DoublePress.MaximumDelay = "1s"
	set.Reverse.Modifier = "command"
	set.HUD.Duration = "2s"
	applySchemaDefaults(&set)

	if set.Logger.Level != "debug" || set.DoublePress.MaximumDelay != "1s" ||
		set.Reverse.Modifier != "command" || set.HUD.Duration != "2s" {
		t.Errorf("applySchemaDefaults overwrote populated fields: %+v", set)
	}
}

func TestDefaultLogPath(t *testing.T) {
	t.Run("uses home directory", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		want := filepath.Join(home, "Library", "Logs", "betterglobekey.log")
		if got := defaultLogPath(); got != want {
			t.Errorf("defaultLogPath() = %q, want %q", got, want)
		}
	})

	t.Run("falls back when home is unavailable", func(t *testing.T) {
		t.Setenv("HOME", "")

		if got := defaultLogPath(); got != "betterglobekey.log" {
			t.Errorf("defaultLogPath() = %q, want the relative fallback", got)
		}
	})
}

func TestDetectVersionMalformedYAML(t *testing.T) {
	if _, err := detectVersion([]byte("{ this is not: valid: yaml")); err == nil {
		t.Fatal("detectVersion accepted malformed YAML")
	}
}

func TestMigratePropagatesDetectError(t *testing.T) {
	if _, _, err := migrate([]byte("{ not: valid: yaml")); err == nil {
		t.Fatal("migrate accepted malformed YAML")
	}
}

func TestMigrateV1ToV2RejectsInvalidBody(t *testing.T) {
	// A scalar where the v1 schema expects a map of input sources.
	if _, err := migrateV1ToV2([]byte("input_sources: not-a-map\n")); err == nil {
		t.Fatal("migrateV1ToV2 accepted an invalid v1 body")
	}
}

func TestSaveWriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "config.yaml")

	if err := Save(path, sampleConfig()); err == nil {
		t.Fatal("Save succeeded despite an unwritable path")
	}
}

func TestWatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte("version: 2\n"), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	changed := make(chan struct{}, 8)
	done := make(chan error, 1)

	go func() { done <- Watch(ctx, path, func() { changed <- struct{}{} }) }()

	if !awaitChange(t, path, changed) {
		t.Fatal("onChange was not called after modifying the watched file")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Watch returned %v, want nil after cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Watch did not return after context cancellation")
	}
}

// awaitChange repeatedly rewrites path until the watcher reports a change or the
// deadline elapses. Rewriting in a loop avoids racing the watcher's registration.
func awaitChange(t *testing.T, path string, changed <-chan struct{}) bool {
	t.Helper()

	deadline := time.After(5 * time.Second)

	// Space rewrites wider than the watcher's debounce so a quiet gap lets the
	// coalesced onChange fire; continuous writes would keep resetting the timer.
	tick := time.NewTicker(300 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-changed:
			return true
		case <-deadline:
			return false
		case <-tick.C:
			_ = os.WriteFile(path, []byte("version: 2\n# touched\n"), 0o600)
		}
	}
}
