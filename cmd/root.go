// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"context"
	"os"

	"github.com/Serpentiel/betterglobekey/internal/pkg/eventhandler"
	"github.com/Serpentiel/betterglobekey/internal/provide"
	"github.com/Serpentiel/betterglobekey/pkg/logger"
	hook "github.com/robotn/gohook"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// rootCmd is the rootCmd command for the application.
var rootCmd = &cobra.Command{
	Use:   "betterglobekey",
	Short: "Make the Globe key great again!",
	Run: func(cmd *cobra.Command, _ []string) {
		fx.New(
			fx.Supply(cmd),

			fx.Provide(
				provide.Viper,
				provide.Logger,

				eventhandler.NewEventHandler,
			),

			fx.WithLogger(provide.FxeventLogger),

			fx.Invoke(func(lc fx.Lifecycle, v *viper.Viper, l logger.Logger, e *eventhandler.EventHandler) {
				lc.Append(fx.Hook{
					OnStart: func(_ context.Context) error {
						go func() {
							handlers := e.Handlers()

							for e := range hook.Start() {
								handler, ok := handlers[e.Rawcode]
								if !ok {
									continue
								}

								if e.Kind == hook.KeyUp {
									handler.KeyUp()()
								}
							}
						}()

						l.Info("event handler has been initialized")

						return nil
					},
					OnStop: func(_ context.Context) error {
						hook.End()

						return nil
					},
				})
			}),
		).Run()
	},
}

// Execute is the function that is called to execute the rootCmd command.
func Execute() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "defines the config file to use")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
