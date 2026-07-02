package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/app"
	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	"github.com/Serpentiel/betterglobekey/internal/domain/switcher"
	"github.com/Serpentiel/betterglobekey/internal/infra/inputsource"
	"github.com/Serpentiel/betterglobekey/internal/infra/logging"
	"github.com/spf13/cobra"
)

// validConfig is a loadable v2 document whose lone source is deliberately absent
// from any real system, so doctor exercises the unavailable-source warning.
const validConfig = `version: 2
logger:
  path: /tmp/betterglobekey-test.log
  level: info
  retention: {days: 7, files: 2}
double_press: {enabled: true, maximum_delay: 250ms}
reverse: {enabled: true, modifier: shift}
hud: {enabled: true, duration: 900ms, show_collection: true}
collections:
  - name: primary
    sources: [com.example.definitely.not.a.real.source]
`

func cmdWithConfig(t *testing.T, value string) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("config", "", "path to the config file")

	if value != "" {
		if err := cmd.Flags().Set("config", value); err != nil {
			t.Fatalf("set config flag: %v", err)
		}
	}

	return cmd
}

func TestResolveConfigPath(t *testing.T) {
	t.Run("explicit flag wins", func(t *testing.T) {
		got, err := resolveConfigPath(cmdWithConfig(t, "/etc/custom.yaml"))
		if err != nil {
			t.Fatalf("resolveConfigPath: %v", err)
		}

		if got != "/etc/custom.yaml" {
			t.Errorf("path = %q, want /etc/custom.yaml", got)
		}
	})

	t.Run("defaults to home", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		got, err := resolveConfigPath(cmdWithConfig(t, ""))
		if err != nil {
			t.Fatalf("resolveConfigPath: %v", err)
		}

		if want := filepath.Join(home, configFileName); got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
	})
}

func TestResolveSocketPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := resolveSocketPath()
	if err != nil {
		t.Fatalf("resolveSocketPath: %v", err)
	}

	if want := filepath.Join(home, socketFileName); got != want {
		t.Errorf("socket = %q, want %q", got, want)
	}
}

func TestReportUnavailableSources(t *testing.T) {
	cfg := domainconfig.Config{
		Collections: []domainconfig.Collection{
			{Name: "primary", Sources: []string{"a", "b"}},
			{Name: "coding", Sources: []string{"missing"}},
		},
	}

	var buf bytes.Buffer
	reportUnavailableSources(&buf, cfg, []string{"a", "b"})

	out := buf.String()
	if !strings.Contains(out, `"missing"`) || !strings.Contains(out, `"coding"`) {
		t.Errorf("expected a warning for the missing source; got:\n%s", out)
	}

	if strings.Contains(out, `"a"`) || strings.Contains(out, `"b"`) {
		t.Errorf("available sources should not warn; got:\n%s", out)
	}
}

type fakeLister struct {
	ids    []string
	called bool
}

func (f *fakeLister) All() []string {
	f.called = true

	return f.ids
}

func TestEnsureConfig(t *testing.T) {
	t.Run("no-op when config exists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(path, []byte("version: 2\n"), 0o600); err != nil {
			t.Fatalf("seed: %v", err)
		}

		lister := &fakeLister{ids: []string{"x"}}
		if err := ensureConfig(path, lister); err != nil {
			t.Fatalf("ensureConfig: %v", err)
		}

		if lister.called {
			t.Error("ensureConfig queried input sources despite an existing config")
		}

		data, _ := os.ReadFile(path)
		if string(data) != "version: 2\n" {
			t.Error("ensureConfig overwrote an existing config")
		}
	})

	t.Run("seeds a default when missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		lister := &fakeLister{ids: []string{"com.apple.keylayout.US"}}

		if err := ensureConfig(path, lister); err != nil {
			t.Fatalf("ensureConfig: %v", err)
		}

		if !lister.called {
			t.Error("ensureConfig did not query input sources when seeding")
		}

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("ensureConfig did not create the config: %v", err)
		}
	})
}

func TestRealClockNow(t *testing.T) {
	before := time.Now()
	got := realClock{}.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("Now() = %v, want within [%v, %v]", got, before, after)
	}
}

func TestNewRootCmd(t *testing.T) {
	root := newRootCmd("1.2.3", "deadbeef")

	if root.Use != "betterglobekey" {
		t.Errorf("Use = %q, want betterglobekey", root.Use)
	}

	if root.Version != "1.2.3" {
		t.Errorf("Version = %q, want 1.2.3", root.Version)
	}

	if root.Annotations["commit"] != "deadbeef" {
		t.Errorf("commit annotation = %q, want deadbeef", root.Annotations["commit"])
	}

	if root.PersistentFlags().Lookup("config") == nil {
		t.Error("root is missing the --config flag")
	}

	names := map[string]bool{}
	for _, sub := range root.Commands() {
		names[sub.Name()] = true
	}

	for _, want := range []string{"list", "current", "doctor", "edit"} {
		if !names[want] {
			t.Errorf("root is missing the %q subcommand", want)
		}
	}
}

func TestRunEditUsesEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	// `true` accepts the path argument and exits 0, standing in for an editor.
	t.Setenv("EDITOR", "true")

	cmd := cmdWithConfig(t, path)
	cmd.SetOut(&bytes.Buffer{})

	if err := runEdit(cmd, nil); err != nil {
		t.Fatalf("runEdit: %v", err)
	}
}

func TestListCommandOutputsSources(t *testing.T) {
	cmd := newListCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("list: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("list produced no output; the system should have at least one input source")
	}
}

func TestCurrentCommandOutputsSource(t *testing.T) {
	cmd := newCurrentCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("current: %v", err)
	}

	if !strings.Contains(buf.String(), "\t") {
		t.Errorf("current output = %q, want an id<TAB>name line", buf.String())
	}
}

func TestResolveConfigPathHomeError(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := resolveConfigPath(cmdWithConfig(t, "")); err == nil {
		t.Fatal("resolveConfigPath succeeded without a home directory")
	}
}

func TestResolveSocketPathHomeError(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := resolveSocketPath(); err == nil {
		t.Fatal("resolveSocketPath succeeded without a home directory")
	}
}

func testLogger(t *testing.T) *logging.Logger {
	t.Helper()

	logger := logging.New(domainconfig.Logger{
		Path:  filepath.Join(t.TempDir(), "test.log"),
		Level: "info",
	})
	t.Cleanup(func() { _ = logger.Close() })

	return logger
}

func TestStartControlServer(t *testing.T) {
	// A short home keeps the derived socket path within the macOS sun_path limit;
	// t.TempDir nests under a long, test-named directory that can overflow it.
	home, err := os.MkdirTemp("", "bgk")
	if err != nil {
		t.Fatalf("temp home: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(home) })
	t.Setenv("HOME", home)

	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if err := startControlServer(ctx, cfgPath, inputsource.New(), testLogger(t), "1.0.0", "abc"); err != nil {
		t.Fatalf("startControlServer: %v", err)
	}

	socket := filepath.Join(home, socketFileName)
	waitForSocket(t, socket)

	cancel()

	// The socket is removed once the context is cancelled.
	for range 200 {
		if _, err := os.Stat(socket); os.IsNotExist(err) {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Errorf("control socket %q was not removed after cancellation", socket)
}

func TestWatchConfigReloadsOnChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	logger := testLogger(t)
	daemon := app.NewDaemon(nil, nil, nil)

	rebuilt := make(chan struct{}, 8)
	build := func(domainconfig.Config) *switcher.Switcher {
		rebuilt <- struct{}{}

		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watchConfig(ctx, path, logger, daemon, build)

	deadline := time.After(5 * time.Second)

	tick := time.NewTicker(300 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-rebuilt:
			return
		case <-deadline:
			t.Fatal("watchConfig did not rebuild the switcher after a config change")
		case <-tick.C:
			_ = os.WriteFile(path, []byte(validConfig+"# touched\n"), 0o600)
		}
	}
}

func TestRunDoctorReportsInvalidConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("version: 2\ndouble_press: {maximum_delay: nonsense}\n"), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	cmd := cmdWithConfig(t, path)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := runDoctor(cmd, nil); err != nil {
		t.Fatalf("runDoctor: %v", err)
	}

	if !strings.Contains(buf.String(), "INVALID") {
		t.Errorf("doctor did not flag the invalid config; got:\n%s", buf.String())
	}
}

func TestRunDoctorReportsStatus(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home) // control socket resolves here and is absent -> not running

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(validConfig), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	cmd := cmdWithConfig(t, path)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := runDoctor(cmd, nil); err != nil {
		t.Fatalf("runDoctor: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"accessibility:", "config file:", "config:", "current source:", "not available"} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor output missing %q; got:\n%s", want, out)
		}
	}
}
