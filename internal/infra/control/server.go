// Package control hosts a local gRPC server that lets a companion application
// read and update the daemon's configuration and inspect input sources. It is
// served over a Unix domain socket in the user's home directory and is intended
// for local, single-user use only.
package control

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
	"github.com/Serpentiel/betterglobekey/internal/infra/config"
	"github.com/Serpentiel/betterglobekey/internal/infra/keymap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// socketFileMode restricts the control socket to the owning user.
const socketFileMode os.FileMode = 0o600

// Sources is the input-source port the control server reports to clients.
type Sources interface {
	// All returns the IDs of every selectable input source.
	All() []string
	// Name returns the localized display name for the given input source ID.
	Name(id string) string
}

// Keyboards is the keyboard port: the remapper's report on connected keyboards,
// and recording keys from their raw input.
type Keyboards interface {
	// Keyboards reports every connected keyboard against trigger, and the
	// system-wide key mappings.
	Keyboards(trigger domainconfig.Trigger) ([]keymap.Status, []keymap.Mapping, error)
	// Record returns the next key pressed on the keyboard with keyboardID, or
	// on any keyboard when it is empty. It fails with an error wrapping
	// ErrRawInput when raw input cannot be read.
	Record(ctx context.Context, keyboardID string) (RecordedKey, error)
}

// RecordedKey is a key pressed while recording.
type RecordedKey struct {
	// KeyboardID is the keyboard it was pressed on (see domainconfig.KeyboardID).
	KeyboardID string
	// Usage is its HID usage, usage page included.
	Usage uint64
}

// ErrRawInput reports that raw keyboard input cannot be read, typically for
// want of Accessibility.
var ErrRawInput = errors.New("reading raw keyboard input failed")

// ErrInputMonitoring reports that Input Monitoring is refused, so keys cannot
// be recorded: without it, macOS passes only a few keys, such as Caps Lock,
// through raw input. macOS grants it with Accessibility.
var ErrInputMonitoring = errors.New("input monitoring is not granted")

// recordTimeout caps how long RecordKey waits for a key.
const recordTimeout = 30 * time.Second

// Permissions are the macOS permissions the daemon holds.
type Permissions struct {
	// Accessibility is required by the global event tap, and brings the
	// Input Monitoring that reading keyboards' raw input needs.
	Accessibility bool
}

// OnMain runs fn on the main thread, blocking until it returns. Input-source
// queries must run there, as the underlying macOS APIs require a thread with a
// running run loop.
type OnMain func(fn func())

// Server hosts the ConfigService over a Unix domain socket.
type Server struct {
	controlv1.UnimplementedConfigServiceServer

	path        string
	sources     Sources
	keyboards   Keyboards
	onMain      OnMain
	version     string
	commit      string
	permissions func() Permissions
	grpc        *grpc.Server
}

// NewServer builds a control server that reads and writes the configuration at
// path and reports input sources via sources and keyboards via keyboards.
// Input-source queries are run via onMain. version and commit identify the
// running build. permissions reports the permissions the daemon holds,
// evaluated in the daemon's own process.
func NewServer(
	path string,
	sources Sources,
	keyboards Keyboards,
	onMain OnMain,
	version, commit string,
	permissions func() Permissions,
) *Server {
	srv := &Server{
		path:        path,
		sources:     sources,
		keyboards:   keyboards,
		onMain:      onMain,
		version:     version,
		commit:      commit,
		permissions: permissions,
		grpc:        grpc.NewServer(),
	}
	controlv1.RegisterConfigServiceServer(srv.grpc, srv)

	return srv
}

// Serve listens on the Unix socket at socketPath and serves until Stop is called.
// Any pre-existing socket file is removed first.
func (s *Server) Serve(socketPath string) error {
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing stale control socket %q: %w", socketPath, err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listening on control socket %q: %w", socketPath, err)
	}

	if err := os.Chmod(socketPath, socketFileMode); err != nil {
		return fmt.Errorf("securing control socket %q: %w", socketPath, err)
	}

	return s.grpc.Serve(listener)
}

// Stop gracefully stops the server, unblocking Serve.
func (s *Server) Stop() {
	s.grpc.GracefulStop()
}

// GetConfig returns the daemon's current configuration.
func (s *Server) GetConfig(
	_ context.Context,
	_ *controlv1.GetConfigRequest,
) (*controlv1.GetConfigResponse, error) {
	cfg, err := config.Load(s.path)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "loading config: %v", err)
	}

	return &controlv1.GetConfigResponse{Config: toProto(cfg)}, nil
}

// ApplyConfig validates and persists a new configuration. The daemon's file
// watcher reloads it once it is written.
func (s *Server) ApplyConfig(
	_ context.Context,
	req *controlv1.ApplyConfigRequest,
) (*controlv1.ApplyConfigResponse, error) {
	cfg, err := fromProto(req.GetConfig())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	if err := config.Save(s.path, cfg); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	return &controlv1.ApplyConfigResponse{}, nil
}

// ListInputSources returns every selectable input source with its localized name.
func (s *Server) ListInputSources(
	_ context.Context,
	_ *controlv1.ListInputSourcesRequest,
) (*controlv1.ListInputSourcesResponse, error) {
	var sources []*controlv1.InputSource

	s.onMain(func() {
		ids := s.sources.All()

		sources = make([]*controlv1.InputSource, 0, len(ids))
		for _, id := range ids {
			sources = append(sources, &controlv1.InputSource{Id: id, Name: s.sources.Name(id)})
		}
	})

	return &controlv1.ListInputSourcesResponse{Sources: sources}, nil
}

// GetVersion returns the running daemon's version and build commit.
func (s *Server) GetVersion(
	_ context.Context,
	_ *controlv1.GetVersionRequest,
) (*controlv1.GetVersionResponse, error) {
	return &controlv1.GetVersionResponse{Version: s.version, Commit: s.commit}, nil
}

// GetStatus reports the daemon's runtime state. accessibility_trusted is
// evaluated in the daemon's own process, so it reflects the daemon's grant
// rather than that of the process that launched the caller.
func (s *Server) GetStatus(
	_ context.Context,
	_ *controlv1.GetStatusRequest,
) (*controlv1.GetStatusResponse, error) {
	permissions := s.permissions()

	return &controlv1.GetStatusResponse{
		AccessibilityTrusted: permissions.Accessibility,
	}, nil
}

// ListKeyboards reports the connected keyboards against the configured trigger.
func (s *Server) ListKeyboards(
	_ context.Context,
	_ *controlv1.ListKeyboardsRequest,
) (*controlv1.ListKeyboardsResponse, error) {
	cfg, err := config.Load(s.path)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "loading config: %v", err)
	}

	statuses, systemWide, err := s.keyboards.Keyboards(cfg.Trigger)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "listing keyboards: %v", err)
	}

	return keyboardsToProto(statuses, systemWide), nil
}

// RecordKey returns the next key pressed, as a trigger key when it can be one.
func (s *Server) RecordKey(
	ctx context.Context,
	req *controlv1.RecordKeyRequest,
) (*controlv1.RecordKeyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, recordTimeout)
	defer cancel()

	key, err := s.keyboards.Record(ctx, req.GetKeyboardId())

	switch {
	case errors.Is(err, ErrInputMonitoring):
		return nil, status.Error(codes.FailedPrecondition,
			"betterglobekey can't read keyboards without Accessibility. Turn on betterglobekey under System "+
				"Settings > Privacy & Security > Accessibility.")
	case errors.Is(err, ErrRawInput):
		return nil, status.Errorf(codes.FailedPrecondition,
			"%v. Turn on betterglobekey under System Settings > Privacy & Security > Accessibility.", err)
	case err != nil:
		return nil, status.FromContextError(err).Err()
	}

	trigger, _ := keymap.TriggerKeyFor(key.Usage)

	name := keymap.UsageName(key.Usage)
	if trigger == domainconfig.TriggerGlobe {
		name = "Globe"
	}

	return &controlv1.RecordKeyResponse{Key: trigger, KeyName: name, KeyboardId: key.KeyboardID}, nil
}
