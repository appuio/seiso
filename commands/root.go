package commands

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "image-cleanup",
		Short: "An image cleanup CLI",
		Long:  `tbd`,
	}
)

func init() {
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
