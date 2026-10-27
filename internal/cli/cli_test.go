package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/infra/control"
)

func TestInputMonitoringStatusLine(t *testing.T) {
	if got := inputMonitoringStatusLine(true); got != "input monitoring: granted" {
		t.Errorf("granted line = %q", got)
	}

	if got := inputMonitoringStatusLine(false); !strings.Contains(got, "not granted") {
		t.Errorf("denied line = %q", got)
	}
}

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
			server := control.NewServer("", nil, nil, func(fn func()) { fn() }, "", "", func() control.Permissions {
				return control.Permissions{Accessibility: trusted, InputMonitoring: !trusted}
			})

			go func() { _ = server.Serve(socketPath) }()

			t.Cleanup(server.Stop)

			waitForSocket(t, socketPath)

			status, running := queryDaemonStatus(socketPath)
			if !running {
				t.Fatal("running = false, want true")
			}

			if status.GetAccessibilityTrusted() != trusted || status.GetInputMonitoringGranted() == trusted {
				t.Errorf("status = %+v, want accessibility %v and input monitoring %v", status, trusted, !trusted)
			}
		})
	}
}

func TestQueryDaemonAccessibilityNotRunning(t *testing.T) {
	status, running := queryDaemonStatus(shortSocketPath(t))
	if running || status != nil {
		t.Errorf("got running=%v status=%+v, want neither", running, status)
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
