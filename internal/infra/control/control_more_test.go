package control

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestFromProtoNil(t *testing.T) {
	if _, err := fromProto(nil); err == nil {
		t.Fatal("fromProto accepted a nil config")
	}
}

func TestFromProtoInvalidHUDDuration(t *testing.T) {
	proto := sampleProtoConfig()
	proto.Hud.Duration = "not-a-duration"

	if _, err := fromProto(proto); err == nil {
		t.Fatal("fromProto accepted an invalid hud.duration")
	}
}

func TestGetConfigLoadError(t *testing.T) {
	// An empty path cannot be read, so GetConfig must surface a load error.
	client := newTestClient(t, "", fakeSources{}, true)

	if _, err := client.GetConfig(context.Background(), &controlv1.GetConfigRequest{}); err == nil {
		t.Fatal("GetConfig succeeded with an unreadable path")
	}
}

func TestApplyConfigRejectsUnsavable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	client := newTestClient(t, path, fakeSources{}, true)

	// Valid durations (so fromProto succeeds) but an empty logger path, which the
	// domain validation invoked by Save rejects.
	invalid := sampleProtoConfig()
	invalid.Logger.Path = ""

	if _, err := client.ApplyConfig(
		context.Background(),
		&controlv1.ApplyConfigRequest{Config: invalid},
	); err == nil {
		t.Fatal("ApplyConfig persisted a config that fails validation")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("ApplyConfig wrote a file for an invalid config")
	}
}

func TestServeOverUnixSocket(t *testing.T) {
	dir, err := os.MkdirTemp("", "bgkctl")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	socket := filepath.Join(dir, "s")

	// A stale file at the socket path must be removed by Serve before listening.
	if err = os.WriteFile(socket, []byte("stale"), 0o600); err != nil {
		t.Fatalf("seed stale socket: %v", err)
	}

	server := NewServer("", fakeSources{}, func(fn func()) { fn() }, "9.9.9", "cafe", func() bool { return true })

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(socket) }()

	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///"+socket,
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			var dialer net.Dialer

			return dialer.DialContext(ctx, "unix", socket)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	// Serve removes the stale file, then binds the real socket; retry until the
	// listener is accepting connections.
	client := controlv1.NewConfigServiceClient(conn)

	var resp *controlv1.GetVersionResponse

	for range 200 {
		resp, err = client.GetVersion(context.Background(), &controlv1.GetVersionRequest{})
		if err == nil {
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("GetVersion over unix socket: %v", err)
	}

	if resp.GetVersion() != "9.9.9" {
		t.Errorf("version = %q, want 9.9.9", resp.GetVersion())
	}

	server.Stop()

	if err := <-serveErr; err != nil {
		t.Errorf("Serve returned %v, want nil after Stop", err)
	}
}
