package commands

import (
	log "github.com/sirupsen/logrus"
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
func Execute() {
	if err:= rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Could not execute command.")
	}
}
