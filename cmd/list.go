// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"fmt"

	"github.com/Serpentiel/betterglobekey/pkg/inputsource"
	"github.com/spf13/cobra"
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available input sources",
	Run: func(_ *cobra.Command, _ []string) {
		for _, v := range inputsource.All() {
			// nolint:forbidigo
			fmt.Println(v)
		}
	},
}

// init is the function that is called when the package is initialized. This should be avoided if possible.
func init() {
	rootCmd.AddCommand(listCmd)
}
