// Package control hosts a local gRPC server that lets a companion application
// read and update the daemon's configuration and inspect input sources. It is
// served over a Unix domain socket in the user's home directory and is intended
// for local, single-user use only.
package control

import (
	"context"
	"fmt"
	"net"
	"os"

	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
	"github.com/Serpentiel/betterglobekey/internal/infra/config"
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

// OnMain runs fn on the main thread, blocking until it returns. Input-source
// queries must run there, as the underlying macOS APIs require a thread with a
// running run loop.
type OnMain func(fn func())

// Server hosts the ConfigService over a Unix domain socket.
type Server struct {
	controlv1.UnimplementedConfigServiceServer

	path    string
	sources Sources
	onMain  OnMain
	version string
	commit  string
	grpc    *grpc.Server
}

// NewServer builds a control server that reads and writes the configuration at
// path and reports input sources via sources. Input-source queries are run via
// onMain. version and commit identify the running build.
func NewServer(path string, sources Sources, onMain OnMain, version, commit string) *Server {
	srv := &Server{
		path:    path,
		sources: sources,
		onMain:  onMain,
		version: version,
		commit:  commit,
		grpc:    grpc.NewServer(),
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
