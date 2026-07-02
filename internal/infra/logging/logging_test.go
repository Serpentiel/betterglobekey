package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]zapcore.Level{
		"debug":    zapcore.DebugLevel,
		"info":     zapcore.InfoLevel,
		"warn":     zapcore.WarnLevel,
		"error":    zapcore.ErrorLevel,
		"":         zapcore.InfoLevel,
		"nonsense": zapcore.InfoLevel,
		"INFO":     zapcore.InfoLevel, // case-sensitive: unknown -> default
	}

	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Errorf("parseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNewWritesStructuredEntriesToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	logger := New(config.Logger{Path: path, Level: "info", RetentionDays: 1, RetentionFiles: 1})

	logger.Info("started", "port", 8080)
	logger.Error("boom", "reason", "kaboom")

	if err := logger.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	out := readFile(t, path)
	for _, want := range []string{"started", "port", "8080", "boom", "reason", "kaboom"} {
		if !strings.Contains(out, want) {
			t.Errorf("log file missing %q; got:\n%s", want, out)
		}
	}

	if !strings.Contains(out, "{") {
		t.Errorf("file output is not JSON-structured; got:\n%s", out)
	}
}

func TestNewRespectsConfiguredLevel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	logger := New(config.Logger{Path: path, Level: "error"})

	logger.Info("below-threshold")
	logger.Error("at-threshold")
	_ = logger.Close()

	out := readFile(t, path)
	if strings.Contains(out, "below-threshold") {
		t.Errorf("info entry logged despite error level; got:\n%s", out)
	}

	if !strings.Contains(out, "at-threshold") {
		t.Errorf("error entry missing; got:\n%s", out)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log %q: %v", path, err)
	}

	return string(data)
}
