package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	"gopkg.in/yaml.v3"
)

const v1Doc = `logger:
  path: y.log
  retain:
    days: 10
    copies: 5
double_press:
  maximum_delay: 200
input_sources:
  zeta: [z1]
  alpha: [a1, a2]
`

const v2Doc = `version: 2
logger:
  path: x.log
  retention:
    days: 7
    files: 2
double_press:
  maximum_delay: 300ms
collections:
  - name: primary
    sources: [a, b]
  - name: coding
    sources: [c]
`

func TestDetectVersion(t *testing.T) {
	if v, _ := detectVersion([]byte(v1Doc)); v != 1 {
		t.Fatalf("unversioned doc detected as v%d, want 1", v)
	}

	if v, _ := detectVersion([]byte(v2Doc)); v != 2 {
		t.Fatalf("versioned doc detected as v%d, want 2", v)
	}
}

func TestMigrateV1ToV2(t *testing.T) {
	out, from, err := migrate([]byte(v1Doc))
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if from != 1 {
		t.Fatalf("from = %d, want 1", from)
	}

	cfg, err := toDomain(parse(t, out))
	if err != nil {
		t.Fatalf("toDomain: %v", err)
	}

	if cfg.DoublePressMaxDelay != 200*time.Millisecond {
		t.Errorf("delay = %s, want 200ms", cfg.DoublePressMaxDelay)
	}

	if cfg.Logger != (domainconfig.Logger{Path: "y.log", RetentionDays: 10, RetentionFiles: 5}) {
		t.Errorf("logger = %+v", cfg.Logger)
	}

	// v1 maps are unordered, so migration must emit collections alphabetically.
	names := []string{cfg.Collections[0].Name, cfg.Collections[1].Name}
	if !slices.Equal(names, []string{"alpha", "zeta"}) {
		t.Errorf("collection order = %v, want [alpha zeta]", names)
	}

	if !slices.Equal(cfg.Collections[0].Sources, []string{"a1", "a2"}) {
		t.Errorf("alpha sources = %v", cfg.Collections[0].Sources)
	}
}

func TestMigrateAlreadyCurrentIsNoOp(t *testing.T) {
	_, from, err := migrate([]byte(v2Doc))
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if from != 2 {
		t.Fatalf("from = %d, want 2", from)
	}
}

func TestMigrateRejectsNewerVersion(t *testing.T) {
	if _, _, err := migrate([]byte("version: 99\n")); err == nil {
		t.Fatal("expected error for newer-than-supported version")
	}
}

func TestLoadV2(t *testing.T) {
	path := write(t, "config.yaml", v2Doc)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.DoublePressMaxDelay != 300*time.Millisecond {
		t.Errorf("delay = %s, want 300ms", cfg.DoublePressMaxDelay)
	}

	if len(cfg.Collections) != 2 || cfg.Collections[1].Name != "coding" {
		t.Errorf("collections = %+v", cfg.Collections)
	}
}

func TestLoadMigratesV1AndWritesBackup(t *testing.T) {
	path := write(t, "config.yaml", v1Doc)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.DoublePressMaxDelay != 200*time.Millisecond {
		t.Errorf("delay = %s, want 200ms", cfg.DoublePressMaxDelay)
	}

	// The on-disk file must have been upgraded to the current version.
	data, _ := os.ReadFile(path)
	if v, _ := detectVersion(data); v != currentVersion {
		t.Errorf("on-disk version = %d, want %d", v, currentVersion)
	}

	// The original must be preserved as a versioned backup.
	backup, err := os.ReadFile(path + ".v1.bak")
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}

	if string(backup) != v1Doc {
		t.Error("backup does not match original v1 document")
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	tests := map[string]string{
		"bad duration": `version: 2
logger: {path: x, retention: {days: 1, files: 1}}
double_press: {maximum_delay: nope}
`,
		"empty path": `version: 2
logger: {path: "", retention: {days: 1, files: 1}}
double_press: {maximum_delay: 250ms}
`,
	}

	for name, doc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(write(t, "c.yaml", doc)); err == nil {
				t.Fatalf("expected error for %s", name)
			}
		})
	}
}

func TestWriteDefaultRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := WriteDefault(path, []string{"s1", "s2"}); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Collections) != 1 || cfg.Collections[0].Name != "primary" {
		t.Fatalf("collections = %+v", cfg.Collections)
	}

	if !slices.Equal(cfg.Collections[0].Sources, []string{"s1", "s2"}) {
		t.Errorf("sources = %v", cfg.Collections[0].Sources)
	}

	if cfg.DoublePressMaxDelay != 250*time.Millisecond {
		t.Errorf("delay = %s, want 250ms", cfg.DoublePressMaxDelay)
	}
}

func parse(t *testing.T, data []byte) schemaV2 {
	t.Helper()

	var schema schemaV2
	if err := yaml.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parse: %v", err)
	}

	return schema
}

func write(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	return path
}
