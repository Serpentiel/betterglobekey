package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/infra/control"
)

func TestAccessibilityStatusLine(t *testing.T) {
	cases := []struct {
		name    string
		running bool
		trusted bool
		want    string
	}{
		{"not running", false, false, "service not running"},
		{"granted", true, true, ": granted"},
		{"not granted", true, false, "NOT granted"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := accessibilityStatusLine(tc.running, tc.trusted)
			if !strings.Contains(got, tc.want) {
				t.Errorf("line = %q, want it to contain %q", got, tc.want)
			}
		})
	}
}

// shortSocketPath returns a socket path short enough for the macOS sun_path
// limit (t.TempDir paths can be too long).
func shortSocketPath(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "bgk")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	return filepath.Join(dir, "s")
}

func TestQueryDaemonAccessibilityRunning(t *testing.T) {
	for _, trusted := range []bool{true, false} {
		name := "granted"
		if !trusted {
			name = "denied"
		}

		t.Run(name, func(t *testing.T) {
			socketPath := shortSocketPath(t)
			server := control.NewServer("", nil, func(fn func()) { fn() }, "", "", func() bool { return trusted })

			go func() { _ = server.Serve(socketPath) }()

			t.Cleanup(server.Stop)

			waitForSocket(t, socketPath)

			gotTrusted, running := queryDaemonAccessibility(socketPath)
			if !running {
				t.Fatal("running = false, want true")
			}

			if gotTrusted != trusted {
				t.Errorf("trusted = %v, want %v", gotTrusted, trusted)
			}
		})
	}
}

func TestQueryDaemonAccessibilityNotRunning(t *testing.T) {
	trusted, running := queryDaemonAccessibility(shortSocketPath(t))
	if running || trusted {
		t.Errorf("got running=%v trusted=%v, want false false", running, trusted)
	}
}

func waitForSocket(t *testing.T, path string) {
	t.Helper()

	for range 200 {
		if _, err := os.Stat(path); err == nil {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("socket %q did not appear", path)
}
