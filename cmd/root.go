// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"os"
	"strings"

	"github.com/Serpentiel/betterglobekey/internal/assets"
	"github.com/Serpentiel/betterglobekey/internal/eventlistener"
	"github.com/Serpentiel/betterglobekey/pkg/inputsource"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// configFileMode is the file mode for the config file.
const configFileMode os.FileMode = 0660

// logger is the logger for the application.
var logger *zap.SugaredLogger

// rootCmd represents the root command.
var rootCmd = &cobra.Command{
	Use:   "betterglobekey",
	Short: "Make the Globe key great again!",
	Run:   runRootCmd,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "defines the config file to use")

	cobra.OnInitialize(setup)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// setup is the function that is called when the root command is being initialized.
func setup() {
	l, err := zap.NewProduction()

	cobra.CheckErr(err)

	logger = l.Sugar()
}

// runRootCmd is the work function for the root command.
func runRootCmd(cmd *cobra.Command, args []string) {
	{
		configFile, err := cmd.PersistentFlags().GetString("config")

		cobra.CheckErr(err)

		if configFile == "" {
			homeDir, err := os.UserHomeDir()

			cobra.CheckErr(err)

			viper.AddConfigPath(homeDir)
			viper.SetConfigName(".betterglobekey")
			viper.SetConfigType("yaml")

			configFile = strings.Join([]string{homeDir, ".betterglobekey.yaml"}, string(os.PathSeparator))
		} else {
			viper.SetConfigFile(configFile)
		}

		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			err = os.WriteFile(configFile, assets.ExampleConfig, configFileMode)
			if err != nil {
				logger.Fatalw("failed to create example config", zap.Error(err))
			}

			if err = viper.ReadInConfig(); err != nil {
				logger.Fatalw("failed to read config", zap.Error(err))
			}

			viper.Set("input_sources.primary", inputsource.All())

			if err = viper.WriteConfig(); err != nil {
				logger.Fatalw("failed to write config", zap.Error(err))
			}
		}
	}

	logger.Info("betterglobekey has been started")

	go eventlistener.Run(logger)

	select {}
}
