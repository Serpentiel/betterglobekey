package control

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type fakeSources struct {
	all     []string
	current string
	names   map[string]string
}

func (f fakeSources) All() []string { return f.all }

func (f fakeSources) Current() string { return f.current }

func (f fakeSources) Name(id string) string {
	if name, ok := f.names[id]; ok {
		return name
	}

	return id
}

// newTestClient starts a control server backed by path and sources on an
// in-memory listener and returns a connected client.
func newTestClient(t *testing.T, path string, sources Sources) controlv1.ConfigServiceClient {
	t.Helper()

	listener := bufconn.Listen(1 << 20)
	server := NewServer(path, sources, func(fn func()) { fn() })

	go func() { _ = server.grpc.Serve(listener) }()

	t.Cleanup(server.grpc.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	return controlv1.NewConfigServiceClient(conn)
}

func sampleProtoConfig() *controlv1.Config {
	return &controlv1.Config{
		Logger:      &controlv1.Logger{Path: "x.log", Level: "info", RetentionDays: 7, RetentionFiles: 2},
		DoublePress: &controlv1.DoublePress{Enabled: true, MaximumDelay: "300ms"},
		Reverse:     &controlv1.Reverse{Enabled: true, Modifier: "shift"},
		Hud:         &controlv1.Hud{Enabled: true, Duration: "900ms", ShowCollection: true},
		Collections: []*controlv1.Collection{
			{Name: "primary", Sources: []string{"a", "b"}},
			{Name: "coding", Sources: []string{"c"}},
		},
	}
}

func TestApplyThenGetConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	client := newTestClient(t, path, fakeSources{})

	if _, err := client.ApplyConfig(
		context.Background(),
		&controlv1.ApplyConfigRequest{Config: sampleProtoConfig()},
	); err != nil {
		t.Fatalf("ApplyConfig: %v", err)
	}

	got, err := client.GetConfig(context.Background(), &controlv1.GetConfigRequest{})
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	cfg := got.GetConfig()
	if cfg.GetDoublePress().GetMaximumDelay() != "300ms" {
		t.Errorf("delay = %q, want 300ms", cfg.GetDoublePress().GetMaximumDelay())
	}

	if !cfg.GetHud().GetEnabled() {
		t.Error("hud.enabled = false, want true")
	}

	if cfg.GetReverse().GetModifier() != "shift" {
		t.Errorf("reverse.modifier = %q, want shift", cfg.GetReverse().GetModifier())
	}

	if len(cfg.GetCollections()) != 2 {
		t.Fatalf("collections = %d, want 2", len(cfg.GetCollections()))
	}

	if cfg.GetCollections()[0].GetName() != "primary" {
		t.Errorf("first collection = %q, want primary", cfg.GetCollections()[0].GetName())
	}
}

func TestApplyConfigRejectsInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	client := newTestClient(t, path, fakeSources{})

	invalid := sampleProtoConfig()
	invalid.DoublePress.MaximumDelay = "not-a-duration"

	if _, err := client.ApplyConfig(
		context.Background(),
		&controlv1.ApplyConfigRequest{Config: invalid},
	); err == nil {
		t.Fatal("ApplyConfig accepted an invalid duration")
	}
}

func TestListInputSources(t *testing.T) {
	sources := fakeSources{
		all:   []string{"com.apple.keylayout.US", "com.apple.keylayout.Russian"},
		names: map[string]string{"com.apple.keylayout.US": "U.S."},
	}
	client := newTestClient(t, "", sources)

	resp, err := client.ListInputSources(context.Background(), &controlv1.ListInputSourcesRequest{})
	if err != nil {
		t.Fatalf("ListInputSources: %v", err)
	}

	if len(resp.GetSources()) != 2 {
		t.Fatalf("sources = %d, want 2", len(resp.GetSources()))
	}

	if resp.GetSources()[0].GetName() != "U.S." {
		t.Errorf("first name = %q, want U.S.", resp.GetSources()[0].GetName())
	}
}

func TestGetCurrentSource(t *testing.T) {
	sources := fakeSources{
		current: "com.apple.keylayout.US",
		names:   map[string]string{"com.apple.keylayout.US": "U.S."},
	}
	client := newTestClient(t, "", sources)

	resp, err := client.GetCurrentSource(context.Background(), &controlv1.GetCurrentSourceRequest{})
	if err != nil {
		t.Fatalf("GetCurrentSource: %v", err)
	}

	if resp.GetSource().GetId() != "com.apple.keylayout.US" {
		t.Errorf("id = %q, want com.apple.keylayout.US", resp.GetSource().GetId())
	}

	if resp.GetSource().GetName() != "U.S." {
		t.Errorf("name = %q, want U.S.", resp.GetSource().GetName())
	}
}
