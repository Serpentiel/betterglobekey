package control

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
	"github.com/Serpentiel/betterglobekey/internal/infra/keymap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// fakeKeyboards reports a built-in keyboard and a remapped third-party one,
// and records the keys in recorded.
type fakeKeyboards struct {
	err       error
	recorded  RecordedKey
	recordErr error
}

func (f fakeKeyboards) Record(ctx context.Context, keyboardID string) (RecordedKey, error) {
	if f.recordErr != nil {
		return RecordedKey{}, f.recordErr
	}

	if f.recorded.KeyboardID == "" || (keyboardID != "" && keyboardID != f.recorded.KeyboardID) {
		<-ctx.Done()

		return RecordedKey{}, ctx.Err()
	}

	return f.recorded, nil
}

func (f fakeKeyboards) Keyboards(trigger domainconfig.Trigger) ([]keymap.Status, []keymap.Mapping, error) {
	if f.err != nil {
		return nil, nil, f.err
	}

	capsSilenced := keymap.Mapping{Src: keymap.Usages[domainconfig.TriggerCapsLock], Dst: keymap.DedicatedUsage}
	escape := keymap.Mapping{Src: 0x7_0000_0035, Dst: 0x7_0000_0029}

	return []keymap.Status{
			{Keyboard: keymap.Keyboard{Vendor: 0x05ac, Product: 0x0281, Name: "Apple Keyboard", BuiltIn: true}},
			{
				Keyboard: keymap.Keyboard{Vendor: 0x046d, Product: 0xc33f, Name: "Logitech", Transport: "USB"},
				Trigger:  trigger.Keyboards["046d:c33f"],
				Listed:   true,
				Ours:     &capsSilenced,
				Others:   []keymap.Mapping{escape},
			},
		},
		[]keymap.Mapping{escape},
		nil
}

type fakeSources struct {
	all   []string
	names map[string]string
}

func (f fakeSources) All() []string { return f.all }

func (f fakeSources) Name(id string) string {
	if name, ok := f.names[id]; ok {
		return name
	}

	return id
}

// newTestClient starts a control server backed by path and sources on an
// in-memory listener and returns a connected client.
func newTestClient(t *testing.T, path string, sources Sources, trusted bool) controlv1.ConfigServiceClient {
	t.Helper()

	listener := bufconn.Listen(1 << 20)
	server := NewServer(
		path, sources, fakeKeyboards{}, func(fn func()) { fn() }, "1.2.3", "abc1234",
		func() Permissions { return Permissions{Accessibility: trusted} },
	)

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
		Trigger: &controlv1.Trigger{
			Keyboards: map[string]*controlv1.KeyboardTrigger{
				"046d:c33f": {Key: "insert", Mode: "pass_through"},
			},
		},
		Collections: []*controlv1.Collection{
			{Name: "primary", Sources: []string{"a", "b"}},
			{Name: "coding", Sources: []string{"c"}},
		},
	}
}

func TestApplyThenGetConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	client := newTestClient(t, path, fakeSources{}, true)

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

	trigger := cfg.GetTrigger()

	logitech := trigger.GetKeyboards()["046d:c33f"]
	if logitech.GetKey() != "insert" || logitech.GetMode() != "pass_through" {
		t.Errorf("trigger = %+v, want Insert passed through on one keyboard", trigger)
	}

	if len(cfg.GetCollections()) != 2 {
		t.Fatalf("collections = %d, want 2", len(cfg.GetCollections()))
	}

	if cfg.GetCollections()[0].GetName() != "primary" {
		t.Errorf("first collection = %q, want primary", cfg.GetCollections()[0].GetName())
	}
}

func TestApplyConfigWithoutTriggerKeepsGlobe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	client := newTestClient(t, path, fakeSources{}, true)

	// A companion that predates the trigger sends no trigger at all.
	legacy := sampleProtoConfig()
	legacy.Trigger = nil

	if _, err := client.ApplyConfig(context.Background(), &controlv1.ApplyConfigRequest{Config: legacy}); err != nil {
		t.Fatalf("ApplyConfig: %v", err)
	}

	got, err := client.GetConfig(context.Background(), &controlv1.GetConfigRequest{})
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	trigger := got.GetConfig().GetTrigger()
	if len(trigger.GetKeyboards()) != 0 {
		t.Errorf("trigger = %+v, want globe", trigger)
	}
}

func TestApplyConfigRejectsInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	client := newTestClient(t, path, fakeSources{}, true)

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
	client := newTestClient(t, "", sources, true)

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

func TestGetVersion(t *testing.T) {
	client := newTestClient(t, "", fakeSources{}, true)

	resp, err := client.GetVersion(context.Background(), &controlv1.GetVersionRequest{})
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}

	if resp.GetVersion() != "1.2.3" || resp.GetCommit() != "abc1234" {
		t.Errorf("version/commit = %q/%q, want 1.2.3/abc1234", resp.GetVersion(), resp.GetCommit())
	}
}

func TestGetStatus(t *testing.T) {
	for _, trusted := range []bool{true, false} {
		t.Run(fmt.Sprintf("trusted=%v", trusted), func(t *testing.T) {
			client := newTestClient(t, "", fakeSources{}, trusted)

			resp, err := client.GetStatus(context.Background(), &controlv1.GetStatusRequest{})
			if err != nil {
				t.Fatalf("GetStatus: %v", err)
			}

			if resp.GetAccessibilityTrusted() != trusted {
				t.Errorf("accessibility_trusted = %v, want %v", resp.GetAccessibilityTrusted(), trusted)
			}
		})
	}
}

func TestListKeyboards(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	client := newTestClient(t, path, fakeSources{}, true)

	config := sampleProtoConfig()
	config.Trigger = &controlv1.Trigger{
		Keyboards: map[string]*controlv1.KeyboardTrigger{"046d:c33f": {Key: "caps_lock"}},
	}

	if _, err := client.ApplyConfig(context.Background(), &controlv1.ApplyConfigRequest{Config: config}); err != nil {
		t.Fatalf("ApplyConfig: %v", err)
	}

	got, err := client.ListKeyboards(context.Background(), &controlv1.ListKeyboardsRequest{})
	if err != nil {
		t.Fatalf("ListKeyboards: %v", err)
	}

	if len(got.GetKeyboards()) != 2 || len(got.GetSystemMappings()) != 1 {
		t.Fatalf("response = %+v", got)
	}

	builtIn, logitech := got.GetKeyboards()[0], got.GetKeyboards()[1]
	if builtIn.GetId() != "05ac:0281" || !builtIn.GetHasGlobe() || !builtIn.GetBuiltIn() ||
		builtIn.GetTriggerMapping() != nil {
		t.Errorf("built-in = %+v", builtIn)
	}

	if logitech.GetId() != "046d:c33f" || logitech.GetHasGlobe() || logitech.GetTrigger().GetKey() != "caps_lock" ||
		logitech.GetTransport() != "USB" {
		t.Errorf("logitech = %+v", logitech)
	}

	if m := logitech.GetTriggerMapping(); m.GetSrcName() != "Caps Lock" || m.GetDstName() != "F24" {
		t.Errorf("trigger mapping = %+v", m)
	}

	others := logitech.GetOtherMappings()
	if len(others) != 1 || others[0].GetSrcName() != "Grave Accent" || others[0].GetDstName() != "Escape" {
		t.Errorf("other mappings = %+v", others)
	}
}

func TestListKeyboardsErrors(t *testing.T) {
	t.Run("config", func(t *testing.T) {
		client := newTestClient(t, filepath.Join(t.TempDir(), "missing.yaml"), fakeSources{}, true)
		if _, err := client.ListKeyboards(context.Background(), &controlv1.ListKeyboardsRequest{}); err == nil {
			t.Error("ListKeyboards succeeded without a config")
		}
	})

	t.Run("keyboards", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")

		server := NewServer(
			path, fakeSources{}, fakeKeyboards{err: fmt.Errorf("boom")}, func(fn func()) { fn() }, "", "",
			func() Permissions { return Permissions{Accessibility: true} },
		)

		request := &controlv1.ApplyConfigRequest{Config: sampleProtoConfig()}
		if _, err := server.ApplyConfig(context.Background(), request); err != nil {
			t.Fatalf("ApplyConfig: %v", err)
		}

		if _, err := server.ListKeyboards(context.Background(), &controlv1.ListKeyboardsRequest{}); err == nil {
			t.Error("ListKeyboards hid the keyboard error")
		}
	})
}

func TestRecordKey(t *testing.T) {
	insert := RecordedKey{KeyboardID: "046d:c33f", Usage: 0x7_0000_0049}
	keyA := RecordedKey{KeyboardID: "046d:c33f", Usage: 0x7_0000_0004}
	globe := RecordedKey{KeyboardID: "05ac:0281", Usage: keymap.GlobeUsage}

	tests := []struct {
		name     string
		recorded RecordedKey
		want     *controlv1.RecordKeyResponse
	}{
		{
			"trigger key", insert,
			&controlv1.RecordKeyResponse{Key: "insert", KeyName: "Insert", KeyboardId: "046d:c33f"},
		},
		{"other key", keyA, &controlv1.RecordKeyResponse{KeyName: "A", KeyboardId: "046d:c33f"}},
		{"globe key", globe, &controlv1.RecordKeyResponse{Key: "globe", KeyName: "Globe", KeyboardId: "05ac:0281"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer("", fakeSources{}, fakeKeyboards{recorded: test.recorded}, nil, "", "",
				func() Permissions { return Permissions{} })

			got, err := server.RecordKey(t.Context(), &controlv1.RecordKeyRequest{KeyboardId: test.recorded.KeyboardID})
			if err != nil {
				t.Fatalf("RecordKey: %v", err)
			}

			if got.GetKey() != test.want.GetKey() || got.GetKeyName() != test.want.GetKeyName() ||
				got.GetKeyboardId() != test.want.GetKeyboardId() {
				t.Errorf("RecordKey = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestRecordKeyErrors(t *testing.T) {
	tests := []struct {
		name      string
		keyboards fakeKeyboards
		want      codes.Code
	}{
		{"raw input", fakeKeyboards{recordErr: fmt.Errorf("%w (0xe00002e2)", ErrRawInput)}, codes.FailedPrecondition},
		{"no key in time", fakeKeyboards{}, codes.DeadlineExceeded},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer("", fakeSources{}, test.keyboards, nil, "", "",
				func() Permissions { return Permissions{} })

			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
			defer cancel()

			_, err := server.RecordKey(ctx, &controlv1.RecordKeyRequest{})
			if status.Code(err) != test.want {
				t.Errorf("err = %v, want code %v", err, test.want)
			}
		})
	}
}
