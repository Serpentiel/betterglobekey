// Package provide is the package that contains all of the providers for the dependency injection container.
package provide

import (
	"os"
	"strings"

	"github.com/Serpentiel/betterglobekey/internal/assets"
	"github.com/Serpentiel/betterglobekey/pkg/inputsource"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Viper is a function which provides a viper.Viper instance.
func Viper(cmd *cobra.Command) (*viper.Viper, error) {
	const (
		// configFileName is the name of the config file.
		configFileName string = ".betterglobekey"

		// configFileMode is the mode that the config file will be created with.
		configFileMode os.FileMode = 0660
	)

	v := viper.New()

	configFile, err := cmd.PersistentFlags().GetString("config")
	if err != nil {
		return nil, err
	}

	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		v.AddConfigPath(homeDir)

		v.SetConfigName(configFileName)

		v.SetConfigType("yaml")

		configFile = strings.Join([]string{homeDir, configFileName + ".yaml"}, string(os.PathSeparator))
	} else {
		v.SetConfigFile(configFile)
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		err = os.WriteFile(configFile, assets.ExampleConfig, configFileMode)
		if err != nil {
			return nil, err
		}

		if err = v.ReadInConfig(); err != nil {
			return nil, err
		}

		viper.Set("input_sources.primary", inputsource.All())

		if err = viper.WriteConfig(); err != nil {
			return nil, err
		}
	}

	return v, nil
}
