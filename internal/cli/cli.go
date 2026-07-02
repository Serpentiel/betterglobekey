// Package cli is the command-line entry point: it parses commands and wires the
// composition root that assembles and runs the daemon.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/app"
	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	"github.com/Serpentiel/betterglobekey/internal/domain/switcher"
	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
	"github.com/Serpentiel/betterglobekey/internal/infra/config"
	"github.com/Serpentiel/betterglobekey/internal/infra/control"
	"github.com/Serpentiel/betterglobekey/internal/infra/feedback"
	"github.com/Serpentiel/betterglobekey/internal/infra/inputsource"
	"github.com/Serpentiel/betterglobekey/internal/infra/keyboard"
	"github.com/Serpentiel/betterglobekey/internal/infra/logging"
	"github.com/Serpentiel/betterglobekey/pkg/accessibility"
	"github.com/Serpentiel/betterglobekey/pkg/hud"
	"github.com/Serpentiel/betterglobekey/pkg/mainthread"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// configFileName is the default config file name, looked up in the user's home directory.
const configFileName = ".betterglobekey.yaml"

// socketFileName is the control socket name, created in the user's home directory.
const socketFileName = ".betterglobekey.sock"

// listColumnPadding is the gap between columns in the `list` output.
const listColumnPadding = 2

// daemonStatusTimeout bounds the doctor's status query to the running daemon.
const daemonStatusTimeout = 2 * time.Second

// Execute builds the root command and runs it.
func Execute(version, commit string) error {
	return newRootCmd(version, commit).Execute()
}

// newRootCmd builds the root command, which runs the daemon. The build commit is
// stored in the command's annotations so the daemon can report it.
func newRootCmd(version, commit string) *cobra.Command {
	root := &cobra.Command{
		Use:           "betterglobekey",
		Short:         "Make the Globe key great again!",
		Version:       version,
		Annotations:   map[string]string{"commit": commit},
		RunE:          runDaemon,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringP("config", "c", "", "path to the config file")
	root.AddCommand(newListCmd(), newCurrentCmd(), newDoctorCmd(), newEditCmd())

	return root
}

// newListCmd lists available input sources with their localized names.
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available input sources",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sources := inputsource.New()

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, listColumnPadding, ' ', 0)

			for _, id := range sources.All() {
				if _, err := fmt.Fprintf(writer, "%s\t%s\n", id, sources.Name(id)); err != nil {
					return err
				}
			}

			return writer.Flush()
		},
	}
}

// newCurrentCmd prints the currently active input source.
func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Print the currently active input source",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sources := inputsource.New()

			id := sources.Current()

			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", id, sources.Name(id))

			return err
		},
	}
}

// newDoctorCmd diagnoses configuration and permissions.
func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose configuration and permissions",
		RunE:  runDoctor,
	}
}

// newEditCmd opens the configuration file in the user's editor.
func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open the configuration file in your editor",
		RunE:  runEdit,
	}
}

// runDaemon is the composition root: it resolves the config, assembles the daemon
// from domain and infrastructure components, watches the config for changes, and
// runs until interrupted.
func runDaemon(cmd *cobra.Command, _ []string) error {
	path, err := resolveConfigPath(cmd)
	if err != nil {
		return err
	}

	controller := inputsource.New()

	if err = ensureConfig(path, controller); err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	logger := logging.New(cfg.Logger)

	if !accessibility.Trusted() {
		logger.Error("accessibility permission not granted; the Globe key is ignored until it is " +
			"enabled in System Settings > Privacy & Security > Accessibility")
		accessibility.Prompt()
	}

	build := func(c domainconfig.Config) *switcher.Switcher {
		keyboard.SetReverseModifier(c.Reverse.Modifier)
		hud.SetDuration(c.HUD.Duration)

		var notifier switcher.Notifier
		if c.HUD.Enabled {
			notifier = feedback.NewHUD(controller, c.HUD.ShowCollection)
		}

		return switcher.New(c, controller, realClock{}, logger, notifier)
	}

	daemon := app.NewDaemon(build(cfg), keyboard.New(), logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go watchConfig(ctx, path, logger, daemon, build)

	root := cmd.Root()
	if err = startControlServer(ctx, path, controller, logger, root.Version, root.Annotations["commit"]); err != nil {
		return err
	}

	// The HUD owns the main thread's AppKit run loop; the daemon runs the event
	// tap on its own goroutine. Stop the run loop when a signal arrives.
	daemon.Start()
	defer daemon.Stop()

	go func() {
		<-ctx.Done()
		hud.Stop()
	}()

	hud.Run()

	return nil
}

// watchConfig reloads the daemon's switcher whenever the config file changes.
func watchConfig(
	ctx context.Context,
	path string,
	logger *logging.Logger,
	daemon *app.Daemon,
	build func(domainconfig.Config) *switcher.Switcher,
) {
	err := config.Watch(ctx, path, func() {
		reloaded, err := config.Load(path)
		if err != nil {
			logger.Error("config reload failed", "error", err)

			return
		}

		daemon.Reload(build(reloaded))
		logger.Info("configuration reloaded")
	})
	if err != nil {
		logger.Error("config watcher stopped", "error", err)
	}
}

// startControlServer launches the local gRPC control server on a background
// goroutine and arranges for it to stop when ctx is cancelled. It returns an
// error only if the socket path cannot be resolved.
func startControlServer(
	ctx context.Context,
	path string,
	sources control.Sources,
	logger *logging.Logger,
	version string,
	commit string,
) error {
	socketPath, err := resolveSocketPath()
	if err != nil {
		return err
	}

	server := control.NewServer(path, sources, mainthread.Run, version, commit, accessibility.Trusted)

	go func() {
		if err := server.Serve(socketPath); err != nil {
			logger.Error("control server stopped", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		server.Stop()

		if err := os.Remove(socketPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			logger.Error("removing control socket failed", "error", err)
		}
	}()

	logger.Info("control server listening", "socket", socketPath)

	return nil
}

// runDoctor reports on configuration validity, permissions, and the current source.
func runDoctor(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	sources := inputsource.New()

	// Ask the running daemon rather than checking this process: AXIsProcessTrusted
	// reflects the responsible process (e.g. the terminal that launched us), not
	// the daemon that actually needs the permission.
	socketPath, err := resolveSocketPath()
	if err != nil {
		return err
	}

	trusted, running := queryDaemonAccessibility(socketPath)
	fmt.Fprintln(out, accessibilityStatusLine(running, trusted))

	path, err := resolveConfigPath(cmd)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "config file:   %s\n", path)

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(out, "config:        INVALID (%v)\n", err)
	} else {
		fmt.Fprintf(out, "config:        valid (%d collection(s))\n", len(cfg.Collections))
		reportUnavailableSources(out, cfg, sources.All())
	}

	current := sources.Current()
	fmt.Fprintf(out, "current source: %s (%s)\n", current, sources.Name(current))

	return nil
}

// reportUnavailableSources warns about configured sources that the system does not offer.
func reportUnavailableSources(out io.Writer, cfg domainconfig.Config, available []string) {
	for _, collection := range cfg.Collections {
		for _, id := range collection.Sources {
			if !slices.Contains(available, id) {
				fmt.Fprintf(out, "warning:       source %q in collection %q is not available\n", id, collection.Name)
			}
		}
	}
}

// runEdit opens the configuration file in $EDITOR, falling back to `open`.
func runEdit(cmd *cobra.Command, _ []string) error {
	path, err := resolveConfigPath(cmd)
	if err != nil {
		return err
	}

	if err = ensureConfig(path, inputsource.New()); err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")

	var command *exec.Cmd
	if editor == "" {
		command = exec.Command("open", path) //nolint:gosec // path is the user's config file
	} else {
		command = exec.Command(editor, path) //nolint:gosec // editor and path are user-controlled
	}

	command.Stdin = os.Stdin
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = os.Stderr

	return command.Run()
}

// sourceLister lists the system's available input source IDs. It is the subset
// of the input-source controller ensureConfig needs to seed a default config,
// narrowed to an interface so it can be faked in tests.
type sourceLister interface {
	All() []string
}

// ensureConfig writes a default, seeded config if none exists at path.
func ensureConfig(path string, sources sourceLister) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return config.WriteDefault(path, sources.All())
	}

	return nil
}

// resolveConfigPath returns the config path from the --config flag, defaulting to
// the file in the user's home directory.
func resolveConfigPath(cmd *cobra.Command) (string, error) {
	path, err := cmd.Flags().GetString("config")
	if err != nil {
		return "", err
	}

	if path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, configFileName), nil
}

// queryDaemonAccessibility asks the running daemon, over the control socket,
// whether it holds the Accessibility permission. running is false when the
// daemon cannot be reached (for example, it is not started).
func queryDaemonAccessibility(socketPath string) (trusted, running bool) {
	conn, err := grpc.NewClient(
		"passthrough:///"+socketPath,
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			var dialer net.Dialer

			return dialer.DialContext(ctx, "unix", socketPath)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false, false
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), daemonStatusTimeout)
	defer cancel()

	resp, err := controlv1.NewConfigServiceClient(conn).GetStatus(ctx, &controlv1.GetStatusRequest{})
	if err != nil {
		return false, false
	}

	return resp.GetAccessibilityTrusted(), true
}

// accessibilityStatusLine renders the doctor's accessibility line from the
// daemon's reported state. The permission belongs to the daemon, so it is only
// meaningful while the service is running.
func accessibilityStatusLine(running, trusted bool) string {
	switch {
	case !running:
		return "accessibility: unknown (service not running; " +
			"start it with `brew services start betterglobekey`, then re-run)"
	case trusted:
		return "accessibility: granted"
	default:
		return "accessibility: NOT granted (enable in System Settings > Privacy & Security > Accessibility)"
	}
}

// resolveSocketPath returns the path to the control socket in the user's home
// directory.
func resolveSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, socketFileName), nil
}

// realClock is the production Clock backed by the wall clock.
type realClock struct{}

// Now returns the current time.
func (realClock) Now() time.Time {
	return time.Now()
}
