// Package cli is the command-line entry point: it parses commands and wires the
// composition root that assembles and runs the daemon.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	"github.com/Serpentiel/betterglobekey/internal/infra/accessibility"
	"github.com/Serpentiel/betterglobekey/internal/infra/config"
	"github.com/Serpentiel/betterglobekey/internal/infra/feedback"
	"github.com/Serpentiel/betterglobekey/internal/infra/hud"
	"github.com/Serpentiel/betterglobekey/internal/infra/inputsource"
	"github.com/Serpentiel/betterglobekey/internal/infra/keyboard"
	"github.com/Serpentiel/betterglobekey/internal/infra/logging"
	"github.com/spf13/cobra"
)

// configFileName is the default config file name, looked up in the user's home directory.
const configFileName = ".betterglobekey.yaml"

// listColumnPadding is the gap between columns in the `list` output.
const listColumnPadding = 2

// Execute builds the root command and runs it.
func Execute(version string) error {
	return newRootCmd(version).Execute()
}

// newRootCmd builds the root command, which runs the daemon.
func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "betterglobekey",
		Short:         "Make the Globe key great again!",
		Version:       version,
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
		var notifier switcher.Notifier
		if c.HUD {
			notifier = feedback.NewHUD(controller)
		}

		return switcher.New(c, controller, realClock{}, logger, notifier)
	}

	daemon := app.NewDaemon(build(cfg), keyboard.New(), logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go watchConfig(ctx, path, logger, daemon, build)

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

// runDoctor reports on configuration validity, permissions, and the current source.
func runDoctor(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	sources := inputsource.New()

	if accessibility.Trusted() {
		fmt.Fprintln(out, "accessibility: granted")
	} else {
		fmt.Fprintln(out, "accessibility: NOT granted (enable in System Settings > Privacy & Security > Accessibility)")
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

// ensureConfig writes a default, seeded config if none exists at path.
func ensureConfig(path string, sources inputsource.Controller) error {
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

// realClock is the production Clock backed by the wall clock.
type realClock struct{}

// Now returns the current time.
func (realClock) Now() time.Time {
	return time.Now()
}
