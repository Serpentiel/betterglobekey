// Package cli is the command-line entry point: it parses commands and wires the
// composition root that assembles and runs the daemon.
package cli

import (
	"cmp"
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
	"strings"
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
	"github.com/Serpentiel/betterglobekey/internal/infra/keymap"
	"github.com/Serpentiel/betterglobekey/internal/infra/logging"
	"github.com/Serpentiel/betterglobekey/pkg/accessibility"
	"github.com/Serpentiel/betterglobekey/pkg/hud"
	"github.com/Serpentiel/betterglobekey/pkg/inputmonitoring"
	"github.com/Serpentiel/betterglobekey/pkg/mainthread"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// configFileName is the default config file name, looked up in the user's home directory.
const configFileName = ".betterglobekey.yaml"

// socketFileName is the control socket name, created in the user's home directory.
const socketFileName = ".betterglobekey.sock"

// keymapStateFileName records the daemon's trigger key remapping, in the user's
// home directory, so a crashed run's remapping is undone on the next start.
const keymapStateFileName = ".betterglobekey.keymap.json"

// listColumnPadding is the gap between columns in the `list` output.
const listColumnPadding = 2

// daemonStatusTimeout bounds the doctor's status query to the running daemon.
const daemonStatusTimeout = 2 * time.Second

// keymapSyncInterval is how often the remapper re-applies the trigger key, to
// pick up keyboards connected since and mappings the system dropped.
const keymapSyncInterval = 2 * time.Second

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
		logger.Error("accessibility permission not granted; the trigger key is ignored until it is " +
			"enabled in System Settings > Privacy & Security > Accessibility")
		accessibility.Prompt()
	}

	remapper, err := newRemapper()
	if err != nil {
		return err
	}

	defer restoreKeymap(remapper, logger)

	build := newSwitcherBuilder(controller, remapper, logger)

	daemon := app.NewDaemon(build(cfg), keyboard.New(), logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go watchConfig(ctx, path, logger, daemon, build)
	go syncKeymap(ctx, remapper, logger)

	root := cmd.Root()
	if err = startControlServer(
		ctx, path, controller, keyboardsPort{remapper}, logger, root.Version, root.Annotations["commit"],
	); err != nil {
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

// newSwitcherBuilder returns the function that applies a configuration: it sets
// up the trigger key, reverse modifier, and HUD, and builds the matching switcher.
// It runs at startup and on every reload.
func newSwitcherBuilder(
	controller inputsource.Controller,
	remapper *keymap.Remapper,
	logger *logging.Logger,
) func(domainconfig.Config) *switcher.Switcher {
	remapper.Observe(watchTriggers)
	keyboard.OnRawInput(logRawInput(logger))

	return func(c domainconfig.Config) *switcher.Switcher {
		keyboard.SetReverseModifier(c.Reverse.Modifier)

		if readsRawInput(c.Trigger) && !inputmonitoring.Granted() {
			logger.Error("input monitoring permission not granted; dedicated and passed-through trigger keys are " +
				"ignored until it is enabled in System Settings > Privacy & Security > Input Monitoring")
		}

		if err := remapper.Apply(c.Trigger); err != nil {
			logger.Error("remapping the trigger keys failed", "error", err)
		}

		hud.SetDuration(c.HUD.Duration)

		var notifier switcher.Notifier
		if c.HUD.Enabled {
			notifier = feedback.NewHUD(controller, c.HUD.ShowCollection)
		}

		return switcher.New(c, controller, realClock{}, logger, notifier)
	}
}

// syncKeymap keeps the trigger key remapped on keyboards as they come and go,
// until ctx is done.
func syncKeymap(ctx context.Context, remapper *keymap.Remapper, logger *logging.Logger) {
	remapper.Run(ctx, keymapSyncInterval, func(err error) {
		logger.Error("remapping the trigger key failed", "error", err)
	})
}

// restoreKeymap undoes the trigger key remapping when the daemon exits.
func restoreKeymap(remapper *keymap.Remapper, logger *logging.Logger) {
	if err := remapper.Restore(); err != nil {
		logger.Error("restoring key mappings failed", "error", err)
	}
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

// permissions reports the macOS permissions this process holds.
func permissions() control.Permissions {
	return control.Permissions{Accessibility: accessibility.Trusted(), InputMonitoring: inputmonitoring.Granted()}
}

// startControlServer launches the local gRPC control server on a background
// goroutine and arranges for it to stop when ctx is cancelled. It returns an
// error only if the socket path cannot be resolved.
func startControlServer(
	ctx context.Context,
	path string,
	sources control.Sources,
	keyboards control.Keyboards,
	logger *logging.Logger,
	version string,
	commit string,
) error {
	socketPath, err := resolveSocketPath()
	if err != nil {
		return err
	}

	server := control.NewServer(path, sources, keyboards, mainthread.Run, version, commit, permissions)

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

	status, running := queryDaemonStatus(socketPath)
	fmt.Fprintln(out, accessibilityStatusLine(running, status.GetAccessibilityTrusted()))

	if running {
		fmt.Fprintln(out, inputMonitoringStatusLine(status.GetInputMonitoringGranted()))
	}

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
		fmt.Fprintf(out, "trigger:       %s\n", triggerStatus(cfg.Trigger))
		reportKeyboards(out, cfg.Trigger)
		reportUnavailableSources(out, cfg, sources.All())
	}

	current := sources.Current()
	fmt.Fprintf(out, "current source: %s (%s)\n", current, sources.Name(current))

	return nil
}

// triggerStatus describes how many keyboards set their own trigger.
func triggerStatus(trigger domainconfig.Trigger) string {
	if listed := len(trigger.Keyboards); listed > 0 {
		return fmt.Sprintf("%d keyboard(s) set their own; others use their Globe key, if any", listed)
	}

	return "Globe key on keyboards that have one"
}

// describeTrigger renders a keyboard trigger: its key and how it is detected.
func describeTrigger(trigger domainconfig.KeyboardTrigger) string {
	switch {
	case trigger.None():
		return "none"
	case trigger.Globe():
		return "Globe key"
	case trigger.Dedicated():
		return trigger.Key + " (silenced, read directly)"
	default:
		return trigger.Key + " (passed through, read directly)"
	}
}

// reportKeyboards lists the connected keyboards and their key mappings.
func reportKeyboards(out io.Writer, trigger domainconfig.Trigger) {
	statuses, systemWide, err := listKeyboards(trigger)
	if err != nil {
		fmt.Fprintf(out, "keyboards:     unknown (%v)\n", err)

		return
	}

	for _, line := range keyboardLines(statuses, systemWide) {
		fmt.Fprintln(out, line)
	}
}

// listKeyboards reports the connected keyboards against trigger.
func listKeyboards(trigger domainconfig.Trigger) ([]keymap.Status, []keymap.Mapping, error) {
	remapper, err := newRemapper()
	if err != nil {
		return nil, nil, err
	}

	return remapper.Keyboards(trigger)
}

// keyboardLines renders the doctor's keyboard report.
func keyboardLines(statuses []keymap.Status, systemWide []keymap.Mapping) []string {
	lines := make([]string, 0, len(statuses)+1)

	for _, status := range statuses {
		lines = append(lines, keyboardLine(status))
	}

	if len(systemWide) > 0 {
		described := make([]string, 0, len(systemWide))
		for _, mapping := range systemWide {
			described = append(described, describeMapping(mapping))
		}

		lines = append(lines, "system-wide:   "+strings.Join(described, "; "))
	}

	return lines
}

// keyboardLine renders one keyboard: its identity, its trigger, and its key
// mappings.
func keyboardLine(status keymap.Status) string {
	details := []string{status.KeyboardID()}

	if status.BuiltIn {
		details = append(details, "built-in")
	} else if status.Transport != "" {
		details = append(details, status.Transport)
	}

	if status.HasGlobe() {
		details = append(details, "has Globe key")
	}

	trigger := "trigger " + describeTrigger(status.Trigger)
	if status.Listed {
		trigger += ", set for this keyboard"
	}

	mappings := []string{trigger}

	switch {
	case status.Ours != nil:
		mappings = append(mappings, keymap.UsageName(status.Ours.Src)+" silenced")
	case status.Trigger.Remapped():
		mappings = append(mappings, "not remapped yet (is the service running?)")
	}

	for _, other := range status.Others {
		mappings = append(mappings, describeMapping(other)+" (other tool)")
	}

	return fmt.Sprintf("keyboard:      %s (%s): %s",
		cmp.Or(status.Name, "Unnamed keyboard"), strings.Join(details, ", "), strings.Join(mappings, "; "))
}

// describeMapping renders a key mapping with key names.
func describeMapping(mapping keymap.Mapping) string {
	return keymap.UsageName(mapping.Src) + " → " + keymap.UsageName(mapping.Dst)
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

// queryDaemonStatus asks the running daemon, over the control socket, which
// permissions it holds. running is false when the daemon cannot be reached (for
// example, it is not started).
func queryDaemonStatus(socketPath string) (status *controlv1.GetStatusResponse, running bool) {
	conn, err := grpc.NewClient(
		"passthrough:///"+socketPath,
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			var dialer net.Dialer

			return dialer.DialContext(ctx, "unix", socketPath)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, false
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), daemonStatusTimeout)
	defer cancel()

	resp, err := controlv1.NewConfigServiceClient(conn).GetStatus(ctx, &controlv1.GetStatusRequest{})
	if err != nil {
		return nil, false
	}

	return resp, true
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

// inputMonitoringStatusLine renders the doctor's Input Monitoring line from the
// daemon's reported state.
func inputMonitoringStatusLine(granted bool) string {
	if granted {
		return "input monitoring: granted"
	}

	return "input monitoring: not granted (trigger keys other than the Globe key need it; enable in " +
		"System Settings > Privacy & Security > Input Monitoring, then restart the service)"
}

// newRemapper returns the trigger key remapper backed by the system's HID key
// mapping table, with its state file in the user's home directory.
func newRemapper() (*keymap.Remapper, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return keymap.New(hidStore{}, filepath.Join(home, keymapStateFileName)), nil
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
