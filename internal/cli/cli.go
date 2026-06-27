// Package cli is the command-line entry point: it parses commands and wires the
// composition root that assembles and runs the daemon.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/app"
	"github.com/Serpentiel/betterglobekey/internal/domain/switcher"
	"github.com/Serpentiel/betterglobekey/internal/infra/config"
	"github.com/Serpentiel/betterglobekey/internal/infra/inputsource"
	"github.com/Serpentiel/betterglobekey/internal/infra/keyboard"
	"github.com/Serpentiel/betterglobekey/internal/infra/logging"
	"github.com/spf13/cobra"
)

// configFileName is the default config file name, looked up in the user's home directory.
const configFileName = ".betterglobekey.yaml"

// Execute builds the root command and runs it.
func Execute() error {
	return newRootCmd().Execute()
}

// newRootCmd builds the root command, which runs the daemon.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "betterglobekey",
		Short:         "Make the Globe key great again!",
		RunE:          runDaemon,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringP("config", "c", "", "path to the config file")
	root.AddCommand(newListCmd())

	return root
}

// newListCmd builds the command that lists available input sources.
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available input sources",
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, id := range inputsource.New().All() {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), id); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// runDaemon is the composition root: it resolves the config, assembles the
// daemon from domain and infrastructure components, and runs it until interrupted.
func runDaemon(cmd *cobra.Command, _ []string) error {
	path, err := resolveConfigPath(cmd)
	if err != nil {
		return err
	}

	sources := inputsource.New()

	if _, statErr := os.Stat(path); errors.Is(statErr, fs.ErrNotExist) {
		if err = config.WriteDefault(path, sources.All()); err != nil {
			return err
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	logger := logging.New(cfg.Logger)
	sw := switcher.New(cfg, sources, realClock{}, logger)
	daemon := app.NewDaemon(sw, keyboard.New(), logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return daemon.Run(ctx)
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
